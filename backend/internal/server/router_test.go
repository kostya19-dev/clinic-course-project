package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"clinic.local/backend/internal/model"
	"clinic.local/backend/internal/validation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type patientServiceStub struct {
	createFn func(context.Context, model.CreatePatientCommand) (uuid.UUID, error)
}

func (s patientServiceStub) Create(ctx context.Context, command model.CreatePatientCommand) (uuid.UUID, error) {
	if s.createFn == nil {
		return uuid.Nil, nil
	}

	return s.createFn(ctx, command)
}

type catalogServiceStub struct{}

func (catalogServiceStub) ListCurrentPriceList(context.Context) ([]model.PriceListItem, error) {
	return []model.PriceListItem{}, nil
}

type scheduleServiceStub struct{}

func (scheduleServiceStub) GetDoctorSchedule(context.Context, uuid.UUID, time.Time, time.Time) (model.ScheduleView, error) {
	return model.ScheduleView{}, nil
}

type appointmentServiceStub struct {
	bookFn            func(context.Context, model.BookAppointmentCommand) (uuid.UUID, error)
	cancelFn          func(context.Context, model.CancelAppointmentCommand) error
	completeFn        func(context.Context, model.CompleteVisitCommand) (model.CompleteVisitResult, error)
	registerPaymentFn func(context.Context, model.RegisterPaymentCommand) (uuid.UUID, error)
}

func (s appointmentServiceStub) Book(ctx context.Context, command model.BookAppointmentCommand) (uuid.UUID, error) {
	if s.bookFn == nil {
		return uuid.Nil, nil
	}

	return s.bookFn(ctx, command)
}

func (s appointmentServiceStub) Cancel(ctx context.Context, command model.CancelAppointmentCommand) error {
	if s.cancelFn == nil {
		return nil
	}

	return s.cancelFn(ctx, command)
}

func (s appointmentServiceStub) Complete(ctx context.Context, command model.CompleteVisitCommand) (model.CompleteVisitResult, error) {
	if s.completeFn == nil {
		return model.CompleteVisitResult{}, nil
	}

	return s.completeFn(ctx, command)
}

func (s appointmentServiceStub) RegisterPayment(ctx context.Context, command model.RegisterPaymentCommand) (uuid.UUID, error) {
	if s.registerPaymentFn == nil {
		return uuid.Nil, nil
	}

	return s.registerPaymentFn(ctx, command)
}

type authServiceStub struct {
	loginFn           func(context.Context, model.LoginCommand) (model.AuthenticatedUser, error)
	registerPatientFn func(context.Context, model.RegisterPatientAccountCommand) (model.AuthenticatedUser, error)
}

func (s authServiceStub) Login(ctx context.Context, command model.LoginCommand) (model.AuthenticatedUser, error) {
	if s.loginFn == nil {
		return model.AuthenticatedUser{}, nil
	}

	return s.loginFn(ctx, command)
}

func (s authServiceStub) RegisterPatient(ctx context.Context, command model.RegisterPatientAccountCommand) (model.AuthenticatedUser, error) {
	if s.registerPatientFn == nil {
		return model.AuthenticatedUser{}, nil
	}

	return s.registerPatientFn(ctx, command)
}

var routerValidationOnce sync.Once

func prepareRouter(t *testing.T, patientSvc patientServiceStub, appointmentSvc appointmentServiceStub) *gin.Engine {
	return prepareRouterWithAuth(t, patientSvc, appointmentSvc, authServiceStub{})
}

func prepareRouterWithAuth(t *testing.T, patientSvc patientServiceStub, appointmentSvc appointmentServiceStub, authSvc authServiceStub) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	routerValidationOnce.Do(func() {
		if err := validation.Register(); err != nil {
			t.Fatalf("не удалось зарегистрировать валидаторы: %v", err)
		}
	})

	return NewRouter(patientSvc, catalogServiceStub{}, scheduleServiceStub{}, appointmentSvc, authSvc)
}

func performJSONRequest(t *testing.T, router http.Handler, method string, path string, payload any) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("не удалось сериализовать payload: %v", err)
	}

	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func decodeResponseBody(t *testing.T, recorder *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("не удалось декодировать тело ответа: %v", err)
	}

	return payload
}

