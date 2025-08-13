package apiv1

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/olexsmir/onasty/internal/jwtutil"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/service/notesrv"
	"github.com/olexsmir/onasty/internal/service/usersrv"
)

var ErrUnauthorized = errors.New("unauthorized")

type response struct {
	Message string `json:"message"`
}

func errorResponse(c *gin.Context, err error) {
	if errors.Is(err, usersrv.ErrProviderNotSupported) ||
		errors.Is(err, models.ErrResetPasswordTokenAlreadyUsed) ||
		errors.Is(err, models.ErrResetPasswordTokenExpired) ||
		errors.Is(err, models.ErrUserEmailIsAlreadyInUse) ||
		errors.Is(err, models.ErrUserIsAlreadyVerified) ||
		errors.Is(err, models.ErrUserIsNotActivated) ||
		errors.Is(err, models.ErrUserInvalidEmail) ||
		errors.Is(err, models.ErrUserInvalidPassword) ||
		errors.Is(err, models.ErrUserNotFound) ||
		errors.Is(err, models.ErrUserWrongCredentials) ||
		// notes
		errors.Is(err, notesrv.ErrNotePasswordNotProvided) ||
		errors.Is(err, models.ErrNoteContentIsEmpty) ||
		errors.Is(err, models.ErrNoteSlugIsAlreadyInUse) ||
		errors.Is(err, models.ErrNoteSlugIsInvalid) {
		newError(c, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, models.ErrNoteExpired) {
		newError(c, http.StatusGone, err.Error())
		return
	}

	if errors.Is(err, models.ErrNoteNotFound) ||
		errors.Is(err, models.ErrVerificationTokenNotFound) {
		newErrorStatus(c, http.StatusNotFound, err.Error())
		return
	}

	if errors.Is(err, ErrUnauthorized) ||
		errors.Is(err, jwtutil.ErrTokenExpired) ||
		errors.Is(err, jwtutil.ErrTokenSignatureInvalid) {
		newErrorStatus(c, http.StatusUnauthorized, err.Error())
		return
	}

	newInternalError(c, err)
}

func newError(c *gin.Context, status int, msg string) {
	slog.ErrorContext(c.Request.Context(), msg, "status", status)
	c.AbortWithStatusJSON(status, response{msg})
}

func newErrorStatus(c *gin.Context, status int, msg string) {
	slog.ErrorContext(c.Request.Context(), msg, "status", status)
	c.AbortWithStatus(status)
}

func newInternalError(c *gin.Context, err error, msg ...string) {
	slog.ErrorContext(c.Request.Context(), err.Error(), "status", "internal error")

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
