package validation

import (
    "sync"
    "testing"
    "time"

    "clinic.local/backend/internal/dto"
    "clinic.local/backend/internal/model"
    "github.com/gin-gonic/gin/binding"
    "github.com/go-playground/validator/v10"
    "github.com/google/uuid"
)

var registerOnce sync.Once

func prepareValidator(t *testing.T) *validator.Validate {
    t.Helper()

    registerOnce.Do(func() {
        if err := Register(); err != nil {
            t.Fatalf("не удалось зарегистрировать валидаторы: %v", err)
        }
    })

    engine, ok := binding.Validator.Engine().(*validator.Validate)
    if !ok {
        t.Fatal("не удалось получить validator engine")
    }

    return engine
}

func TestCreatePatientRequest_Valid(t *testing.T) {
    t.Helper()

    engine := prepareValidator(t)
    request := dto.CreatePatientRequest{
        LastName:  "Иванова",
        FirstName: "Анна",
        BirthDate: model.Date{Time: time.Date(1995, time.March, 10, 0, 0, 0, 0, time.UTC)},
        Phone:     "+79991112233",
        Email:     "anna@example.com",
        Status:    model.PatientStatusActive,
    }

    if err := engine.Struct(request); err != nil {
        t.Fatalf("ожидалась успешная валидация, получена ошибка: %v", err)
    }
}

func TestCreatePatientRequest_InvalidPhone(t *testing.T) {
    t.Helper()

    engine := prepareValidator(t)
    request := dto.CreatePatientRequest{
        LastName:  "Иванова",
        FirstName: "Анна",
        BirthDate: model.Date{Time: time.Date(1995, time.March, 10, 0, 0, 0, 0, time.UTC)},
        Phone:     "12345",
        Email:     "anna@example.com",
        Status:    model.PatientStatusActive,
    }

    if err := engine.Struct(request); err == nil {
        t.Fatal("ожидалась ошибка валидации телефона")
    }
}

func TestCreatePatientRequest_InvalidEmail(t *testing.T) {
    t.Helper()

    engine := prepareValidator(t)
    request := dto.CreatePatientRequest{
        LastName:  "Иванова",
        FirstName: "Анна",
        BirthDate: model.Date{Time: time.Date(1995, time.March, 10, 0, 0, 0, 0, time.UTC)},
        Phone:     "+79991112233",
        Email:     "not-an-email",
        Status:    model.PatientStatusActive,
    }

    if err := engine.Struct(request); err == nil {
        t.Fatal("ожидалась ошибка валидации email")
    }
}

func TestCreatePatientRequest_InvalidStatus(t *testing.T) {
    t.Helper()

    engine := prepareValidator(t)
    request := dto.CreatePatientRequest{
        LastName:  "Иванова",
        FirstName: "Анна",
        BirthDate: model.Date{Time: time.Date(1995, time.March, 10, 0, 0, 0, 0, time.UTC)},
        Phone:     "+79991112233",
        Email:     "anna@example.com",
        Status:    model.PatientStatus("unexpected"),
    }

    if err := engine.Struct(request); err == nil {
        t.Fatal("ожидалась ошибка валидации статуса")
    }
}

func TestBookAppointmentRequest_RejectsNilUUID(t *testing.T) {
    t.Helper()

    engine := prepareValidator(t)
    request := dto.BookAppointmentRequest{
        PatientID:       uuid.Nil,
        DoctorID:        uuid.New(),
        ServiceID:       uuid.New(),
        StartsAt:        time.Now().Add(time.Hour),
        CreatedByUserID: uuid.New(),
    }

    if err := engine.Struct(request); err == nil {
        t.Fatal("ожидалась ошибка валидации nil UUID")
    }
}
