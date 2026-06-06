package handler

import (
    "net/http"

    "clinic.local/backend/internal/apierror"
    "clinic.local/backend/internal/dto"
    "clinic.local/backend/internal/model"
    "clinic.local/backend/internal/service"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

type AppointmentHandler struct {
    service service.AppointmentManager
}

func NewAppointmentHandler(service service.AppointmentManager) *AppointmentHandler {
    return &AppointmentHandler{service: service}
}

func (h *AppointmentHandler) BookAppointment(c *gin.Context) {
    var request dto.BookAppointmentRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        apierror.Write(c, err)
        return
    }

    appointmentID, err := h.service.Book(c.Request.Context(), model.BookAppointmentCommand{
        PatientID:       request.PatientID,
        DoctorID:        request.DoctorID,
        ServiceID:       request.ServiceID,
        StartsAt:        request.StartsAt,
        CreatedByUserID: request.CreatedByUserID,
        Note:            request.Note,
    })
    if err != nil {
        apierror.Write(c, err)
        return
    }

    c.JSON(http.StatusCreated, dto.IDResponse{ID: appointmentID})
}

func (h *AppointmentHandler) CancelAppointment(c *gin.Context) {
    appointmentID, err := uuid.Parse(c.Param("appointmentId"))
    if err != nil {
        apierror.WriteMessage(c, http.StatusBadRequest, "Некорректный идентификатор записи")
        return
    }

    var request dto.CancelAppointmentRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        apierror.Write(c, err)
        return
    }

    err = h.service.Cancel(c.Request.Context(), model.CancelAppointmentCommand{
        AppointmentID: appointmentID,
        ActorUserID:   request.ActorUserID,
        Reason:        request.Reason,
    })
    if err != nil {
        apierror.Write(c, err)
        return
    }

    c.Status(http.StatusNoContent)
}

func (h *AppointmentHandler) CompleteVisit(c *gin.Context) {
    appointmentID, err := uuid.Parse(c.Param("appointmentId"))
    if err != nil {
        apierror.WriteMessage(c, http.StatusBadRequest, "Некорректный идентификатор записи")
        return
    }

    var request dto.CompleteVisitRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        apierror.Write(c, err)
        return
    }

    result, err := h.service.Complete(c.Request.Context(), model.CompleteVisitCommand{
        AppointmentID: appointmentID,
        DoctorUserID:  request.DoctorUserID,
        Summary:       request.Summary,
        Diagnosis:     request.Diagnosis,
        Services:      request.Services,
    })
    if err != nil {
        apierror.Write(c, err)
        return
    }

    c.JSON(http.StatusCreated, dto.CompleteVisitResponse{
        VisitID:   result.VisitID,
        InvoiceID: result.InvoiceID,
    })
}

func (h *AppointmentHandler) RegisterPayment(c *gin.Context) {
    invoiceID, err := uuid.Parse(c.Param("invoiceId"))
    if err != nil {
        apierror.WriteMessage(c, http.StatusBadRequest, "Некорректный идентификатор счета")
        return
    }

    var request dto.RegisterPaymentRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        apierror.Write(c, err)
        return
    }

    paymentID, err := h.service.RegisterPayment(c.Request.Context(), model.RegisterPaymentCommand{
        InvoiceID:         invoiceID,
        ReceivedByUserID:  request.ReceivedByUserID,
        Amount:            request.Amount,
        PaymentMethod:     request.PaymentMethod,
        ExternalReference: request.ExternalReference,
    })
    if err != nil {
        apierror.Write(c, err)
        return
    }

    c.JSON(http.StatusCreated, dto.IDResponse{ID: paymentID})
}