package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AppointmentStatus string

type PatientStatus string

type PaymentMethod string

type UserRole string

const (
	AppointmentStatusBooked    AppointmentStatus = "booked"
	AppointmentStatusCancelled AppointmentStatus = "cancelled"
	AppointmentStatusCompleted AppointmentStatus = "completed"
)

const (
	PatientStatusActive   PatientStatus = "active"
	PatientStatusInactive PatientStatus = "inactive"
	PatientStatusArchived PatientStatus = "archived"
)

const (
	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodCard     PaymentMethod = "card"
	PaymentMethodTransfer PaymentMethod = "transfer"
)

const (
	UserRoleAdministrator UserRole = "administrator"
	UserRolePatient       UserRole = "patient"
	UserRoleDoctor        UserRole = "doctor"
)

type Date struct {
	time.Time
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return err
	}

	d.Time = parsed
	return nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return json.Marshal("")
	}

	return json.Marshal(d.Format("2006-01-02"))
}

func (d Date) Value() (driver.Value, error) {
	if d.Time.IsZero() {
		return nil, nil
	}

	return d.Time, nil
}

func (d Date) IsZero() bool {
	return d.Time.IsZero()
}

type Patient struct {
	ID         uuid.UUID     `json:"id"`
	LastName   string        `json:"lastName"`
	FirstName  string        `json:"firstName"`
	MiddleName string        `json:"middleName,omitempty"`
	BirthDate  Date          `json:"birthDate"`
	Phone      string        `json:"phone"`
	Email      string        `json:"email,omitempty"`
	Status     PatientStatus `json:"status"`
}

type PriceListItem struct {
	ServiceID       uuid.UUID `json:"serviceId"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	DurationMinutes int       `json:"durationMinutes"`
	Price           float64   `json:"price"`
	CurrencyCode    string    `json:"currencyCode"`
	EffectiveFrom   time.Time `json:"effectiveFrom"`
}

type ScheduleSlot struct {
	ID                  uuid.UUID `json:"id"`
	DoctorID            uuid.UUID `json:"doctorId"`
	StartsAt            time.Time `json:"startsAt"`
	EndsAt              time.Time `json:"endsAt"`
	SlotDurationMinutes int       `json:"slotDurationMinutes"`
}

type AppointmentSummary struct {
	ID              uuid.UUID         `json:"id"`
	PatientID       uuid.UUID         `json:"patientId"`
	DoctorID        uuid.UUID         `json:"doctorId"`
	BookedServiceID uuid.UUID         `json:"bookedServiceId"`
	StartsAt        time.Time         `json:"startsAt"`
	EndsAt          time.Time         `json:"endsAt"`
	Status          AppointmentStatus `json:"status"`
}

type ScheduleView struct {
	DoctorID     uuid.UUID            `json:"doctorId"`
	From         time.Time            `json:"from"`
	To           time.Time            `json:"to"`
	Availability []ScheduleSlot       `json:"availability"`
	Appointments []AppointmentSummary `json:"appointments"`
}

type CompleteVisitResult struct {
	VisitID   uuid.UUID `json:"visitId"`
	InvoiceID uuid.UUID `json:"invoiceId"`
}

type UserAccount struct {
	ID           uuid.UUID
	Login        string
	PasswordHash string
	FullName     string
	Email        string
	Role         UserRole
	DoctorID     *uuid.UUID
	PatientID    *uuid.UUID
}

type AuthenticatedUser struct {
	ID        uuid.UUID  `json:"id"`
	Login     string     `json:"login"`
	FullName  string     `json:"fullName"`
	Email     string     `json:"email"`
	Role      UserRole   `json:"role"`
	DoctorID  *uuid.UUID `json:"doctorId,omitempty"`
	PatientID *uuid.UUID `json:"patientId,omitempty"`
}

type LoginCommand struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegisterPatientAccountCommand struct {
	LastName     string        `json:"lastName"`
	FirstName    string        `json:"firstName"`
	MiddleName   string        `json:"middleName"`
	BirthDate    Date          `json:"birthDate"`
	Phone        string        `json:"phone"`
	Email        string        `json:"email"`
	PasswordHash string        `json:"passwordHash"`
	Status       PatientStatus `json:"status"`
}

type CreatePatientCommand struct {
	LastName   string        `json:"lastName"`
	FirstName  string        `json:"firstName"`
	MiddleName string        `json:"middleName"`
	BirthDate  Date          `json:"birthDate"`
	Phone      string        `json:"phone"`
	Email      string        `json:"email"`
	Status     PatientStatus `json:"status"`
}

type BookAppointmentCommand struct {
	PatientID       uuid.UUID `json:"patientId"`
	DoctorID        uuid.UUID `json:"doctorId"`
	ServiceID       uuid.UUID `json:"serviceId"`
	StartsAt        time.Time `json:"startsAt"`
	CreatedByUserID uuid.UUID `json:"createdByUserId"`
	Note            string    `json:"note"`
}

type CancelAppointmentCommand struct {
	AppointmentID uuid.UUID `json:"appointmentId"`
	ActorUserID   uuid.UUID `json:"actorUserId"`
	Reason        string    `json:"reason"`
}

type PerformedService struct {
	ServiceID uuid.UUID `json:"service_id"`
	Quantity  int       `json:"quantity"`
}

type CompleteVisitCommand struct {
	AppointmentID uuid.UUID          `json:"appointmentId"`
	DoctorUserID  uuid.UUID          `json:"doctorUserId"`
	Summary       string             `json:"summary"`
	Diagnosis     string             `json:"diagnosis"`
	Services      []PerformedService `json:"services"`
}

type RegisterPaymentCommand struct {
	InvoiceID         uuid.UUID     `json:"invoiceId"`
	ReceivedByUserID  uuid.UUID     `json:"receivedByUserId"`
	Amount            float64       `json:"amount"`
	PaymentMethod     PaymentMethod `json:"paymentMethod"`
	ExternalReference string        `json:"externalReference"`
}
