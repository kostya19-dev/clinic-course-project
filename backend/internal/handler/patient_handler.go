package handler

import (
    "net/http"

    "clinic.local/backend/internal/apierror"
    "clinic.local/backend/internal/dto"
    "clinic.local/backend/internal/model"
    "clinic.local/backend/internal/service"
    "github.com/gin-gonic/gin"
)

type PatientHandler struct {
    service service.PatientCreator
}

func NewPatientHandler(service service.PatientCreator) *PatientHandler {
    return &PatientHandler{service: service}
}

func (h *PatientHandler) CreatePatient(c *gin.Context) {
    var request dto.CreatePatientRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        apierror.Write(c, err)
        return
    }

    patientID, err := h.service.Create(c.Request.Context(), model.CreatePatientCommand{
        LastName:   request.LastName,
        FirstName:  request.FirstName,
        MiddleName: request.MiddleName,
        BirthDate:  request.BirthDate,
        Phone:      request.Phone,
        Email:      request.Email,
        Status:     request.Status,
    })
    if err != nil {
        apierror.Write(c, err)
        return
    }

    c.JSON(http.StatusCreated, dto.IDResponse{ID: patientID})
}
