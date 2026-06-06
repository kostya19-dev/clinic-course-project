package service

import (
    "context"

    "clinic.local/backend/internal/model"
    "clinic.local/backend/internal/repository"
    "github.com/google/uuid"
)

type PatientService struct {
    repo repository.PatientRepository
}

func NewPatientService(repo repository.PatientRepository) *PatientService {
    return &PatientService{repo: repo}
}

func (s *PatientService) Create(ctx context.Context, command model.CreatePatientCommand) (uuid.UUID, error) {
    return s.repo.Create(ctx, command)
}
