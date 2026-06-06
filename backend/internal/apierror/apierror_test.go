package apierror

import (
    "net/http"
    "testing"

    "github.com/go-playground/validator/v10"
    "github.com/jackc/pgx/v5/pgconn"
)

func TestResolve_ValidationError(t *testing.T) {
    t.Helper()

    validate := validator.New()

    type payload struct {
        Email string `validate:"email"`
    }

    err := validate.Struct(payload{Email: "not-an-email"})
    status, message := Resolve(err)

    if status != http.StatusBadRequest {
        t.Fatalf("ожидался статус 400, получен %d", status)
    }

    if message != "Email указан в неверном формате" {
        t.Fatalf("ожидалось сообщение о некорректном email, получено %q", message)
    }
}

func TestResolve_PostgresConflictByCode(t *testing.T) {
    t.Helper()

    status, message := Resolve(&pgconn.PgError{Code: "23P01", Message: "conflict"})

    if status != http.StatusConflict {
        t.Fatalf("ожидался статус 409, получен %d", status)
    }

    if message != "Конфликт записи по времени" {
        t.Fatalf("ожидалось сообщение о конфликте записи, получено %q", message)
    }
}

func TestResolve_PostgresNotFoundByCode(t *testing.T) {
    t.Helper()

    status, message := Resolve(&pgconn.PgError{Code: "P0002", Message: "Запись не найдена"})

    if status != http.StatusNotFound {
        t.Fatalf("ожидался статус 404, получен %d", status)
    }

    if message != "Запись не найдена" {
        t.Fatalf("ожидалось сообщение о ненайденной записи, получено %q", message)
    }
}

func TestResolve_PostgresBusinessMessage(t *testing.T) {
    t.Helper()

    status, message := Resolve(&pgconn.PgError{Code: "22023", Message: "Врач отсутствует в выбранный интервал времени"})

    if status != http.StatusConflict {
        t.Fatalf("ожидался статус 409, получен %d", status)
    }

    if message != "Врач отсутствует в выбранный интервал времени" {
        t.Fatalf("ожидалось сообщение об отсутствии врача, получено %q", message)
    }
}

func TestResolve_PostgresDefaultBadRequest(t *testing.T) {
    t.Helper()

    status, message := Resolve(&pgconn.PgError{Code: "22023", Message: "Произвольная ошибка бизнес-правила"})

    if status != http.StatusBadRequest {
        t.Fatalf("ожидался статус 400, получен %d", status)
    }

    if message != "Переданы некорректные данные" {
        t.Fatalf("ожидалось общее сообщение о некорректных данных, получено %q", message)
    }
}
