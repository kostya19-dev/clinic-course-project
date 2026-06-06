package config

import (
    "os"
    "time"
)

type Config struct {
    HTTPAddress       string
    DatabaseURL       string
    RedisAddr         string
    ScheduleCacheTTL  time.Duration
    PriceListCacheTTL time.Duration
}

func Load() Config {
    return Config{
        HTTPAddress:       getEnv("HTTP_ADDRESS", ":8080"),
        DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/clinic_db?sslmode=disable"),
        RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
        ScheduleCacheTTL:  getDurationEnv("SCHEDULE_CACHE_TTL", time.Minute),
        PriceListCacheTTL: getDurationEnv("PRICE_LIST_CACHE_TTL", 10*time.Minute),
    }
}

func getEnv(key, fallback string) string {
    value := os.Getenv(key)
    if value == "" {
        return fallback
    }

    return value
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
    value := os.Getenv(key)
    if value == "" {
        return fallback
    }

    parsed, err := time.ParseDuration(value)
    if err != nil {
        return fallback
    }

    return parsed
}
