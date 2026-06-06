package handler

import "github.com/gin-gonic/gin"

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
    return &HealthHandler{}
}

func (h *HealthHandler) GetHealth(c *gin.Context) {
    c.JSON(200, gin.H{"status": "ok"})
}
