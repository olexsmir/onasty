package apiv1

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
}

func (a *APIV1) signInHandler(_ *gin.Context) {}

func (a *APIV1) refreshTokensHandler(_ *gin.Context) {}

func (a *APIV1) logOutHandler(_ *gin.Context) {}
