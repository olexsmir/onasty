package web

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	userIDCtxKey = "userID"
)

// v1AuthorizedMiddleware is a middleware that checks if user is authorized
// and if so sets user metadata to context
//
// being authorized is required for making the request for specific endpoint
func (h *Handler) v1AuthorizedMiddleware(c *gin.Context) {
	header := c.GetHeader("Authorization")
	if header == "" {
		newError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		newError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if len(headerParts[1]) == 0 {
		newError(c, http.StatusUnauthorized, "token is empty")
		return
	}

	userID, err := h.userService.ParseToken(headerParts[1])
	if err != nil {
		newInternalError(c, err)
		return
	}

	c.Set(userIDCtxKey, userID)
	c.Next()
}

func getUserID(c *gin.Context) uuid.UUID {
	userID, exists := c.Get(userIDCtxKey)
	if !exists {
		return uuid.Nil
	}

	if id, ok := userID.(uuid.UUID); ok {
		return id
	}

	return uuid.Nil
}

// v1CouldBeAuthorizedMiddleware is a middleware that checks if user is authorized and
// if so sets user metadata to context
//
// it is NOT required to be authorized for making the request for specific endpoint
func (Handler) v1CouldBeAuthorizedMiddleware(c *gin.Context) {
	c.Next()
}
