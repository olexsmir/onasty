package apiv1

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/olexsmir/onasty/internal/models"
)

type response struct {
	Message string `json:"message"`
}

func errorResponse(c *gin.Context, err error) {
	if errors.Is(err, models.ErrUserEmailIsAlreadyInUse) ||
		errors.Is(err, models.ErrUsernameIsAlreadyInUse) ||
		errors.Is(err, models.ErrNoteContentIsEmpty) ||
		errors.Is(err, models.ErrNoteSlugIsAlreadyInUse) {
		newError(c, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, models.ErrNoteExpired) {
		newError(c, http.StatusGone, err.Error())
		return
	}

	if errors.Is(err, models.ErrNoteNotFound) {
		newErrorStatus(c, http.StatusNotFound, err.Error())
		return
	}

	if errors.Is(err, models.ErrUserNotFound) {
		newErrorStatus(c, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrUnauthorized) ||
		errors.Is(err, models.ErrUserWrongCredentials) {
		newErrorStatus(c, http.StatusUnauthorized, err.Error())
		return
	}

	newInternalError(c, err)
}

func newError(c *gin.Context, status int, msg string) { //nolint:unparam // TODO: remove me later
	slog.Error(msg, "status", status)
	c.AbortWithStatusJSON(status, response{msg})
}

func newErrorStatus(c *gin.Context, status int, msg string) {
	slog.Error(msg, "status", status)
	c.AbortWithStatus(status)
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
