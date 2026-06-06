# Курсовой проект: медицинская клиника

Репозиторий приведен к требованиям курсового проекта по разработке информационной системы медицинской клиники.

Целевой стек по ТЗ:
- Frontend: React + TypeScript
- Паттерн UI: View -> ViewModel/Hook -> ApiClient
- Backend: Go + Gin
- Паттерн backend: Repository -> Service -> Handler
- СУБД: PostgreSQL
- Доступ к БД: `database/sql` + `pgx`
- Кэш: Redis
- Отчет: LaTeX

Структура репозитория:
- `backend/` — Go API
- `frontend/` — клиент на React + TypeScript
- `db/` — схема PostgreSQL, процедуры, триггеры и SQL-скрипты анализа
- `docs/` — анализ предметной области и принятые проектные решения
- `tests/` — сценарии функционального и нефункционального тестирования
- `report/` — расчетно-пояснительная записка в LaTeX

## Важно

- Docker в текущей версии проекта не используется.
- PostgreSQL и Redis запускаются как локальные службы внутри WSL.
- Старый каркас на `.NET` не используется для итоговой реализации по ТЗ.
- Основной путь запуска и проверки завязан на `Makefile` в корне проекта.

## Подробный запуск в WSL + VS Code

### 1. Что должно быть установлено в WSL

Если среда еще не подготовлена, в Ubuntu WSL должны быть доступны:
- `make`
- Go
- Node.js и npm
- PostgreSQL и `psql`
- Redis и `redis-cli`
- `pdflatex`
- пакет `texlive-lang-cyrillic`
- `graphviz`
- `librsvg2-bin`

Пример установки в Ubuntu:

```bash
sudo apt update
sudo apt install -y make postgresql postgresql-client redis-server nodejs npm texlive-latex-base texlive-latex-recommended texlive-latex-extra texlive-lang-cyrillic graphviz librsvg2-bin
sudo snap install go --classic
```

Если TeX Live уже установлен, этот шаг можно пропустить.

### 2. Как открыть проект в VS Code

1. Открыть VS Code.
2. Установить расширение `Remote - WSL`, если оно еще не установлено.
3. Выбрать `Open Folder in WSL`.
4. Открыть каталог проекта:

```text
/home/kosty/myhome/sem6/db_cursework/clinic-course-project
```

5. Все дальнейшие команды выполнять во встроенном терминале VS Code из корня проекта.

### 3. Проверить инструменты

```bash
make check-tools
```

Ожидаемо должны определиться:
- Go
- Node.js
- npm
- `psql`
- `redis-cli`
- `pdflatex`

### 4. Поднять PostgreSQL и Redis

Если службы еще не запущены, выполнить:

```bash
sudo service postgresql start
sudo service redis-server start
```

Проверка доступности:

```bash
make check-services
```

Ожидаемый результат:
- PostgreSQL отвечает на `localhost:5432`
- Redis отвечает `PONG`

### 5. Подготовить базу данных

По умолчанию используются такие параметры:
- база данных: `clinic_db`
- пользователь PostgreSQL: `postgres`
- пароль PostgreSQL: `postgres`

Создание базы данных:

```bash
make db-create
```

Загрузка схемы:

```bash
make db-schema
```

Загрузка тестовых данных:

```bash
make db-seed
```

Загрузка большого набора данных для исследований производительности:

```bash
make db-bulk-seed
```

Если нужен быстрый минимальный вариант одной командой:

```bash
make db-prepare
```

Если у вас другой пароль PostgreSQL, его можно передать прямо в команду:

```bash
make DB_PASSWORD=ваш_пароль db-prepare
```

### 6. Проверить SQL-сценарии

```bash
make db-test
```

Важно:
- в сценариях есть проверка конфликта записи на одно и то же время;
- поэтому в выводе PostgreSQL может появиться диагностическое сообщение о конфликте по ограничению `appointments_doctor_no_overlap`;
- это ожидаемая часть теста, а не признак поломки, если команда завершается успешно.

### 7. Подготовить backend

```bash
make backend-deps
make backend-test
```

Запуск backend:

```bash
make backend-run
```

Backend стартует на:
- `http://localhost:8080`

Минимальная ручная проверка:

```bash
curl http://localhost:8080/api/v1/health
```

Ожидаемый ответ:

```json
{"status":"ok"}
```

### 8. Подготовить frontend

Во втором терминале VS Code выполнить:

```bash
make frontend-deps
make frontend-build
make frontend-run
```

Frontend стартует на:
- `http://localhost:5173`

Важно:
- frontend работает в dev-режиме через Vite;
- запросы `/api` проксируются на backend `http://localhost:8080`;
- значит backend должен быть уже запущен.

### 9. Что можно проверить вручную в интерфейсе

Во frontend уже есть демонстрационные экраны:
- регистрация пациента;
- запись на прием;
- отмена записи;
- проведение приема;
- регистрация оплаты;
- просмотр расписания врача;
- просмотр прайс-листа.

Главная точка входа UI:
- `frontend/src/App.tsx`

### 10. Сборка РПЗ

```bash
make report-build
```

Готовый файл:
- `report/main.pdf`

### 11. Полная автоматизированная проверка

Быстрая сквозная проверка без SQL-сценариев:

```bash
make all-checks
```

Полная проверка, включая SQL-сценарии и сборку РПЗ:

```bash
make full-checks
```

### 12. Минимальный путь от нуля до рабочего состояния

Если нужен самый короткий сценарий запуска, порядок такой:

```bash
sudo service postgresql start
sudo service redis-server start
make db-prepare
make backend-test
make frontend-build
make report-build
```

Потом открыть два терминала.

Терминал 1:

```bash
make backend-run
```

Терминал 2:

```bash
make frontend-run
```

После этого открыть в браузере:
- `http://localhost:5173`

## Параметры по умолчанию

Если ничего не переопределять, проект стартует так:
- PostgreSQL: `postgres://postgres:postgres@localhost:5432/clinic_db?sslmode=disable`
- Redis: `localhost:6379`
- backend: `:8080`
- frontend: `:5173`

При необходимости параметры можно передать в `make`:

```bash
make DB_PASSWORD=ваш_пароль backend-run
make REDIS_ADDR=localhost:6379 backend-run
make HTTP_ADDRESS=:8080 backend-run
```
