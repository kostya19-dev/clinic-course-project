package postgres

import (
    "context"
    "database/sql"
    "time"

    "clinic.local/backend/internal/model"
    "github.com/google/uuid"
)

type ScheduleRepository struct {
    db *sql.DB
}

func NewScheduleRepository(db *sql.DB) *ScheduleRepository {
    return &ScheduleRepository{db: db}
}

func (r *ScheduleRepository) ListAvailability(ctx context.Context, doctorID uuid.UUID, from, to time.Time) ([]model.ScheduleSlot, error) {
    rows, err := r.db.QueryContext(ctx, `
        SELECT id, doctor_id, starts_at, ends_at, slot_duration_minutes
        FROM scheduling.doctor_availability
        WHERE doctor_id = $1
          AND is_active = TRUE
          AND starts_at < $3
          AND ends_at > $2
        ORDER BY starts_at
    `, doctorID, from, to)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    slots := make([]model.ScheduleSlot, 0)
    for rows.Next() {
        var slot model.ScheduleSlot
        if err := rows.Scan(
            &slot.ID,
            &slot.DoctorID,
            &slot.StartsAt,
            &slot.EndsAt,
            &slot.SlotDurationMinutes,
        ); err != nil {
            return nil, err
        }

        slots = append(slots, slot)
    }

    return slots, rows.Err()
}

func (r *ScheduleRepository) ListAppointments(ctx context.Context, doctorID uuid.UUID, from, to time.Time) ([]model.AppointmentSummary, error) {
    rows, err := r.db.QueryContext(ctx, `
        SELECT id, patient_id, doctor_id, booked_service_id, starts_at, ends_at, status
        FROM scheduling.appointments
        WHERE doctor_id = $1
          AND starts_at < $3
          AND ends_at > $2
          AND status IN ('booked', 'completed')
        ORDER BY starts_at
    `, doctorID, from, to)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    items := make([]model.AppointmentSummary, 0)
    for rows.Next() {
        var item model.AppointmentSummary
        var status string
        if err := rows.Scan(
            &item.ID,
            &item.PatientID,
            &item.DoctorID,
            &item.BookedServiceID,
            &item.StartsAt,
            &item.EndsAt,
            &status,
        ); err != nil {
            return nil, err
        }

        item.Status = model.AppointmentStatus(status)
        items = append(items, item)
    }

    return items, rows.Err()
}
