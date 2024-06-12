package apiv1

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type response struct {
	Message string `json:"message"`
}

func newError(c *gin.Context, status int, msg string) {
	slog.With("status", status).Error(msg)
	c.AbortWithStatusJSON(status, response{msg})
}

func newInternalError(c *gin.Context, err error, msg ...string) {
	slog.With("status", "internal error").Error(err.Error())

	if len(msg) != 0 {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response{
			Message: msg[0],
		})
		return
	}

	c.AbortWithStatusJSON(http.StatusInternalServerError, response{
		Message: "internal error",
	})
}
