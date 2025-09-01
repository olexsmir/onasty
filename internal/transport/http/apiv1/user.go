package apiv1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olexsmir/onasty/internal/dtos"
)

type getMeResponse struct {
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	LastLoginAt  time.Time `json:"last_login_at"`
	NotesCreated int       `json:"notes_created"`
}

func (a *APIV1) getMeHandler(c *gin.Context) {
	uinfo, err := a.usersrv.GetUserInfo(c.Request.Context(), a.getUserID(c))
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, getMeResponse{
		Email:        uinfo.Email,
		CreatedAt:    uinfo.CreatedAt,
		LastLoginAt:  uinfo.LastLoginAt,
		NotesCreated: uinfo.NotesCreated,
	})
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (a *APIV1) changePasswordHandler(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		invalidRequest(c)
		return
	}

	if err := a.usersrv.ChangePassword(
		c.Request.Context(),
		a.getUserID(c),
		dtos.ChangeUserPassword{
			CurrentPassword: req.CurrentPassword,
			NewPassword:     req.NewPassword,
		}); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusOK)
}

type requestResetPasswordRequest struct {
	Email string `json:"email"`
}

func (a *APIV1) requestResetPasswordHandler(c *gin.Context) {
	var req requestResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		invalidRequest(c)
		return
	}

	if err := a.usersrv.RequestPasswordReset(
		c.Request.Context(),
		dtos.RequestResetPassword{
			Email: req.Email,
		},
	); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusOK)
}

type resetPasswordRequest struct {
	Password string `json:"password"`
}

func (a *APIV1) resetPasswordHandler(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		invalidRequest(c)
		return
	}

	if err := a.usersrv.ResetPassword(
		c.Request.Context(),
		dtos.ResetPassword{
			Token:       c.Param("token"),
			NewPassword: req.Password,
		},
	); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusOK)
}

type changeEmailRequest struct {
	NewEmail string `json:"new_email"`
}

func (a *APIV1) requestEmailChangeHandler(c *gin.Context) {
	var req changeEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		invalidRequest(c)
		return
	}

	if err := a.usersrv.RequestEmailChange(
		c.Request.Context(),
		a.getUserID(c),
		dtos.ChangeEmail{
			NewEmail: req.NewEmail,
		}); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (a *APIV1) changeEmailHandler(c *gin.Context) {
	if err := a.usersrv.ChangeEmail(
		c.Request.Context(),
		c.Param("token"),
	); err != nil {
		errorResponse(c, err)
		return
	}

	c.String(http.StatusOK, "email changed")
}

func (a *APIV1) verifyHandler(c *gin.Context) {
	if err := a.usersrv.Verify(
		c.Request.Context(),
		c.Param("token"),
	); err != nil {
		errorResponse(c, err)
		return
	}

	c.String(http.StatusOK, "email verified")
}

type resendVerificationEmailRequest struct {
	Email string `json:"email"`
}

func (a *APIV1) resendVerificationEmailHandler(c *gin.Context) {
	var req resendVerificationEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		invalidRequest(c)
		return
	}

	if err := a.usersrv.ResendVerificationEmail(
		c.Request.Context(),
		dtos.ResendVerificationEmail{
			Email: req.Email,
		}); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusOK)
}
