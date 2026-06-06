package dto

import (
	"time"

	"clinic.local/backend/internal/model"
	"github.com/google/uuid"
)

type CreatePatientRequest struct {
	LastName   string              `json:"lastName" binding:"required,min=1,max=100"`
	FirstName  string              `json:"firstName" binding:"required,min=1,max=100"`
	MiddleName string              `json:"middleName" binding:"omitempty,max=100"`
	BirthDate  model.Date          `json:"birthDate" binding:"required,date_required"`
	Phone      string              `json:"phone" binding:"required,phone"`
	Email      string              `json:"email" binding:"omitempty,email,max=150"`
	Status     model.PatientStatus `json:"status" binding:"required,oneof=active inactive archived"`
}

type BookAppointmentRequest struct {
	PatientID       uuid.UUID `json:"patientId" binding:"required,not_nil_uuid"`
	DoctorID        uuid.UUID `json:"doctorId" binding:"required,not_nil_uuid"`
	ServiceID       uuid.UUID `json:"serviceId" binding:"required,not_nil_uuid"`
	StartsAt        time.Time `json:"startsAt" binding:"required"`
	CreatedByUserID uuid.UUID `json:"createdByUserId" binding:"required,not_nil_uuid"`
	Note            string    `json:"note" binding:"omitempty,max=500"`
}

type CancelAppointmentRequest struct {
	ActorUserID uuid.UUID `json:"actorUserId" binding:"required,not_nil_uuid"`
	Reason      string    `json:"reason" binding:"required,min=3,max=500"`
}

type CompleteVisitRequest struct {
	DoctorUserID uuid.UUID                `json:"doctorUserId" binding:"required,not_nil_uuid"`
	Summary      string                   `json:"summary" binding:"omitempty,max=2000"`
	Diagnosis    string                   `json:"diagnosis" binding:"omitempty,max=2000"`
	Services     []model.PerformedService `json:"services"`
}

type RegisterPaymentRequest struct {
	ReceivedByUserID  uuid.UUID           `json:"receivedByUserId" binding:"required,not_nil_uuid"`
	Amount            float64             `json:"amount" binding:"required,gt=0"`
	PaymentMethod     model.PaymentMethod `json:"paymentMethod" binding:"required,oneof=cash card transfer"`
	ExternalReference string              `json:"externalReference" binding:"omitempty,max=100"`
}

type LoginRequest struct {
	Login    string `json:"login" binding:"required,min=3,max=100"`
	Password string `json:"password" binding:"required,min=4,max=100"`
}

type RegisterPatientAccountRequest struct {
	LastName   string     `json:"lastName" binding:"required,min=1,max=100"`
	FirstName  string     `json:"firstName" binding:"required,min=1,max=100"`
	MiddleName string     `json:"middleName" binding:"omitempty,max=100"`
	BirthDate  model.Date `json:"birthDate" binding:"required,date_required"`
	Phone      string     `json:"phone" binding:"required,phone"`
	Email      string     `json:"email" binding:"required,email,max=150"`
	Password   string     `json:"password" binding:"required,min=8,max=100"`
}

type IDResponse struct {
	ID uuid.UUID `json:"id"`
}

type CompleteVisitResponse struct {
	VisitID   uuid.UUID `json:"visitId"`
	InvoiceID uuid.UUID `json:"invoiceId"`
}

type LoginResponse struct {
	ID        uuid.UUID      `json:"id"`
	Login     string         `json:"login"`
	FullName  string         `json:"fullName"`
	Email     string         `json:"email"`
	Role      model.UserRole `json:"role"`
	DoctorID  *uuid.UUID     `json:"doctorId,omitempty"`
	PatientID *uuid.UUID     `json:"patientId,omitempty"`
}
