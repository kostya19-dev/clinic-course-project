package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	cachepkg "clinic.local/backend/internal/cache"
	"clinic.local/backend/internal/config"
	"clinic.local/backend/internal/repository/postgres"
	"clinic.local/backend/internal/server"
	"clinic.local/backend/internal/service"
	"clinic.local/backend/internal/validation"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
)

func main() {
	if err := validation.Register(); err != nil {
		log.Fatalf("register validators: %v", err)
	}

	cfg := config.Load()

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer redisClient.Close()

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("ping redis: %v", err)
	}

	cache := cachepkg.NewRedisCache(redisClient)

	patientRepo := postgres.NewPatientRepository(db)
	catalogRepo := postgres.NewCatalogRepository(db)
	scheduleRepo := postgres.NewScheduleRepository(db)
	appointmentRepo := postgres.NewAppointmentRepository(db)
	authRepo := postgres.NewAuthRepository(db)

	patientService := service.NewPatientService(patientRepo)
	catalogService := service.NewCatalogService(catalogRepo, cache, cfg.PriceListCacheTTL)
	scheduleService := service.NewScheduleService(scheduleRepo, cache, cfg.ScheduleCacheTTL)
	appointmentService := service.NewAppointmentService(appointmentRepo, cache)
	authService := service.NewAuthService(authRepo)

	router := server.NewRouter(patientService, catalogService, scheduleService, appointmentService, authService)

	if err := router.Run(cfg.HTTPAddress); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
