package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RegisterRoutes вешает все маршруты на переданный Engine.
func RegisterRoutes(engine *gin.Engine, userHandler *UserHandler) {
	api := engine.Group("/api")
	api.GET("/users", userHandler.List)
	api.POST("/users", userHandler.Create)
	api.GET("/users/:id", userHandler.Get)
	api.PATCH("/users/:id", userHandler.Update)
	api.DELETE("/users/:id", userHandler.Delete)

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
