package postgres

import (
    "context"
    "database/sql"
    "encoding/json"

    "clinic.local/backend/internal/model"
    "github.com/google/uuid"
)

type AppointmentRepository struct {
    db *sql.DB
}

func NewAppointmentRepository(db *sql.DB) *AppointmentRepository {
    return &AppointmentRepository{db: db}
}

func (r *AppointmentRepository) Book(ctx context.Context, command model.BookAppointmentCommand) (uuid.UUID, error) {
    var appointmentID uuid.UUID
    err := r.db.QueryRowContext(ctx, `
        SELECT scheduling.book_appointment($1, $2, $3, $4, $5, $6)
    `,
        command.PatientID,
        command.DoctorID,
        command.ServiceID,
        command.StartsAt,
        command.CreatedByUserID,
        nullableString(command.Note),
    ).Scan(&appointmentID)

    return appointmentID, err
}

func (r *AppointmentRepository) Cancel(ctx context.Context, command model.CancelAppointmentCommand) (uuid.UUID, error) {
    var doctorID uuid.UUID
    if err := r.db.QueryRowContext(ctx, `
        SELECT doctor_id
        FROM scheduling.appointments
        WHERE id = $1
    `, command.AppointmentID).Scan(&doctorID); err != nil {
        return uuid.Nil, err
    }

    _, err := r.db.ExecContext(ctx, `
        SELECT scheduling.cancel_appointment($1, $2, $3)
    `, command.AppointmentID, command.ActorUserID, command.Reason)
    if err != nil {
        return uuid.Nil, err
    }

    return doctorID, nil
}

func (r *AppointmentRepository) CompleteVisit(ctx context.Context, command model.CompleteVisitCommand) (model.CompleteVisitResult, uuid.UUID, error) {
    payload, err := json.Marshal(command.Services)
    if err != nil {
        return model.CompleteVisitResult{}, uuid.Nil, err
    }

    var doctorID uuid.UUID
    if err := r.db.QueryRowContext(ctx, `
        SELECT doctor_id
        FROM scheduling.appointments
        WHERE id = $1
    `, command.AppointmentID).Scan(&doctorID); err != nil {
        return model.CompleteVisitResult{}, uuid.Nil, err
    }

    result := model.CompleteVisitResult{}
    err = r.db.QueryRowContext(ctx, `
        SELECT medical.complete_visit($1, $2, $3, $4, $5::jsonb)
    `,
        command.AppointmentID,
        command.DoctorUserID,
        nullableString(command.Summary),
        nullableString(command.Diagnosis),
        string(payload),
    ).Scan(&result.VisitID)
    if err != nil {
        return model.CompleteVisitResult{}, uuid.Nil, err
    }

    err = r.db.QueryRowContext(ctx, `
        SELECT id
        FROM billing.invoices
        WHERE visit_id = $1
    `, result.VisitID).Scan(&result.InvoiceID)
    if err != nil {
        return model.CompleteVisitResult{}, uuid.Nil, err
    }

    return result, doctorID, nil
}

func (r *AppointmentRepository) RegisterPayment(ctx context.Context, command model.RegisterPaymentCommand) (uuid.UUID, error) {
    var paymentID uuid.UUID
    err := r.db.QueryRowContext(ctx, `
        SELECT billing.register_payment($1, $2, $3, $4, $5)
    `,
        command.InvoiceID,
        command.Amount,
        string(command.PaymentMethod),
        command.ReceivedByUserID,
        nullableString(command.ExternalReference),
    ).Scan(&paymentID)

    return paymentID, err
}