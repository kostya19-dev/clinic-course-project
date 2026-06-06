# Контрольная точка проекта

## Текущее состояние

Проект приведен к целевому стеку и ключевым ограничениям курсовой работы:
- backend: Go + Gin
- frontend: React + TypeScript
- база данных: PostgreSQL
- кэш: Redis
- архитектура backend: Repository -> Service -> Handler
- архитектура frontend: View -> ViewModel/Hook -> ApiClient
- отчет: LaTeX на русском языке

## Что уже сделано

### Требования и предметная область

- На русском языке оформлены требования, роли, бизнес-правила, API, кэширование и тестовые сценарии в `docs/requirements.md`.
- Учтены дополнительные ограничения:
  - UUID вместо целочисленных идентификаторов
  - enum-статусы вместо произвольных строк
  - типизированная дата рождения в backend
  - валидация email и телефона
  - русскоязычная проектная документация
  - русскоязычные сообщения API

### База данных

- В `db/schema.sql` описаны таблицы, связи, индексы, процедуры и триггеры.
- Все основные идентификаторы переведены на UUID.
- Для статусов и метода оплаты используются enum-типы PostgreSQL.
- Для email и телефона заданы домены PostgreSQL с проверками формата.
- Реализованы процедуры:
  - `scheduling.book_appointment`
  - `scheduling.cancel_appointment`
  - `medical.complete_visit`
  - `billing.register_payment`
- Для предотвращения конфликтов используются exclusion constraints.
- В WSL реально созданы и запущены PostgreSQL 16 и Redis 7.
- Схема `db/schema.sql`, seed-данные `db/seed.sql`, сценарий `db/bulk_seed.sql` и сценарии `db/test_scenarios.sql` успешно прогнаны на локальном PostgreSQL.

### SQL-артефакты

- `db/seed.sql` содержит детерминированные тестовые данные с фиксированными UUID.
- `db/bulk_seed.sql` создает детерминированный массовый набор данных для исследований производительности и выполнения требований по объему данных.
- `db/test_scenarios.sql` содержит исполнимые сценарии проверки:
  - валидации пациента
  - записи на прием
  - отмены записи
  - завершения приема
  - частичной и полной оплаты
- `db/concurrency_session_1.sql` и `db/concurrency_session_2.sql` подготовлены для ручной проверки конкурентной записи.
- `db/concurrency_cleanup.sql` очищает данные после сценария конкурентности.
- `db/performance.sql` содержит шаблоны запросов для `EXPLAIN ANALYZE`.

### Backend

- В `backend/` создан каркас приложения на Go.
- Реализованы слои DTO, model, repository, service и handler.
- В `backend/internal/service/contracts.go` выделены интерфейсы сервисного слоя для HTTP-уровня.
- `backend/internal/server/router.go` и handlers переведены на зависимости от интерфейсов, а не от конкретных реализаций.
- Подключены кастомные валидаторы в `backend/internal/validation/validation.go`.
- Добавлен единый слой обработки ошибок PostgreSQL и ошибок валидации с русскими сообщениями.
- Проведение приема возвращает в API идентификаторы приема и сформированного счета, что позволяет связать сценарий оплаты с фактическим результатом работы backend.
- Добавлены unit-тесты для:
  - слоя валидации
  - маппинга ошибок PostgreSQL
  - HTTP-слоя роутера и handlers через `httptest`
  - кэширования прайс-листа и расписания
- В WSL выполнены `go mod tidy` и `go test ./...`; backend-тесты проходят успешно.
- Проверен запуск backend на реальных PostgreSQL и Redis; подтверждены ответы endpoint-ов `GET /api/v1/health` и `GET /api/v1/services/pricelist`.
- Выполнен реальный API-сценарий записи на прием, конфликта по времени и отмены записи со статусами 201, 409 и 204 соответственно.
- Выполнен реальный API-сценарий проведения приема, формирования счета и регистрации оплаты со статусами 201, 201, 201 и 201 соответственно.
- Логика Redis-кэширования прайс-листа и расписания присутствует и подтверждена как unit-тестами, так и реальной проверкой Redis-ключей при работе backend.

### Frontend

- В `frontend/` реализован демонстрационный интерфейс на React + TypeScript.
- Интерфейс приведен к русскоязычному виду и не выходит за рамки ТЗ.
- Реализованы демонстрационные экраны:
  - регистрация пациента
  - запись на прием
  - отмена записи
  - проведение приема
  - регистрация оплаты
  - просмотр расписания врача
  - просмотр прайс-листа
