package handler

import (
	"errors"
	"net/http"

	"clinic.local/backend/internal/apierror"
	"clinic.local/backend/internal/dto"
	"clinic.local/backend/internal/model"
	"clinic.local/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service service.Authenticator
}

func NewAuthHandler(service service.Authenticator) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request dto.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		apierror.Write(c, err)
		return
	}

	user, err := h.service.Login(c.Request.Context(), model.LoginCommand{
		Login:    request.Login,
		Password: request.Password,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			apierror.WriteMessage(c, http.StatusUnauthorized, "Неверный логин или пароль")
			return
		}

		apierror.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		ID:        user.ID,
		Login:     user.Login,
		FullName:  user.FullName,
		Email:     user.Email,
		Role:      user.Role,
		DoctorID:  user.DoctorID,
		PatientID: user.PatientID,
	})
}

func (h *AuthHandler) RegisterPatient(c *gin.Context) {
	var request dto.RegisterPatientAccountRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		apierror.Write(c, err)
		return
	}

	user, err := h.service.RegisterPatient(c.Request.Context(), model.RegisterPatientAccountCommand{
		LastName:     request.LastName,
		FirstName:    request.FirstName,
		MiddleName:   request.MiddleName,
		BirthDate:    request.BirthDate,
		Phone:        request.Phone,
		Email:        request.Email,
		PasswordHash: request.Password,
	})
	if err != nil {
		apierror.Write(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.LoginResponse{
		ID:        user.ID,
		Login:     user.Login,
		FullName:  user.FullName,
		Email:     user.Email,
		Role:      user.Role,
		PatientID: user.PatientID,
	})
}
