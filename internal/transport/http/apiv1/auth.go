package apiv1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/models"
)

type signUpRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *APIV1) signUpHandler(c *gin.Context) {
	var req signUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		newError(c, http.StatusBadRequest, "invalid request")
		return
	}

	user := models.User{
		Username:    req.Username,
		Email:       req.Email,
		Password:    req.Password,
		CreatedAt:   time.Now(),
		LastLoginAt: time.Now(),
	}
	if err := user.Validate(); err != nil {
		// TODO: find a way to return all errors at once
		newErrorStatus(c, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := a.usersrv.SignUp(c.Request.Context(), dtos.CreateUserDTO{
		Username:    user.Username,
		Email:       user.Email,
		Password:    user.Password,
		CreatedAt:   user.CreatedAt,
		LastLoginAt: user.LastLoginAt,
	}); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

type signInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signInResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (a *APIV1) signInHandler(c *gin.Context) {
	var req signInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		newError(c, http.StatusBadRequest, "invalid request")
		return
	}

	toks, err := a.usersrv.SignIn(c.Request.Context(), dtos.SignInDTO{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, signInResponse{
		AccessToken:  toks.Access,
		RefreshToken: toks.Refresh,
	})
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (a *APIV1) refreshTokensHandler(c *gin.Context) {
	var req refreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		newError(c, http.StatusBadRequest, "invalid request")
		return
	}

	toks, err := a.usersrv.RefreshTokens(c.Request.Context(), req.RefreshToken)
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, signInResponse{
		AccessToken:  toks.Access,
		RefreshToken: toks.Refresh,
	})
}

func (a *APIV1) verifyHandler(c *gin.Context) {
	if err := a.usersrv.Verify(c.Request.Context(), c.Param("token")); err != nil {
		errorResponse(c, err)
		return
	}

	c.String(http.StatusOK, "email verified")
}

func (a *APIV1) resendVerificationEmailHandler(c *gin.Context) {
	var req signInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		newError(c, http.StatusBadRequest, "invalid request")
		return
	}

	if err := a.usersrv.ResendVerificationEmail(c.Request.Context(), dtos.SignInDTO{
		Email:    req.Email,
		Password: req.Password,
	}); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusOK)
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type forgotPasswordResponse struct {
	NewPassword string `json:"new_password"`
}

func (a *APIV1) forgotPasswordHandler(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		newError(c, http.StatusBadRequest, "invalid request")
		return
	}

	newTempPassword, err := a.usersrv.ForgotPassword(c.Request.Context(), req.Email)
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, forgotPasswordResponse{
		NewPassword: newTempPassword,
	})
}

func (a *APIV1) logOutHandler(c *gin.Context) {
	if err := a.usersrv.Logout(c.Request.Context(), a.getUserID(c)); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (a *APIV1) changePasswordHandler(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		newError(c, http.StatusBadRequest, "invalid request")
		return
	}

	if err := a.usersrv.ChangePassord(
		c.Request.Context(),
		dtos.ResetUserPasswordDTO{
			UserID:          a.getUserID(c),
			CurrentPassword: req.CurrentPassword,
			NewPassword:     req.NewPassword,
		}); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusOK)
}