- Для фронтенда используются реальные детерминированные UUID из `db/seed.sql`.
- В WSL установлены Node.js 18.19.1 и npm 9.2.0.
- В каталоге `frontend/` успешно выполнены `npm install` и `npm run build`.
- Подтвержден запуск Vite dev server: `http://127.0.0.1:5173` отвечает со статусом `HTTP 200`.
- В `frontend/vite.config.ts` настроено проксирование `/api` на backend `http://localhost:8080`.

### Исследования

- Выполнен реальный сценарий конкурентной записи: первая транзакция успешно зафиксировалась, вторая завершилась ошибкой exclusion constraint по пересечению интервала.
- Подготовлен и загружен массовый набор данных с более чем 1000 строками в основных таблицах.
- Зафиксированы реальные результаты исследования производительности в `docs/performance.md`.
- По замерам PostgreSQL индексы дают выигрыш на селективных запросах расписания, истории пациента и списка счетов, тогда как широкий отчет по выручке за период обоснованно использует последовательное сканирование.

### Документация и эксплуатация

- Подготовлены `docs/concurrency.md`, `docs/performance.md`, `docs/runbook.md`, `docs/tz_coverage.md` и `docs/checkpoint.md`.
- В `docs/runbook.md` описан порядок запуска схемы, seed-данных, bulk-данных, SQL-сценариев, конкурентной проверки, тестов backend, запуска backend и frontend.

### Отчет

- В каталоге `report/` создан LaTeX-каркас расчетно-пояснительной записки.
- Подготовлены разделы: реферат, введение, аналитический, конструкторский, технологический, исследовательский разделы, заключение, список источников и приложение.
- В WSL установлен пакет `texlive-lang-cyrillic` для сборки отчета на русском языке.
- Отчет успешно собирается командой `pdflatex`; результат находится в `report/main.pdf`.
- В отчет перенесены реальные результаты API-проверок, Redis-кэширования, конкурентного доступа и исследования производительности.
- В отчет встроены пользовательские use case и ER-диаграммы, а также укрупненные схемы базы данных и архитектуры приложения.
- В приложении к отчету зафиксированы демонстрационные экраны frontend и перечень материалов для окончательной сдачи.

## Ограничения текущей среды

На данный момент основная среда практически готова, но остаются завершающие задачи:
- перечень нужных скриншотов уже определен, но сами изображения для окончательной версии РПЗ пока не подготовлены;
- титульные персональные данные и финальная редактура отчета еще требуют доводки;
- демонстрационная часть закрыта по функциональности и теперь требует в основном только финальной упаковки.

## Следующие шаги

1. Подготовить скриншоты основных экранов frontend и вынести их в отчет.
2. Доработать LaTeX-отчет: заполнить титульные персональные данные и выполнить финальную редактуру текста.
3. После правок снова пересобрать `report/main.pdf`.

## Как продолжать позже

При следующем заходе начинать с этих файлов:
- `docs/checkpoint.md`
- `docs/requirements.md`
- `docs/concurrency.md`
- `docs/performance.md`
- `docs/runbook.md`
- `docs/tz_coverage.md`
- `db/schema.sql`
- `db/seed.sql`
- `db/bulk_seed.sql`
- `db/test_scenarios.sql`
- `db/concurrency_session_1.sql`
- `db/concurrency_session_2.sql`
- `backend/internal/dto/dto.go`
- `backend/internal/handler/appointment_handler.go`
- `backend/internal/service/contracts.go`
- `backend/internal/server/router.go`
- `backend/internal/server/router_test.go`
- `frontend/src/App.tsx`
- `frontend/src/views/PatientRegistrationPage.tsx`
- `frontend/src/views/AppointmentWorkflowPage.tsx`
- `frontend/src/views/CompleteVisitPage.tsx`
- `frontend/src/views/PaymentRegistrationPage.tsx`
- `frontend/src/views/DoctorSchedulePage.tsx`
- `frontend/src/views/PriceListPage.tsx`
- `report/main.tex`
- `report/metadata.tex`
- `report/sections/analysis.tex`
- `report/sections/design.tex`
- `report/sections/technology.tex`
- `report/sections/research.tex`
- `report/sections/appendix_a.tex`
- `report/figures/use-case.pdf`
- `report/figures/er-diagram.pdf`
