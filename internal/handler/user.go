package handler

import (
	"net/http"
	"time"

	"arch-oyu-lab3/internal/metrics"
	"arch-oyu-lab3/internal/models"
	"arch-oyu-lab3/internal/service"

	"github.com/gin-gonic/gin"
)

// UserHandler — тонкий слой: разбор HTTP → вызов сервиса → код ответа.
type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) List(c *gin.Context) {
	users, err := h.service.List(c.Request.Context())
	if err != nil {
		writeJSONError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) Get(c *gin.Context) {
	id, ok := uuidFromURL(c)
	if !ok {
		return
	}

	start := time.Now()
	user, err := h.service.GetByID(c.Request.Context(), id)
	metrics.APIUserRequest.WithLabelValues(metrics.GetTagVal).Observe(time.Since(start).Seconds())

	if writeErrorFromService(c, err) {
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Create(c *gin.Context) {
	var body models.UserCreate
	if err := c.ShouldBindJSON(&body); err != nil {
		writeJSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	start := time.Now()
	user, err := h.service.Create(c.Request.Context(), body)
	metrics.APIUserRequest.WithLabelValues(metrics.PostTagVal).Observe(time.Since(start).Seconds())

	if err != nil {
		writeJSONError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) Update(c *gin.Context) {
	id, ok := uuidFromURL(c)
	if !ok {
		return
	}

	var body models.UserUpdate
	if err := c.ShouldBindJSON(&body); err != nil {
		writeJSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.service.Update(c.Request.Context(), id, body)
	if writeErrorFromService(c, err) {
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, ok := uuidFromURL(c)
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); writeErrorFromService(c, err) {
		return
	}
	c.Status(http.StatusNoContent)
}
