package apiv1

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/olexsmir/onasty/internal/service/usersrv"
)

var (
	ErrAuthorizationHeaderIsNotValid = errors.New("authorization header is not valid")
	ErrUnauthorized                  = errors.New("unauthorized")
)

const userIDCtxKey = "userID"

func (a *APIV1) authorizedMiddleware(c *gin.Context) {
	token, ok := getTokenFromAuthHeaders(c)
	if !ok {
		errorResponse(c, ErrUnauthorized)
		return
	}

	if err := saveUserIDToCtx(c, a.userSrv, token); err != nil {
		errorResponse(c, err)
		return
	}

	c.Next()
}

func (a *APIV1) couldBeAuthorizedMiddleware(c *gin.Context) {
	token, ok := getTokenFromAuthHeaders(c)
	if ok {
		if err := saveUserIDToCtx(c, a.userSrv, token); err != nil {
			newInternalError(c, err)
			return
		}
	}

	c.Next()
}

func getTokenFromAuthHeaders(c *gin.Context) (token string, ok bool) { //nolint:nonamedreturns
	header := c.GetHeader("Authorization")
	if header == "" {
		return "", false
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 && headerParts[0] != "Bearer" {
		return "", false
	}

	if len(headerParts[1]) == 0 {
		return "", false
	}

	return headerParts[1], true
}

func saveUserIDToCtx(c *gin.Context, us usersrv.UserServicer, token string) error {
	pl, err := us.ParseToken(token)
	if err != nil {
		return err
	}

	c.Set(userIDCtxKey, pl.UserID)

	return nil
}

// getUserId returns userId from the context
// getting user id is only possible if user is authorized
func getUserID(c *gin.Context) uuid.UUID {
	userID, exists := c.Get(userIDCtxKey)
	if !exists {
		return uuid.Nil
	}
	return uuid.Must(uuid.FromString(userID.(string)))
}