func TestNewRouter_LoginReturnsRole(t *testing.T) {
	t.Helper()

	expectedID := uuid.MustParse("00000000-0000-0000-0000-000000000103")
	expectedDoctorID := uuid.MustParse("00000000-0000-0000-0000-000000000301")
	router := prepareRouterWithAuth(t, patientServiceStub{}, appointmentServiceStub{}, authServiceStub{
		loginFn: func(_ context.Context, command model.LoginCommand) (model.AuthenticatedUser, error) {
			if command.Login != "doctor_ivanov" {
				t.Fatalf("ожидался логин doctor_ivanov, получен %s", command.Login)
			}

			return model.AuthenticatedUser{
				ID:       expectedID,
				Login:    "doctor_ivanov",
				FullName: "Dr. Ivan Ivanov",
				Email:    "doctor@clinic.local",
				Role:     model.UserRoleDoctor,
				DoctorID: &expectedDoctorID,
			}, nil
		},
	})

	response := performJSONRequest(t, router, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"login":    "doctor_ivanov",
		"password": "doctor123",
	})

	if response.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200, получен %d", response.Code)
	}

	payload := decodeResponseBody(t, response)
	if payload["role"] != string(model.UserRoleDoctor) {
		t.Fatalf("ожидалась роль doctor, получено %#v", payload["role"])
	}
	if payload["doctorId"] != expectedDoctorID.String() {
		t.Fatalf("ожидался doctorId %s, получено %#v", expectedDoctorID, payload["doctorId"])
	}
}

func TestNewRouter_RegisterPatientReturnsCreatedSession(t *testing.T) {
	t.Helper()

	expectedUserID := uuid.MustParse("88888888-8888-8888-8888-888888888888")
	expectedPatientID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	router := prepareRouterWithAuth(t, patientServiceStub{}, appointmentServiceStub{}, authServiceStub{
		registerPatientFn: func(_ context.Context, command model.RegisterPatientAccountCommand) (model.AuthenticatedUser, error) {
			if command.Email != "new.patient@example.com" {
				t.Fatalf("ожидался email new.patient@example.com, получен %s", command.Email)
			}

			return model.AuthenticatedUser{
				ID:        expectedUserID,
				Login:     command.Email,
				FullName:  "Новый Пациент",
				Email:     command.Email,
				Role:      model.UserRolePatient,
				PatientID: &expectedPatientID,
			}, nil
		},
	})

	response := performJSONRequest(t, router, http.MethodPost, "/api/v1/auth/patient/register", map[string]any{
		"lastName":  "Пациент",
		"firstName": "Новый",
		"birthDate": "1999-01-12",
		"phone":     "+79995551122",
		"email":     "new.patient@example.com",
		"password":  "StrongPass123",
	})

	if response.Code != http.StatusCreated {
		t.Fatalf("ожидался статус 201, получен %d", response.Code)
	}

	payload := decodeResponseBody(t, response)
	if payload["role"] != string(model.UserRolePatient) {
		t.Fatalf("ожидалась роль patient, получено %#v", payload["role"])
	}
	if payload["patientId"] != expectedPatientID.String() {
		t.Fatalf("ожидался patientId %s, получено %#v", expectedPatientID, payload["patientId"])
	}
}

func TestNewRouter_CreatePatientReturnsCreated(t *testing.T) {
	t.Helper()

	expectedID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	router := prepareRouter(t, patientServiceStub{
		createFn: func(_ context.Context, command model.CreatePatientCommand) (uuid.UUID, error) {
			if command.Phone != "+79991112233" {
				t.Fatalf("ожидался телефон +79991112233, получен %s", command.Phone)
			}

			return expectedID, nil
		},
	}, appointmentServiceStub{})

	response := performJSONRequest(t, router, http.MethodPost, "/api/v1/patients", map[string]any{
		"lastName":  "Иванова",
		"firstName": "Анна",
		"birthDate": "1995-03-10",
		"phone":     "+79991112233",
		"email":     "anna@example.com",
		"status":    "active",
	})

	if response.Code != http.StatusCreated {
		t.Fatalf("ожидался статус 201, получен %d", response.Code)
	}

	payload := decodeResponseBody(t, response)
	if payload["id"] != expectedID.String() {
		t.Fatalf("ожидался идентификатор %s, получен %#v", expectedID, payload["id"])
	}
}

