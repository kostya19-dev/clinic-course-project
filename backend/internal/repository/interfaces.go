package repository

import (
	"context"
	"time"

	"clinic.local/backend/internal/model"
	"github.com/google/uuid"
)

type PatientRepository interface {
	Create(ctx context.Context, command model.CreatePatientCommand) (uuid.UUID, error)
}

type CatalogRepository interface {
	ListCurrentPriceList(ctx context.Context) ([]model.PriceListItem, error)
}

type ScheduleRepository interface {
	ListAvailability(ctx context.Context, doctorID uuid.UUID, from, to time.Time) ([]model.ScheduleSlot, error)
	ListAppointments(ctx context.Context, doctorID uuid.UUID, from, to time.Time) ([]model.AppointmentSummary, error)
}

type AppointmentRepository interface {
	Book(ctx context.Context, command model.BookAppointmentCommand) (uuid.UUID, error)
	Cancel(ctx context.Context, command model.CancelAppointmentCommand) (uuid.UUID, error)
	CompleteVisit(ctx context.Context, command model.CompleteVisitCommand) (model.CompleteVisitResult, uuid.UUID, error)
	RegisterPayment(ctx context.Context, command model.RegisterPaymentCommand) (uuid.UUID, error)
}

type AuthRepository interface {
	FindActiveUserByLogin(ctx context.Context, login string) (model.UserAccount, error)
	CreatePatientAccount(ctx context.Context, command model.RegisterPatientAccountCommand) (model.AuthenticatedUser, error)
}
