package web

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

type response struct {
	Message string `json:"message"`
}

func newRespones(c *gin.Context, status int, msg string) {
	slog.With("status", status).Error(msg)
	c.AbortWithStatusJSON(status, response{msg})
}
