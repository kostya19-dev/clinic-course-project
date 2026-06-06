SHELL := /bin/bash

PROJECT_ROOT := $(CURDIR)
APP_DIR ?= VitaLineClinic
DB_NAME ?= clinic_db
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_HOST ?= localhost
DB_PORT ?= 5432
REDIS_ADDR ?= localhost:6379
HTTP_ADDRESS ?= :8080

GO_BIN := $(shell command -v go 2>/dev/null)
ifeq ($(GO_BIN),)
GO_BIN := /snap/bin/go
endif

LATEX_BIN := $(shell command -v xelatex 2>/dev/null)
ifeq ($(LATEX_BIN),)
LATEX_BIN := $(shell command -v pdflatex 2>/dev/null)
endif

DATABASE_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
PSQL := PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -U $(DB_USER) -d $(DB_NAME)

.PHONY: help check-tools check-services db-create db-schema db-seed db-bulk-seed db-test db-prepare backend-deps backend-test backend-run frontend-deps frontend-build frontend-run app-build app-test app-docker-up app-docker-down app-docker-ps app-api-health app-gui-status report-build report-build-full all-checks full-checks

default: help

help:
	@echo "Доступные цели:"
	@echo "  make check-tools        - проверить наличие go, npm, psql, redis-cli, pdflatex, dot"
	@echo "  make check-services     - проверить доступность PostgreSQL и Redis"
	@echo "  make db-create          - создать базу $(DB_NAME), если ее еще нет"
	@echo "  make db-schema          - загрузить схему PostgreSQL"
	@echo "  make db-seed            - загрузить тестовые данные"
	@echo "  make db-bulk-seed       - загрузить массовые данные для исследований"
	@echo "  make db-test            - прогнать SQL-сценарии проверки"
	@echo "  make db-prepare         - создать БД, загрузить схему и seed-данные"
	@echo "  make backend-deps       - подтянуть зависимости backend"
	@echo "  make backend-test       - прогнать go test ./..."
	@echo "  make backend-run        - запустить backend на :8080"
	@echo "  make frontend-deps      - установить npm-зависимости frontend"
	@echo "  make frontend-build     - собрать frontend"
	@echo "  make frontend-run       - запустить Vite dev server на :5173"
	@echo "  make app-build          - собрать перенесенное приложение VitaLineClinic"
	@echo "  make app-test           - прогнать тесты перенесенного приложения"
	@echo "  make app-docker-up      - запустить PostgreSQL, Redis, backend и frontend"
	@echo "  make app-docker-down    - остановить контейнеры приложения"
	@echo "  make app-docker-ps      - показать контейнеры приложения"
	@echo "  make app-api-health     - проверить /api/v1/health"
	@echo "  make app-gui-status     - проверить frontend на :5173"
	@echo "  make report-build       - собрать РПЗ до конструкторской части в report/main_design.pdf"
	@echo "  make report-build-full  - собрать полную РПЗ в report/main.pdf"
	@echo "  make all-checks         - быстрая сквозная проверка backend, frontend и РПЗ"
	@echo "  make full-checks        - полная проверка, включая SQL-сценарии"

check-tools:
	@echo "go: $$($(GO_BIN) version)"
	@echo "npm: $$(npm --version)"
	@echo "node: $$(node --version)"
	@echo "psql: $$(psql --version)"
	@echo "redis-cli: $$(redis-cli --version)"
	@echo "latex: $$($(LATEX_BIN) --version | head -n 1)"
	@echo "dot: $$(dot -V 2>&1 | head -n 1)"

check-services:
	@pg_isready -h $(DB_HOST) -p $(DB_PORT)
	@redis-cli -h $(word 1,$(subst :, ,$(REDIS_ADDR))) -p $(word 2,$(subst :, ,$(REDIS_ADDR))) ping

db-create:
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -U $(DB_USER) -tAc "SELECT 1 FROM pg_database WHERE datname='$(DB_NAME)'" postgres | grep -q 1 || PGPASSWORD=$(DB_PASSWORD) createdb -h $(DB_HOST) -U $(DB_USER) $(DB_NAME)

db-schema:
	@$(PSQL) -f db/schema.sql

db-seed:
	@$(PSQL) -f db/seed.sql

db-bulk-seed:
	@$(PSQL) -f db/bulk_seed.sql

db-test:
	@$(PSQL) -f db/test_scenarios.sql

db-prepare: db-create db-schema db-seed

backend-deps:
	@cd backend && $(GO_BIN) mod tidy

backend-test:
	@cd backend && $(GO_BIN) test ./...

backend-run:
	@cd backend && DATABASE_URL='$(DATABASE_URL)' REDIS_ADDR='$(REDIS_ADDR)' HTTP_ADDRESS='$(HTTP_ADDRESS)' $(GO_BIN) run ./cmd/api

frontend-deps:
	@cd frontend && npm install

frontend-build:
	@cd frontend && npm run build

frontend-run:
	@cd frontend && npm run dev -- --host 0.0.0.0

app-build:
	@$(MAKE) -C $(APP_DIR) build

app-test:
	@$(MAKE) -C $(APP_DIR) test

app-docker-up:
	@docker compose up -d --build

app-docker-down:
	@docker compose down

app-docker-ps:
	@docker compose ps

app-api-health:
	@curl -s http://localhost:8080/api/v1/health
	@printf '\n'

app-gui-status:
	@curl -s -o /dev/null -w "Frontend status: %{http_code} http://localhost:5173\n" http://localhost:5173

report-build:
	@cd report && $(LATEX_BIN) -interaction=nonstopmode -halt-on-error main_design.tex >/tmp/clinic_report_build.log && $(LATEX_BIN) -interaction=nonstopmode -halt-on-error main_design.tex >>/tmp/clinic_report_build.log

# Полная сборка РПЗ оставлена отдельно, чтобы на этапе правки конструкторской части
# не компилировать технологический и исследовательский разделы.
# report-build:
# 	@cd report && pdflatex -interaction=nonstopmode -halt-on-error main.tex >/tmp/clinic_report_build.log && pdflatex -interaction=nonstopmode -halt-on-error main.tex >>/tmp/clinic_report_build.log

report-build-full:
	@cd report && $(LATEX_BIN) -interaction=nonstopmode -halt-on-error main.tex >/tmp/clinic_report_full_build.log && $(LATEX_BIN) -interaction=nonstopmode -halt-on-error main.tex >>/tmp/clinic_report_full_build.log

all-checks: check-services backend-test frontend-build report-build

full-checks: check-services db-test backend-test frontend-build report-build
