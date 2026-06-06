package service

import (
    "context"
    "fmt"

    cachepkg "clinic.local/backend/internal/cache"
    "clinic.local/backend/internal/model"
    "clinic.local/backend/internal/repository"
    "github.com/google/uuid"
)

type AppointmentService struct {
    repo  repository.AppointmentRepository
    cache cachepkg.Client
}

func NewAppointmentService(repo repository.AppointmentRepository, cache cachepkg.Client) *AppointmentService {
    return &AppointmentService{repo: repo, cache: cache}
}

func (s *AppointmentService) Book(ctx context.Context, command model.BookAppointmentCommand) (uuid.UUID, error) {
    appointmentID, err := s.repo.Book(ctx, command)
    if err != nil {
        return uuid.Nil, err
    }

    s.invalidateDoctorSchedule(ctx, command.DoctorID)
    return appointmentID, nil
}

func (s *AppointmentService) Cancel(ctx context.Context, command model.CancelAppointmentCommand) error {
    doctorID, err := s.repo.Cancel(ctx, command)
    if err != nil {
        return err
    }

    s.invalidateDoctorSchedule(ctx, doctorID)
    return nil
}

func (s *AppointmentService) Complete(ctx context.Context, command model.CompleteVisitCommand) (model.CompleteVisitResult, error) {
    result, doctorID, err := s.repo.CompleteVisit(ctx, command)
    if err != nil {
        return model.CompleteVisitResult{}, err
    }

    s.invalidateDoctorSchedule(ctx, doctorID)
    return result, nil
}

func (s *AppointmentService) RegisterPayment(ctx context.Context, command model.RegisterPaymentCommand) (uuid.UUID, error) {
    return s.repo.RegisterPayment(ctx, command)
}

func (s *AppointmentService) invalidateDoctorSchedule(ctx context.Context, doctorID uuid.UUID) {
    if s.cache == nil {
        return
    }

    pattern := fmt.Sprintf("clinic:schedule:doctor:%s:*", doctorID.String())
    _ = s.cache.DeleteByPattern(ctx, pattern)
}