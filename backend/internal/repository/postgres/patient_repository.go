package postgres

import (
    "context"
    "database/sql"

    "clinic.local/backend/internal/model"
    "github.com/google/uuid"
)

type PatientRepository struct {
    db *sql.DB
}

func NewPatientRepository(db *sql.DB) *PatientRepository {
    return &PatientRepository{db: db}
}

func (r *PatientRepository) Create(ctx context.Context, command model.CreatePatientCommand) (uuid.UUID, error) {
    var patientID uuid.UUID
    err := r.db.QueryRowContext(ctx, `
        INSERT INTO medical.patients (
            last_name,
            first_name,
            middle_name,
            birth_date,
            phone,
            email,
            status
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id
    `,
        command.LastName,
        command.FirstName,
        command.MiddleName,
        command.BirthDate.Time,
        command.Phone,
        nullableString(command.Email),
        string(command.Status),
    ).Scan(&patientID)

    return patientID, err
}

func nullableString(value string) any {
    if value == "" {
        return nil
    }

    return value
}
