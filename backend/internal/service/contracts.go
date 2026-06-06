package service

import (
	"context"
	"time"

	"clinic.local/backend/internal/model"
	"github.com/google/uuid"
)

type PatientCreator interface {
	Create(ctx context.Context, command model.CreatePatientCommand) (uuid.UUID, error)
}

type CatalogReader interface {
	ListCurrentPriceList(ctx context.Context) ([]model.PriceListItem, error)
}

type ScheduleReader interface {
	GetDoctorSchedule(ctx context.Context, doctorID uuid.UUID, from, to time.Time) (model.ScheduleView, error)
}

type AppointmentManager interface {
	Book(ctx context.Context, command model.BookAppointmentCommand) (uuid.UUID, error)
	Cancel(ctx context.Context, command model.CancelAppointmentCommand) error
	Complete(ctx context.Context, command model.CompleteVisitCommand) (model.CompleteVisitResult, error)
	RegisterPayment(ctx context.Context, command model.RegisterPaymentCommand) (uuid.UUID, error)
}

type Authenticator interface {
	Login(ctx context.Context, command model.LoginCommand) (model.AuthenticatedUser, error)
	RegisterPatient(ctx context.Context, command model.RegisterPatientAccountCommand) (model.AuthenticatedUser, error)
}
