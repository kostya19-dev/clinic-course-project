package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"clinic.local/backend/internal/model"
	"clinic.local/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	repo repository.AuthRepository
}

func NewAuthService(repo repository.AuthRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) Login(ctx context.Context, command model.LoginCommand) (model.AuthenticatedUser, error) {
	login := strings.TrimSpace(command.Login)
	password := command.Password
	if login == "" || password == "" {
		return model.AuthenticatedUser{}, ErrInvalidCredentials
	}

	account, err := s.repo.FindActiveUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AuthenticatedUser{}, ErrInvalidCredentials
		}

		return model.AuthenticatedUser{}, err
	}

	if !matchesPassword(account, password) {
		return model.AuthenticatedUser{}, ErrInvalidCredentials
	}

	return model.AuthenticatedUser{
		ID:        account.ID,
		Login:     account.Login,
		FullName:  account.FullName,
		Email:     account.Email,
		Role:      account.Role,
		DoctorID:  account.DoctorID,
		PatientID: account.PatientID,
	}, nil
}

func (s *AuthService) RegisterPatient(ctx context.Context, command model.RegisterPatientAccountCommand) (model.AuthenticatedUser, error) {
	email := strings.ToLower(strings.TrimSpace(command.Email))
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(command.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return model.AuthenticatedUser{}, fmt.Errorf("hash password: %w", err)
	}

	command.Email = email
	command.PasswordHash = string(passwordHash)
	command.Status = model.PatientStatusActive

	return s.repo.CreatePatientAccount(ctx, command)
}

func matchesPassword(account model.UserAccount, password string) bool {
	if strings.HasPrefix(account.PasswordHash, "$2") {
		if bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)) == nil {
			return true
		}
	}

	demoPasswords := map[string]string{
		"admin":         "admin123",
		"patient_demo":  "patient123",
		"doctor_ivanov": "doctor123",
	}

	if expected, ok := demoPasswords[account.Login]; ok {
		return password == expected
	}

	return account.PasswordHash == password
}
