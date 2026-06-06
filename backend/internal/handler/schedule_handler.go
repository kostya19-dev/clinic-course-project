package handler

import (
    "net/http"
    "time"

    "clinic.local/backend/internal/apierror"
    "clinic.local/backend/internal/service"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

type ScheduleHandler struct {
    service service.ScheduleReader
}

func NewScheduleHandler(service service.ScheduleReader) *ScheduleHandler {
    return &ScheduleHandler{service: service}
}

func (h *ScheduleHandler) GetDoctorSchedule(c *gin.Context) {
    doctorID, err := uuid.Parse(c.Param("doctorId"))
    if err != nil {
        apierror.WriteMessage(c, http.StatusBadRequest, "Некорректный идентификатор врача")
        return
    }

    from, err := time.Parse(time.RFC3339, c.Query("from"))
    if err != nil {
        apierror.WriteMessage(c, http.StatusBadRequest, "Параметр from должен быть в формате RFC3339")
        return
    }

    to, err := time.Parse(time.RFC3339, c.Query("to"))
    if err != nil {
        apierror.WriteMessage(c, http.StatusBadRequest, "Параметр to должен быть в формате RFC3339")
        return
    }

    schedule, err := h.service.GetDoctorSchedule(c.Request.Context(), doctorID, from, to)
    if err != nil {
        apierror.Write(c, err)
        return
    }

    c.JSON(http.StatusOK, schedule)
}
