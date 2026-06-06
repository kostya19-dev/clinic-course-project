# Порядок запуска и проверки проекта

## Соответствие ТЗ

Проект запускается строго в рамках утвержденного стека курсовой работы:
- frontend: React + TypeScript;
- backend: Go + Gin;
- база данных: PostgreSQL;
- кэш: Redis;
- отчет: LaTeX.

Docker в текущей версии проекта не используется, потому что по ТЗ он не требуется. PostgreSQL и Redis запускаются как локальные службы в WSL.

## Что должно быть установлено в WSL

Перед началом работы в Ubuntu WSL должны быть доступны:
- `make`;
- Go;
- Node.js и npm;
- PostgreSQL;
- Redis;
- `psql`;
- `redis-cli`;
- `pdflatex`;
- пакет `texlive-lang-cyrillic`;
- `graphviz`;
- `librsvg2-bin`.

Если чего-то нет, пример установки такой:

```bash
sudo apt update
sudo apt install -y make postgresql postgresql-client redis-server nodejs npm texlive-latex-base texlive-latex-recommended texlive-latex-extra texlive-lang-cyrillic graphviz librsvg2-bin
sudo snap install go --classic
```

Проверить наличие инструментов можно командой:

```bash
make check-tools
```

## Как открыть проект в VS Code

1. Открыть VS Code.
2. Выбрать `Open Folder in WSL`.
3. Открыть каталог проекта:

```text
/home/kosty/myhome/sem6/db_cursework/clinic-course-project
```

4. Все команды ниже выполнять из корня проекта.

## Запуск PostgreSQL и Redis

Если службы еще не подняты:

```bash
sudo service postgresql start
sudo service redis-server start
```

Проверить их можно так:

```bash
make check-services
```

Ожидаемый результат:
- PostgreSQL доступен на `localhost:5432`;
- Redis отвечает `PONG`.

## Подготовка базы данных

По умолчанию используются параметры:
- база: `clinic_db`;
- пользователь PostgreSQL: `postgres`;
- пароль PostgreSQL: `postgres`.

### Создать базу данных

```bash
make db-create
```

### Загрузить схему

```bash
make db-schema
```

### Загрузить тестовые данные

```bash
make db-seed
```

### Загрузить массовый набор данных для исследований

```bash
make db-bulk-seed
```

### Быстрый минимальный вариант

Если нужна сразу рабочая БД с тестовыми данными:

```bash
make db-prepare
```

Если у PostgreSQL другой пароль, его можно передать прямо в команду:

```bash
make DB_PASSWORD=ваш_пароль db-prepare
```

## Проверка SQL-сценариев

```bash
make db-test
```

Важно:
- в сценарии есть ожидаемая проверка конфликта по времени;
- поэтому PostgreSQL может вывести сообщение о нарушении ограничения `appointments_doctor_no_overlap`;
- если команда завершается успешно, значит проверка прошла штатно.

## Проверка конкурентной записи

1. Убедиться, что схема и seed-данные загружены.
2. Открыть две SQL-сессии.
3. В первой выполнить:

```bash
PGPASSWORD=postgres psql -h localhost -U postgres -d clinic_db -f db/concurrency_session_1.sql
```

4. Пока первая сессия находится внутри `pg_sleep(10)`, во второй выполнить:

```bash
PGPASSWORD=postgres psql -h localhost -U postgres -d clinic_db -f db/concurrency_session_2.sql
```

5. Ожидаемо:
- первая транзакция фиксируется успешно;
- вторая завершается ошибкой конфликта по времени.

6. После проверки выполнить очистку:

```bash
PGPASSWORD=postgres psql -h localhost -U postgres -d clinic_db -f db/concurrency_cleanup.sql
```

## Backend

### Подготовка backend

```bash
make backend-deps
make backend-test
```

### Запуск backend

```bash
make backend-run
```

Параметры по умолчанию:
- PostgreSQL: `postgres://postgres:postgres@localhost:5432/clinic_db?sslmode=disable`
- Redis: `localhost:6379`
- HTTP: `:8080`

Если нужен другой пароль PostgreSQL:

```bash
make DB_PASSWORD=ваш_пароль backend-run
```

### Что проверить вручную

```bash
curl http://localhost:8080/api/v1/health
curl http://localhost:8080/api/v1/services/pricelist
```

Также доступны ключевые endpoint-ы:
- `POST /api/v1/patients`
- `POST /api/v1/appointments`
- `POST /api/v1/appointments/:appointmentId/cancel`
- `POST /api/v1/appointments/:appointmentId/complete`
- `POST /api/v1/invoices/:invoiceId/payments`

## Frontend

### Подготовка frontend

```bash
make frontend-deps
make frontend-build
```

### Запуск frontend

Открыть второй терминал VS Code и выполнить:

```bash
make frontend-run
```

Frontend будет доступен по адресу:
- `http://localhost:5173`

Важно:
- frontend работает через Vite;
- запросы `/api` проксируются на backend `http://localhost:8080`;
- backend должен быть уже запущен.

## Что проверять во frontend

Во frontend уже реализованы следующие экраны:
- регистрация пациента;
- запись на прием;
- отмена записи;
- проведение приема;
- регистрация оплаты;
- расписание врача;
- прайс-лист услуг.

## Сборка РПЗ

Для сборки расчетно-пояснительной записки выполнить:

```bash
make report-build
```

Готовый PDF:

```text
report/main.pdf
```

## Сквозная локальная проверка

Если нужно быстро пройти минимальный путь от среды до результата:

```bash
sudo service postgresql start
sudo service redis-server start
make db-prepare
make full-checks
```

После этого:

Терминал 1:

```bash
make backend-run
```

Терминал 2:

```bash
make frontend-run
```

Потом открыть в браузере:
- `http://localhost:5173`

## Если контекст сессии закончится

Перед возобновлением работы открыть в таком порядке:
1. `docs/checkpoint.md`
2. `docs/requirements.md`
3. `docs/concurrency.md`
4. `docs/performance.md`
5. `docs/runbook.md`
6. `db/schema.sql`
7. `db/seed.sql`
8. `db/bulk_seed.sql`
9. `db/test_scenarios.sql`
10. `backend/internal/service/contracts.go`
11. `backend/internal/server/router.go`
12. `backend/internal/server/router_test.go`
13. `frontend/src/App.tsx`
14. `frontend/src/views/PatientRegistrationPage.tsx`
15. `frontend/src/views/AppointmentWorkflowPage.tsx`
16. `frontend/src/views/DoctorSchedulePage.tsx`
17. `report/main.tex`
18. `report/sections/technology.tex`
19. `report/sections/research.tex`