func TestNewRouter_CreatePatientReturnsBadRequestForInvalidPayload(t *testing.T) {
	t.Helper()

	router := prepareRouter(t, patientServiceStub{}, appointmentServiceStub{})

	response := performJSONRequest(t, router, http.MethodPost, "/api/v1/patients", map[string]any{
		"lastName":  "Иванова",
		"firstName": "Анна",
		"birthDate": "1995-03-10",
		"phone":     "12345",
		"email":     "anna@example.com",
		"status":    "active",
	})

	if response.Code != http.StatusBadRequest {
		t.Fatalf("ожидался статус 400, получен %d", response.Code)
	}

	payload := decodeResponseBody(t, response)
	if payload["error"] != "Телефон указан в неверном формате" {
		t.Fatalf("ожидалось сообщение о телефоне, получено %#v", payload["error"])
	}
}

func TestNewRouter_BookAppointmentReturnsCreated(t *testing.T) {
	t.Helper()

	expectedID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	router := prepareRouter(t, patientServiceStub{}, appointmentServiceStub{
		bookFn: func(_ context.Context, command model.BookAppointmentCommand) (uuid.UUID, error) {
			if command.Note != "Первичный прием" {
				t.Fatalf("ожидался комментарий Первичный прием, получен %s", command.Note)
			}

			return expectedID, nil
		},
	})

	response := performJSONRequest(t, router, http.MethodPost, "/api/v1/appointments", map[string]any{
		"patientId":       "00000000-0000-0000-0000-000000000601",
		"doctorId":        "00000000-0000-0000-0000-000000000301",
		"serviceId":       "00000000-0000-0000-0000-000000000401",
		"startsAt":        "2026-04-01T09:00:00Z",
		"createdByUserId": "00000000-0000-0000-0000-000000000102",
		"note":            "Первичный прием",
	})

	if response.Code != http.StatusCreated {
		t.Fatalf("ожидался статус 201, получен %d", response.Code)
	}

	payload := decodeResponseBody(t, response)
	if payload["id"] != expectedID.String() {
		t.Fatalf("ожидался идентификатор %s, получен %#v", expectedID, payload["id"])
	}
}

func TestNewRouter_BookAppointmentReturnsConflict(t *testing.T) {
	t.Helper()

	router := prepareRouter(t, patientServiceStub{}, appointmentServiceStub{
		bookFn: func(context.Context, model.BookAppointmentCommand) (uuid.UUID, error) {
			return uuid.Nil, &pgconn.PgError{Code: "23P01", Message: "conflict"}
		},
	})

	response := performJSONRequest(t, router, http.MethodPost, "/api/v1/appointments", map[string]any{
		"patientId":       "00000000-0000-0000-0000-000000000601",
		"doctorId":        "00000000-0000-0000-0000-000000000301",
		"serviceId":       "00000000-0000-0000-0000-000000000401",
		"startsAt":        "2026-04-01T09:00:00Z",
		"createdByUserId": "00000000-0000-0000-0000-000000000102",
		"note":            "Первичный прием",
	})

	if response.Code != http.StatusConflict {
		t.Fatalf("ожидался статус 409, получен %d", response.Code)
	}

	payload := decodeResponseBody(t, response)
	if payload["error"] != "Конфликт записи по времени" {
		t.Fatalf("ожидалось сообщение о конфликте записи, получено %#v", payload["error"])
	}
}

func TestNewRouter_CancelAppointmentReturnsNotFound(t *testing.T) {
	t.Helper()

	router := prepareRouter(t, patientServiceStub{}, appointmentServiceStub{
		cancelFn: func(_ context.Context, command model.CancelAppointmentCommand) error {
			if command.Reason != "Пациент передумал" {
				t.Fatalf("ожидалась причина Пациент передумал, получена %s", command.Reason)
			}

			return &pgconn.PgError{Code: "P0002", Message: "Запись не найдена"}
		},
	})

	response := performJSONRequest(t, router, http.MethodPost, "/api/v1/appointments/33333333-3333-3333-3333-333333333333/cancel", map[string]any{
		"actorUserId": "00000000-0000-0000-0000-000000000102",
		"reason":      "Пациент передумал",
	})

	if response.Code != http.StatusNotFound {
		t.Fatalf("ожидался статус 404, получен %d", response.Code)
	}

	payload := decodeResponseBody(t, response)
	if payload["error"] != "Запись не найдена" {
		t.Fatalf("ожидалось сообщение о ненайденной записи, получено %#v", payload["error"])
	}
}

