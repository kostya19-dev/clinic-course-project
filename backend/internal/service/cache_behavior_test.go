package service

import (
    "context"
    "reflect"
    "strings"
    "testing"
    "time"

    cachepkg "clinic.local/backend/internal/cache"
    "clinic.local/backend/internal/model"
    "github.com/google/uuid"
)

type fakeCache struct {
    values          map[string]string
    deletedPatterns []string
    setCalls        []string
}

func newFakeCache() *fakeCache {
    return &fakeCache{values: make(map[string]string)}
}

func (c *fakeCache) Get(_ context.Context, key string) (string, error) {
    value, ok := c.values[key]
    if !ok {
        return "", cachepkg.ErrCacheMiss
    }

    return value, nil
}

func (c *fakeCache) Set(_ context.Context, key string, value string, _ time.Duration) error {
    c.values[key] = value
    c.setCalls = append(c.setCalls, key)
    return nil
}

func (c *fakeCache) DeleteByPattern(_ context.Context, pattern string) error {
    c.deletedPatterns = append(c.deletedPatterns, pattern)
    prefix := strings.TrimSuffix(pattern, "*")

    for key := range c.values {
        if strings.HasPrefix(key, prefix) {
            delete(c.values, key)
        }
    }

    return nil
}

type fakeCatalogRepository struct {
    items []model.PriceListItem
    calls int
}

func (r *fakeCatalogRepository) ListCurrentPriceList(_ context.Context) ([]model.PriceListItem, error) {
    r.calls++
    return append([]model.PriceListItem(nil), r.items...), nil
}

type fakeScheduleRepository struct {
    availability      []model.ScheduleSlot
    appointments      []model.AppointmentSummary
    availabilityCalls int
    appointmentCalls  int
}

func (r *fakeScheduleRepository) ListAvailability(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]model.ScheduleSlot, error) {
    r.availabilityCalls++
    return append([]model.ScheduleSlot(nil), r.availability...), nil
}

func (r *fakeScheduleRepository) ListAppointments(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]model.AppointmentSummary, error) {
    r.appointmentCalls++
    return append([]model.AppointmentSummary(nil), r.appointments...), nil
}

type fakeAppointmentRepository struct {
    bookAppointmentID uuid.UUID
    cancelDoctorID    uuid.UUID
    completeVisit     model.CompleteVisitResult
    completeDoctorID  uuid.UUID
    paymentID         uuid.UUID
}

func (r *fakeAppointmentRepository) Book(_ context.Context, _ model.BookAppointmentCommand) (uuid.UUID, error) {
    return r.bookAppointmentID, nil
}

func (r *fakeAppointmentRepository) Cancel(_ context.Context, _ model.CancelAppointmentCommand) (uuid.UUID, error) {
    return r.cancelDoctorID, nil
}

func (r *fakeAppointmentRepository) CompleteVisit(_ context.Context, _ model.CompleteVisitCommand) (model.CompleteVisitResult, uuid.UUID, error) {
    return r.completeVisit, r.completeDoctorID, nil
}

func (r *fakeAppointmentRepository) RegisterPayment(_ context.Context, _ model.RegisterPaymentCommand) (uuid.UUID, error) {
    return r.paymentID, nil
}

func TestCatalogServiceCachesPriceList(t *testing.T) {
    t.Parallel()

    repo := &fakeCatalogRepository{
        items: []model.PriceListItem{{
            ServiceID:       uuid.MustParse("00000000-0000-0000-0000-000000000401"),
            Code:            "CONSULT_CARDIO",
            Name:            "Консультация кардиолога",
            DurationMinutes: 30,
            Price:           1500,
            CurrencyCode:    "RUB",
            EffectiveFrom:   time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
        }},
    }
    cache := newFakeCache()
    service := NewCatalogService(repo, cache, time.Minute)

    first, err := service.ListCurrentPriceList(context.Background())
    if err != nil {
        t.Fatalf("ожидалось успешное чтение прайс-листа: %v", err)
    }

    second, err := service.ListCurrentPriceList(context.Background())
    if err != nil {
        t.Fatalf("ожидалось чтение прайс-листа из кэша: %v", err)
    }

    if repo.calls != 1 {
        t.Fatalf("ожидался один вызов репозитория, получено %d", repo.calls)
    }

    if len(cache.setCalls) != 1 || cache.setCalls[0] != "clinic:catalog:pricelist" {
        t.Fatalf("ожидалась запись в ключ clinic:catalog:pricelist, получено %#v", cache.setCalls)
    }

    if !reflect.DeepEqual(first, second) {
        t.Fatalf("ожидался одинаковый результат из БД и из кэша")
    }
}

