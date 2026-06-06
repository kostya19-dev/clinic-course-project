package handler

import (
    "net/http"

    "clinic.local/backend/internal/apierror"
    "clinic.local/backend/internal/service"
    "github.com/gin-gonic/gin"
)

type CatalogHandler struct {
    service service.CatalogReader
}

func NewCatalogHandler(service service.CatalogReader) *CatalogHandler {
    return &CatalogHandler{service: service}
}

func (h *CatalogHandler) GetPriceList(c *gin.Context) {
    items, err := h.service.ListCurrentPriceList(c.Request.Context())
    if err != nil {
        apierror.Write(c, err)
        return
    }

    c.JSON(http.StatusOK, items)
}
