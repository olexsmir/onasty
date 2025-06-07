package apiv1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olexsmir/onasty/internal/dtos"
)

type signUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *APIV1) signUpHandler(c *gin.Context) {
	var req signUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		newError(c, http.StatusBadRequest, "invalid request")
		return
	}

	if _, err := a.usersrv.SignUp(c.Request.Context(), dtos.SignUp{
		Email:       req.Email,
		Password:    req.Password,
		CreatedAt:   time.Now(),
		LastLoginAt: time.Now(),
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

	toks, err := a.usersrv.SignIn(c.Request.Context(), dtos.SignIn{
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

	if err := a.usersrv.ResendVerificationEmail(
		c.Request.Context(),
		dtos.SignIn{
			Email:    req.Email,
			Password: req.Password,
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
		newError(c, http.StatusBadRequest, "invalid request")
		return
	}

	if err := a.usersrv.RequestPasswordReset(c.Request.Context(), dtos.RequestResetPassword{
		Email: req.Email,
	}); err != nil {
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
		newError(c, http.StatusBadRequest, "invalid request")
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

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (a *APIV1) logOutHandler(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		newError(c, http.StatusBadRequest, "invalid request")
		return
	}

	if err := a.usersrv.Logout(c.Request.Context(), a.getUserID(c), req.RefreshToken); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (a *APIV1) logOutAllHandler(c *gin.Context) {
	if err := a.usersrv.LogoutAll(c.Request.Context(), a.getUserID(c)); err != nil {
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

const oatuhStateCookie = "oauth_state"

func (a *APIV1) oauthLoginHandler(c *gin.Context) {
	redirectInfo, err := a.usersrv.GetOAuthURL(c.Param("provider"))
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.SetCookie(
		oatuhStateCookie,
		redirectInfo.State,
		int(time.Minute.Seconds()),
		"/",
		a.domain,
		!a.env.IsDevMode(),
		true,
	)

	c.Redirect(http.StatusSeeOther, redirectInfo.URL)
}

func (a *APIV1) oauthCallbackHandler(c *gin.Context) {
	state := c.Query("state")
	storedState, err := c.Cookie(oatuhStateCookie)
	if err != nil || state != storedState {
		newError(c, http.StatusBadRequest, "invalid oauth state")
		return
	}

	tokens, err := a.usersrv.HandleOAuthLogin(
		c.Request.Context(),
		c.Param("provider"),
		c.Query("code"),
	)
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, signInResponse{
		AccessToken:  tokens.Access,
		RefreshToken: tokens.Refresh,
	})
}
