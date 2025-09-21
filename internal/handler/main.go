package handler

import (
	"net/http"
	"strconv"

	"guthub.com/pardnchiu/go-qemu/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetStatus(c *gin.Context) {
	vmid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid VM ID"})
		return
	}

	status, err := h.service.GetVMStatus(vmid)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to get VM status: %v", err)
		return
	}

	c.String(http.StatusOK, status)
}
