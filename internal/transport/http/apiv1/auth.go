package apiv1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olexsmir/onasty/internal/dtos"
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

	if _, err := a.userSrv.SignUp(c.Request.Context(), dtos.CreateUserDTO{
		Username:    req.Username,
		Email:       req.Email,
		Password:    req.Password,
		CreatedAt:   time.Now(),
		LastLoginAt: time.Now(),
	}); err != nil {
		errorHandler(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

func (a *APIV1) signInHandler(_ *gin.Context) {}

func (a *APIV1) refreshTokensHandler(_ *gin.Context) {}

func (a *APIV1) logOutHandler(_ *gin.Context) {}
