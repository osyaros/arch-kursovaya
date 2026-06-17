package handler

import (
	"errors"
	"net/http"

	"arch-oyu-lab3/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// writeJSONError — единый формат ошибки для API, чтобы хендлеры не дублировали gin.H{"error": ...}.
func writeJSONError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

// uuidFromURL берёт :id из пути и парсит в UUID. При ошибке сама отвечает 400.
func uuidFromURL(c *gin.Context) (uuid.UUID, bool) {
	raw := c.Param("id")
	id, err := uuid.Parse(raw)
	if err != nil {
		writeJSONError(c, http.StatusBadRequest, "некорректный id (нужен UUID)")
		return uuid.UUID{}, false
	}
	return id, true
}

// writeErrorFromService мапит ошибки репозитория в HTTP-коды.
func writeErrorFromService(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, repository.ErrNotFound) {
		writeJSONError(c, http.StatusNotFound, "пользователь не найден")
		return true
	}
	writeJSONError(c, http.StatusInternalServerError, err.Error())
	return true
}
