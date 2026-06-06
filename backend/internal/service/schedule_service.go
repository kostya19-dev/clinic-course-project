package service

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "time"

    cachepkg "clinic.local/backend/internal/cache"
    "clinic.local/backend/internal/model"
    "clinic.local/backend/internal/repository"
    "github.com/google/uuid"
)

type ScheduleService struct {
    repo  repository.ScheduleRepository
    cache cachepkg.Client
    ttl   time.Duration
}

func NewScheduleService(repo repository.ScheduleRepository, cache cachepkg.Client, ttl time.Duration) *ScheduleService {
    return &ScheduleService{repo: repo, cache: cache, ttl: ttl}
}

func (s *ScheduleService) GetDoctorSchedule(ctx context.Context, doctorID uuid.UUID, from, to time.Time) (model.ScheduleView, error) {
    cacheKey := fmt.Sprintf("clinic:schedule:doctor:%s:%s:%s", doctorID.String(), from.UTC().Format(time.RFC3339), to.UTC().Format(time.RFC3339))

    if s.cache != nil {
        raw, err := s.cache.Get(ctx, cacheKey)
        if err == nil {
            var cached model.ScheduleView
            if unmarshalErr := json.Unmarshal([]byte(raw), &cached); unmarshalErr == nil {
                return cached, nil
            }
        } else if !errors.Is(err, cachepkg.ErrCacheMiss) {
            return model.ScheduleView{}, err
        }
    }

    availability, err := s.repo.ListAvailability(ctx, doctorID, from, to)
    if err != nil {
        return model.ScheduleView{}, err
    }

    appointments, err := s.repo.ListAppointments(ctx, doctorID, from, to)
    if err != nil {
        return model.ScheduleView{}, err
    }

    view := model.ScheduleView{
        DoctorID:     doctorID,
        From:         from,
        To:           to,
        Availability: availability,
        Appointments: appointments,
    }

    if s.cache != nil {
        if payload, marshalErr := json.Marshal(view); marshalErr == nil {
            _ = s.cache.Set(ctx, cacheKey, string(payload), s.ttl)
        }
    }

    return view, nil
}