func TestNewRouter_CancelAppointmentRejectsInvalidUUID(t *testing.T) {
	t.Helper()

	router := prepareRouter(t, patientServiceStub{}, appointmentServiceStub{})

	response := performJSONRequest(t, router, http.MethodPost, "/api/v1/appointments/not-a-uuid/cancel", map[string]any{
		"actorUserId": "00000000-0000-0000-0000-000000000102",
		"reason":      "Пациент передумал",
	})

	if response.Code != http.StatusBadRequest {
		t.Fatalf("ожидался статус 400, получен %d", response.Code)
	}

	payload := decodeResponseBody(t, response)
	if payload["error"] != "Некорректный идентификатор записи" {
		t.Fatalf("ожидалось сообщение о некорректном идентификаторе, получено %#v", payload["error"])
	}
}

func TestNewRouter_CompleteVisitReturnsCreated(t *testing.T) {
	t.Helper()

	expectedVisitID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	expectedInvoiceID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	router := prepareRouter(t, patientServiceStub{}, appointmentServiceStub{
		completeFn: func(_ context.Context, command model.CompleteVisitCommand) (model.CompleteVisitResult, error) {
			if command.DoctorUserID != uuid.MustParse("00000000-0000-0000-0000-000000000103") {
				t.Fatalf("ожидался идентификатор врача 00000000-0000-0000-0000-000000000103, получен %s", command.DoctorUserID)
			}

			return model.CompleteVisitResult{
				VisitID:   expectedVisitID,
				InvoiceID: expectedInvoiceID,
			}, nil
		},
	})

	response := performJSONRequest(t, router, http.MethodPost, "/api/v1/appointments/33333333-3333-3333-3333-333333333333/complete", map[string]any{
		"doctorUserId": "00000000-0000-0000-0000-000000000103",
		"summary":      "Прием завершен",
		"diagnosis":    "I10",
		"services": []map[string]any{{
			"service_id": "00000000-0000-0000-0000-000000000402",
			"quantity":   1,
		}},
	})

	if response.Code != http.StatusCreated {
		t.Fatalf("ожидался статус 201, получен %d", response.Code)
	}

	payload := decodeResponseBody(t, response)
	if payload["visitId"] != expectedVisitID.String() {
		t.Fatalf("ожидался visitId %s, получен %#v", expectedVisitID, payload["visitId"])
	}
	if payload["invoiceId"] != expectedInvoiceID.String() {
		t.Fatalf("ожидался invoiceId %s, получен %#v", expectedInvoiceID, payload["invoiceId"])
	}
}

func TestNewRouter_RegisterPaymentReturnsCreated(t *testing.T) {
	t.Helper()

	expectedPaymentID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	router := prepareRouter(t, patientServiceStub{}, appointmentServiceStub{
		registerPaymentFn: func(_ context.Context, command model.RegisterPaymentCommand) (uuid.UUID, error) {
			if command.PaymentMethod != model.PaymentMethodCard {
				t.Fatalf("ожидался способ оплаты card, получен %s", command.PaymentMethod)
			}
			return expectedPaymentID, nil
		},
	})

	response := performJSONRequest(t, router, http.MethodPost, "/api/v1/invoices/77777777-7777-7777-7777-777777777777/payments", map[string]any{
		"receivedByUserId":  "00000000-0000-0000-0000-000000000104",
		"amount":            1500,
		"paymentMethod":     "card",
		"externalReference": "PAY-001",
	})

	if response.Code != http.StatusCreated {
		t.Fatalf("ожидался статус 201, получен %d", response.Code)
	}

	payload := decodeResponseBody(t, response)
	if payload["id"] != expectedPaymentID.String() {
		t.Fatalf("ожидался id %s, получен %#v", expectedPaymentID, payload["id"])
	}
}