func TestScheduleServiceCachesDoctorSchedule(t *testing.T) {
    t.Parallel()

    doctorID := uuid.MustParse("00000000-0000-0000-0000-000000000301")
    from := time.Date(2030, time.January, 15, 0, 0, 0, 0, time.FixedZone("MSK", 3*60*60))
    to := time.Date(2030, time.January, 16, 23, 59, 59, 0, time.FixedZone("MSK", 3*60*60))

    repo := &fakeScheduleRepository{
        availability: []model.ScheduleSlot{{
            ID:                  uuid.MustParse("00000000-0000-0000-0000-000000000701"),
            DoctorID:            doctorID,
            StartsAt:            from,
            EndsAt:              from.Add(30 * time.Minute),
            SlotDurationMinutes: 30,
        }},
        appointments: []model.AppointmentSummary{{
            ID:              uuid.MustParse("00000000-0000-0000-0000-000000000801"),
            PatientID:       uuid.MustParse("00000000-0000-0000-0000-000000000601"),
            DoctorID:        doctorID,
            BookedServiceID: uuid.MustParse("00000000-0000-0000-0000-000000000401"),
            StartsAt:        from.Add(60 * time.Minute),
            EndsAt:          from.Add(90 * time.Minute),
            Status:          model.AppointmentStatusBooked,
        }},
    }
    cache := newFakeCache()
    service := NewScheduleService(repo, cache, time.Minute)

    first, err := service.GetDoctorSchedule(context.Background(), doctorID, from, to)
    if err != nil {
        t.Fatalf("ожидалось успешное получение расписания: %v", err)
    }

    second, err := service.GetDoctorSchedule(context.Background(), doctorID, from, to)
    if err != nil {
        t.Fatalf("ожидалось чтение расписания из кэша: %v", err)
    }

    if repo.availabilityCalls != 1 || repo.appointmentCalls != 1 {
        t.Fatalf("ожидался один вызов каждого метода репозитория, получено availability=%d appointments=%d", repo.availabilityCalls, repo.appointmentCalls)
    }

    if len(cache.setCalls) != 1 {
        t.Fatalf("ожидалась одна запись в кэш расписания, получено %d", len(cache.setCalls))
    }

    if !strings.HasPrefix(cache.setCalls[0], "clinic:schedule:doctor:"+doctorID.String()+":") {
        t.Fatalf("неожиданный ключ кэша расписания: %s", cache.setCalls[0])
    }

    if second.DoctorID != first.DoctorID || !second.From.Equal(first.From) || !second.To.Equal(first.To) {
        t.Fatalf("ожидались одинаковые метаданные расписания из БД и из кэша")
    }

    if len(second.Availability) != len(first.Availability) || len(second.Appointments) != len(first.Appointments) {
        t.Fatalf("ожидался одинаковый состав расписания из БД и из кэша")
    }

    if second.Availability[0].ID != first.Availability[0].ID || !second.Availability[0].StartsAt.Equal(first.Availability[0].StartsAt) || !second.Availability[0].EndsAt.Equal(first.Availability[0].EndsAt) {
        t.Fatalf("ожидались одинаковые интервалы доступности из БД и из кэша")
    }

    if second.Appointments[0].ID != first.Appointments[0].ID || !second.Appointments[0].StartsAt.Equal(first.Appointments[0].StartsAt) || !second.Appointments[0].EndsAt.Equal(first.Appointments[0].EndsAt) || second.Appointments[0].Status != first.Appointments[0].Status {
        t.Fatalf("ожидались одинаковые записи на прием из БД и из кэша")
    }
}

func TestAppointmentServiceInvalidatesScheduleCache(t *testing.T) {
    t.Parallel()

    doctorID := uuid.MustParse("00000000-0000-0000-0000-000000000301")
    otherDoctorID := uuid.MustParse("00000000-0000-0000-0000-000000000999")
    cache := newFakeCache()
    cache.values["clinic:schedule:doctor:"+doctorID.String()+":slot-a"] = "a"
    cache.values["clinic:schedule:doctor:"+otherDoctorID.String()+":slot-b"] = "b"

    repo := &fakeAppointmentRepository{
        bookAppointmentID: uuid.MustParse("00000000-0000-0000-0000-000000000901"),
        cancelDoctorID:    doctorID,
        completeVisit: model.CompleteVisitResult{
            VisitID:   uuid.MustParse("00000000-0000-0000-0000-000000000902"),
            InvoiceID: uuid.MustParse("00000000-0000-0000-0000-000000000904"),
        },
        completeDoctorID: doctorID,
        paymentID:        uuid.MustParse("00000000-0000-0000-0000-000000000903"),
    }
    service := NewAppointmentService(repo, cache)

    _, err := service.Book(context.Background(), model.BookAppointmentCommand{DoctorID: doctorID})
    if err != nil {
        t.Fatalf("ожидалась успешная запись на прием: %v", err)
    }

    if _, exists := cache.values["clinic:schedule:doctor:"+doctorID.String()+":slot-a"]; exists {
        t.Fatalf("ожидалась инвалидация кэша расписания врача после записи")
    }

    if _, exists := cache.values["clinic:schedule:doctor:"+otherDoctorID.String()+":slot-b"]; !exists {
        t.Fatalf("кэш другого врача не должен удаляться")
    }

    cache.values["clinic:schedule:doctor:"+doctorID.String()+":slot-c"] = "c"
    if err := service.Cancel(context.Background(), model.CancelAppointmentCommand{}); err != nil {
        t.Fatalf("ожидалась успешная отмена записи: %v", err)
    }

    if _, exists := cache.values["clinic:schedule:doctor:"+doctorID.String()+":slot-c"]; exists {
        t.Fatalf("ожидалась инвалидация кэша расписания врача после отмены")
    }

    cache.values["clinic:schedule:doctor:"+doctorID.String()+":slot-d"] = "d"
    result, err := service.Complete(context.Background(), model.CompleteVisitCommand{})
    if err != nil {
        t.Fatalf("ожидалось успешное завершение приема: %v", err)
    }
    if result.VisitID == uuid.Nil || result.InvoiceID == uuid.Nil {
        t.Fatalf("ожидался корректный результат завершения приема")
    }

    if _, exists := cache.values["clinic:schedule:doctor:"+doctorID.String()+":slot-d"]; exists {
        t.Fatalf("ожидалась инвалидация кэша расписания врача после завершения приема")
    }

    invalidationsBeforePayment := len(cache.deletedPatterns)
    if _, err := service.RegisterPayment(context.Background(), model.RegisterPaymentCommand{}); err != nil {
        t.Fatalf("ожидалась успешная регистрация оплаты: %v", err)
    }

    if len(cache.deletedPatterns) != invalidationsBeforePayment {
        t.Fatalf("регистрация оплаты не должна инвалидировать кэш расписания")
    }
}