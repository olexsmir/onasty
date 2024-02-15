package web

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/olexsmir/onasty/internal/core/domain"
)

func (h *Handler) bindV1Auth(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/signup", h.v1SignUp)
		auth.POST("/signin", h.v1SignIn)
		auth.POST("/refresh-tokens", h.v1RefreshTokens)

		authorized := auth.Group("/")
		{
			authorized.Use(h.v1AuthorizedMiddleware)
			authorized.POST("/logout", h.v1Logout)
		}
	}
}

type v1SignUpRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h Handler) v1SignUp(c *gin.Context) {
	var req v1SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// TODO: do not expose internal(kinda) error
		newError(c, 400, err.Error())
		return
	}

	if err := h.userService.SignUp(c.Request.Context(), domain.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}); err != nil {
		if errors.Is(err, domain.ErrUserEmailIsAlreadyInUse) ||
			errors.Is(err, domain.ErrEmailIsInvalid) {
			newError(c, http.StatusNotFound, err.Error())
			return
		}

		newInternalError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *Handler) v1SignIn(c *gin.Context)

func (h *Handler) v1RefreshTokens(c *gin.Context)

func (h *Handler) v1Logout(c *gin.Context)
