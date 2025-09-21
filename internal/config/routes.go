package config

import (
	"net/http"

	"goQemu/internal/handler"

	"github.com/gin-gonic/gin"
)

func NewRoutes(r *gin.Engine, h *handler.Handler) {
	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		api.POST("/vm/install", h.Install)
		api.GET("/vm/:id/status", h.GetStatus)

		group := api.Group("/vm/:id")
		{
			group.POST("/start", h.Start)
			group.POST("/stop", h.Stop)
			group.POST("/shutdown", h.Shutdown)
			group.POST("/reboot", h.Reboot)
			group.POST("/destroy", h.Destroy)
		}
	}
}
