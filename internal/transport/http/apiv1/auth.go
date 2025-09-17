package apiv1

import (
	"net/http"
	"net/url"
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
		invalidRequest(c)
		return
	}

	if err := a.authsrv.SignUp(c.Request.Context(), dtos.SignUp{
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
		invalidRequest(c)
		return
	}

	toks, err := a.authsrv.SignIn(c.Request.Context(), dtos.SignIn{
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
		invalidRequest(c)
		return
	}

	toks, err := a.authsrv.RefreshTokens(c.Request.Context(), req.RefreshToken)
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, signInResponse{
		AccessToken:  toks.Access,
		RefreshToken: toks.Refresh,
	})
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (a *APIV1) logOutHandler(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		invalidRequest(c)
		return
	}

	if err := a.authsrv.Logout(
		c.Request.Context(),
		a.getUserID(c),
		req.RefreshToken,
	); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (a *APIV1) logOutAllHandler(c *gin.Context) {
	if err := a.authsrv.LogoutAll(c.Request.Context(), a.getUserID(c)); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

const oatuhStateCookie = "oauth_state"

func (a *APIV1) oauthLoginHandler(c *gin.Context) {
	redirectInfo, err := a.authsrv.GetOAuthURL(c.Param("provider"))
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.SetCookie(
		oatuhStateCookie,
		redirectInfo.State,
		int(time.Minute.Seconds()),
		"/",
		a.frontendURL,
		!a.env.IsDevMode(),
		true,
	)

	c.Redirect(http.StatusSeeOther, redirectInfo.URL)
}

func (a *APIV1) oauthCallbackHandler(c *gin.Context) {
	redURL, err := url.Parse(a.frontendURL + "/oauth/callback")
	if err != nil {
		errorResponse(c, err)
		return
	}

	storedState, err := c.Cookie(oatuhStateCookie)
	if err != nil || c.Query("state") != storedState {
		a.oauthCallbackErrorResponse(c, redURL)
		return
	}

	tokens, err := a.authsrv.HandleOAuthLogin(
		c.Request.Context(),
		c.Param("provider"),
		c.Query("code"),
	)
	if err != nil {
		a.oauthCallbackErrorResponse(c, redURL)
		return
	}

	redURL.RawQuery = url.Values{
		"access_token":  {tokens.Access},
		"refresh_token": {tokens.Refresh},
	}.Encode()

	c.Redirect(http.StatusFound, redURL.String())
}

func (a *APIV1) oauthCallbackErrorResponse(c *gin.Context, u *url.URL) {
	u.RawQuery = url.Values{"error": {"internal server error"}}.Encode()
	c.Redirect(http.StatusFound, u.String())
}
