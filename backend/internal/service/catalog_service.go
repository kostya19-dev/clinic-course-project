package service

import (
    "context"
    "encoding/json"
    "errors"
    "time"

    cachepkg "clinic.local/backend/internal/cache"
    "clinic.local/backend/internal/model"
    "clinic.local/backend/internal/repository"
)

type CatalogService struct {
    repo  repository.CatalogRepository
    cache cachepkg.Client
    ttl   time.Duration
}

func NewCatalogService(repo repository.CatalogRepository, cache cachepkg.Client, ttl time.Duration) *CatalogService {
    return &CatalogService{repo: repo, cache: cache, ttl: ttl}
}

func (s *CatalogService) ListCurrentPriceList(ctx context.Context) ([]model.PriceListItem, error) {
    const cacheKey = "clinic:catalog:pricelist"

    if s.cache != nil {
        raw, err := s.cache.Get(ctx, cacheKey)
        if err == nil {
            var cached []model.PriceListItem
            if unmarshalErr := json.Unmarshal([]byte(raw), &cached); unmarshalErr == nil {
                return cached, nil
            }
        } else if !errors.Is(err, cachepkg.ErrCacheMiss) {
            return nil, err
        }
    }

    items, err := s.repo.ListCurrentPriceList(ctx)
    if err != nil {
        return nil, err
    }

    if s.cache != nil {
        if payload, marshalErr := json.Marshal(items); marshalErr == nil {
            _ = s.cache.Set(ctx, cacheKey, string(payload), s.ttl)
        }
    }

    return items, nil
}
