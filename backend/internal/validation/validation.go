package validation

import (
    "regexp"

    "clinic.local/backend/internal/model"
    "github.com/gin-gonic/gin/binding"
    "github.com/go-playground/validator/v10"
    "github.com/google/uuid"
)

var phonePattern = regexp.MustCompile(`^\+?[1-9]\d{10,14}$`)

func Register() error {
    engine, ok := binding.Validator.Engine().(*validator.Validate)
    if !ok {
        return nil
    }

    if err := engine.RegisterValidation("not_nil_uuid", validateUUID); err != nil {
        return err
    }

    if err := engine.RegisterValidation("phone", validatePhone); err != nil {
        return err
    }

    if err := engine.RegisterValidation("date_required", validateDate); err != nil {
        return err
    }

    return nil
}

func validateUUID(fl validator.FieldLevel) bool {
    value, ok := fl.Field().Interface().(uuid.UUID)
    if !ok {
        return false
    }

    return value != uuid.Nil
}

func validatePhone(fl validator.FieldLevel) bool {
    return phonePattern.MatchString(fl.Field().String())
}

func validateDate(fl validator.FieldLevel) bool {
    value, ok := fl.Field().Interface().(model.Date)
    if !ok {
        return false
    }

    return !value.IsZero()
}
