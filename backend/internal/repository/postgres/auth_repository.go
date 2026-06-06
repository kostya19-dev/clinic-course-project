package postgres

import (
	"context"
	"database/sql"

	"clinic.local/backend/internal/model"
	"github.com/google/uuid"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) FindActiveUserByLogin(ctx context.Context, login string) (model.UserAccount, error) {
	account := model.UserAccount{}
	var userID string
	var role string
	var doctorID sql.NullString
	var patientID sql.NullString

	err := r.db.QueryRowContext(ctx, `
        SELECT
            u.id::text,
            u.login,
            u.password_hash,
            u.full_name,
            COALESCE(u.email, ''),
            r.code,
            d.id::text,
            p.id::text
        FROM security.users u
        JOIN security.user_roles ur ON ur.user_id = u.id
        JOIN security.roles r ON r.id = ur.role_id
        LEFT JOIN medical.doctors d ON d.user_id = u.id AND d.status = 'active'
        LEFT JOIN medical.patients p ON p.status = 'active'
            AND (
                (p.email IS NOT NULL AND u.email IS NOT NULL AND lower(p.email) = lower(u.email))
                OR (u.phone IS NOT NULL AND p.phone = u.phone)
            )
        WHERE (
                lower(u.login) = lower($1)
                OR lower(COALESCE(u.email, '')) = lower($1)
                OR u.phone = $1
            )
          AND u.status = 'active'
        ORDER BY CASE r.code
            WHEN 'administrator' THEN 1
            WHEN 'doctor' THEN 2
            WHEN 'patient' THEN 3
            ELSE 4
        END
        LIMIT 1
    `, login).Scan(
		&userID,
		&account.Login,
		&account.PasswordHash,
		&account.FullName,
		&account.Email,
		&role,
		&doctorID,
		&patientID,
	)
	if err != nil {
		return model.UserAccount{}, err
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return model.UserAccount{}, err
	}

	account.ID = parsedUserID
	account.Role = model.UserRole(role)
	if doctorID.Valid {
		parsedDoctorID, err := uuid.Parse(doctorID.String)
		if err != nil {
			return model.UserAccount{}, err
		}

		account.DoctorID = &parsedDoctorID
	}
	if patientID.Valid {
		parsedPatientID, err := uuid.Parse(patientID.String)
		if err != nil {
			return model.UserAccount{}, err
		}

		account.PatientID = &parsedPatientID
	}

	return account, nil
}

func (r *AuthRepository) CreatePatientAccount(ctx context.Context, command model.RegisterPatientAccountCommand) (model.AuthenticatedUser, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.AuthenticatedUser{}, err
	}
	defer tx.Rollback()

	fullName := command.LastName + " " + command.FirstName
	if command.MiddleName != "" {
		fullName += " " + command.MiddleName
	}

	var userID uuid.UUID
	err = tx.QueryRowContext(ctx, `
        INSERT INTO security.users (
            login,
            password_hash,
            full_name,
            email,
            phone,
            status
        )
        VALUES ($1, $2, $3, $4, $5, 'active')
        RETURNING id
    `,
		command.Email,
		command.PasswordHash,
		fullName,
		command.Email,
		command.Phone,
	).Scan(&userID)
	if err != nil {
		return model.AuthenticatedUser{}, err
	}

	var patientID uuid.UUID
	err = tx.QueryRowContext(ctx, `
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
		command.Email,
		string(command.Status),
	).Scan(&patientID)
	if err != nil {
		return model.AuthenticatedUser{}, err
	}

	_, err = tx.ExecContext(ctx, `
        INSERT INTO security.user_roles (user_id, role_id)
        SELECT $1, id
        FROM security.roles
        WHERE code = 'patient'
    `, userID)
	if err != nil {
		return model.AuthenticatedUser{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.AuthenticatedUser{}, err
	}

	return model.AuthenticatedUser{
		ID:        userID,
		Login:     command.Email,
		FullName:  fullName,
		Email:     command.Email,
		Role:      model.UserRolePatient,
		PatientID: &patientID,
	}, nil
}
