package apierror

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
)

type response struct {
	Error string `json:"error"`
}

func Write(c *gin.Context, err error) {
	status, message := Resolve(err)
	c.JSON(status, response{Error: message})
}

func WriteMessage(c *gin.Context, status int, message string) {
	c.JSON(status, response{Error: message})
}

func Resolve(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}

	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		return http.StatusBadRequest, validationMessage(validationErrs)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return postgresMessage(pgErr)
	}

	return http.StatusInternalServerError, "Внутренняя ошибка сервера"
}

func validationMessage(errs validator.ValidationErrors) string {
	if len(errs) == 0 {
		return "Некорректные входные данные"
	}

	err := errs[0]

	switch err.Field() {
	case "LastName":
		return "Фамилия заполнена некорректно"
	case "FirstName":
		return "Имя заполнено некорректно"
	case "MiddleName":
		return "Отчество заполнено некорректно"
	case "BirthDate":
		return "Дата рождения обязательна и должна быть в формате YYYY-MM-DD"
	case "Phone":
		return "Телефон указан в неверном формате"
	case "Email":
		return "Email указан в неверном формате"
	case "Status":
		return "Передан недопустимый статус"
	case "PatientID":
		return "Некорректный идентификатор пациента"
	case "DoctorID":
		return "Некорректный идентификатор врача"
	case "ServiceID":
		return "Некорректный идентификатор услуги"
	case "CreatedByUserID":
		return "Некорректный идентификатор пользователя, создавшего запись"
	case "ActorUserID":
		return "Некорректный идентификатор пользователя"
	case "DoctorUserID":
		return "Некорректный идентификатор врача-пользователя"
	case "ReceivedByUserID":
		return "Некорректный идентификатор пользователя, принявшего оплату"
	case "PaymentMethod":
		return "Передан недопустимый способ оплаты"
	case "Amount":
		return "Сумма оплаты должна быть больше нуля"
	case "Login":
		return "Логин должен быть указан корректно"
	case "Password":
		return "Пароль должен содержать не менее 8 символов"
	case "Reason":
		return "Причина должна быть указана корректно"
	case "Note":
		return "Комментарий к записи заполнен некорректно"
	default:
		return "Некорректные входные данные"
	}
}

func postgresMessage(pgErr *pgconn.PgError) (int, string) {
	message := strings.ToLower(pgErr.Message)

	switch {
	case strings.Contains(message, "запись") && strings.Contains(message, "не найдена"):
		return http.StatusNotFound, "Запись не найдена"
	case strings.Contains(message, "not found"):
		return http.StatusNotFound, "Запись не найдена"
	case strings.Contains(message, "вне интервала доступности врача"):
		return http.StatusBadRequest, "Выбранное время находится вне расписания врача"
	case strings.Contains(message, "outside doctor availability"):
		return http.StatusBadRequest, "Выбранное время находится вне расписания врача"
	case strings.Contains(message, "врач отсутствует в выбранный интервал времени"):
		return http.StatusConflict, "Врач отсутствует в выбранный интервал времени"
	case strings.Contains(message, "doctor is absent"):
		return http.StatusConflict, "Врач отсутствует в выбранный интервал времени"
	case strings.Contains(message, "отменить можно только запись со статусом booked"):
		return http.StatusBadRequest, "Отменить можно только активную запись"
	case strings.Contains(message, "only booked appointments can be cancelled"):
		return http.StatusBadRequest, "Отменить можно только активную запись"
	case strings.Contains(message, "завершить можно только запись со статусом booked"):
		return http.StatusBadRequest, "Завершить можно только активную запись"
	case strings.Contains(message, "only booked appointments can be completed"):
		return http.StatusBadRequest, "Завершить можно только активную запись"
	case strings.Contains(message, "пациент может отменить запись не позднее чем за 24 часа до приема"):
		return http.StatusBadRequest, "Пациент может отменить запись не позднее чем за 24 часа до приема"
	case strings.Contains(message, "patient cancellation is allowed only 24 hours before the visit"):
		return http.StatusBadRequest, "Пациент может отменить запись не позднее чем за 24 часа до приема"
	case strings.Contains(message, "неактив"):
		return http.StatusBadRequest, "Связанная запись неактивна"
	case strings.Contains(message, "is not active"):
		return http.StatusBadRequest, "Связанная запись неактивна"
	case strings.Contains(message, "не закреплена за врачом"):
		return http.StatusBadRequest, "Услуга не закреплена за выбранным врачом"
	case strings.Contains(message, "not assigned to doctor"):
		return http.StatusBadRequest, "Услуга не закреплена за выбранным врачом"
	}

	switch pgErr.Code {
	case "23P01":
		return http.StatusConflict, "Конфликт записи по времени"
	case "23505":
		return http.StatusConflict, "Нарушено ограничение уникальности"
	case "P0002":
		return http.StatusNotFound, "Запись не найдена"
	case "23503":
		return http.StatusBadRequest, "Передана некорректная ссылка на связанную запись"
	case "23514", "23502", "22P02", "22023":
		return http.StatusBadRequest, "Переданы некорректные данные"
	default:
		return http.StatusBadRequest, pgErr.Message
	}
}
