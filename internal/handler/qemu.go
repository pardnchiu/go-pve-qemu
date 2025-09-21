package handler

import (
	"fmt"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"

	"goQemu/internal/model"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Stop(c *gin.Context) {
	if !h.checkIP(c) {
		c.String(http.StatusForbidden, "this IP is not allowed to perform this action")
		return
	}

	isAvailable, vmid := h.checkAvailable(c)
	if !isAvailable {
		c.String(http.StatusInternalServerError, "failed to stop VM\n")
		return
	}

	if err := h.service.Stop(vmid); err != nil {
		c.String(http.StatusInternalServerError, "failed to stop VM: %v", err)
		return
	}

	c.String(http.StatusOK, "ok")
}

func (h *Handler) Shutdown(c *gin.Context) {
	if !h.checkIP(c) {
		c.String(http.StatusForbidden, "this IP is not allowed to perform this action")
		return
	}

	isAvailable, vmid := h.checkAvailable(c)
	if !isAvailable {
		c.String(http.StatusInternalServerError, "failed to shutdown VM\n")
		return
	}

	if err := h.service.Shutdown(vmid); err != nil {
		c.String(http.StatusInternalServerError, "failed to shutdown VM: %v", err)
		return
	}

	c.String(http.StatusOK, "ok")
}

func (h *Handler) Destroy(c *gin.Context) {
	if !h.checkIP(c) {
		c.String(http.StatusForbidden, "this IP is not allowed to perform this action")
		return
	}

	isAvailable, vmid := h.checkAvailable(c)
	if !isAvailable {
		c.String(http.StatusInternalServerError, "failed to destroy VM\n")
		return
	}

	if err := h.service.Destroy(vmid); err != nil {
		c.String(http.StatusInternalServerError, "failed to destroy VM: %v", err)
		return
	}

	c.String(http.StatusOK, "ok")
}

func (h *Handler) checkIP(c *gin.Context) bool {
	// * disable if not in ALLOW_IPS
	clientIP := c.ClientIP()
	allowIPs := os.Getenv("ALLOW_IPS")
	if allowIPs != "0.0.0.0" && !slices.Contains(strings.Split(allowIPs, ","), clientIP) {
		return false
	}
	return true
}

func (h *Handler) checkAvailable(c *gin.Context) (bool, int) {
	vmid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "VMID is invalid")
		return false, 0
	}

	return true, vmid
}

// * SSE
func (h *Handler) Install(c *gin.Context) {
	if !h.checkIP(c) {
		c.String(http.StatusForbidden, "this IP is not allowed to perform this action")
		return
	}

	var config model.Config

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Message: fmt.Sprintf("please check your input: %v", err),
		})
		return
	}

	h.service.Install(&config, c)

	c.Writer.WriteString("event: close\ndata: {}\n\n")
	c.Writer.Flush()
}

func (h *Handler) Start(c *gin.Context) {
	if !h.checkIP(c) {
		c.String(http.StatusForbidden, "this IP is not allowed to perform this action")
		return
	}

	isAvailable, vmid := h.checkAvailable(c)
	if !isAvailable {
		c.String(http.StatusInternalServerError, "failed to start VM\n")
		return
	}

	err := h.service.Start(c, vmid)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to start VM: %v", err)
		return
	}

	c.Writer.WriteString("event: close\ndata: {}\n\n")
	c.Writer.Flush()
}

func (h *Handler) Reboot(c *gin.Context) {
	if !h.checkIP(c) {
		c.String(http.StatusForbidden, "this IP is not allowed to perform this action")
		return
	}

	isAvailable, vmid := h.checkAvailable(c)
	if !isAvailable {
		c.String(http.StatusInternalServerError, "failed to reboot VM\n")
		return
	}

	err := h.service.Reboot(c, vmid)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to reboot VM: %v", err)
		return
	}

	c.Writer.WriteString("event: close\ndata: {}\n\n")
	c.Writer.Flush()
}
