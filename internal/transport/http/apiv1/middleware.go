package apiv1

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/olexsmir/onasty/internal/metrics"
	"github.com/olexsmir/onasty/internal/models"
)

const userIDCtxKey = "userID"

// authorizedMiddleware is a middleware that checks if user is authorized
// and if so sets user metadata to context
//
// being authorized is required for making the request for specific endpoint
func (a *APIV1) authorizedMiddleware(c *gin.Context) {
	token, ok := getTokenFromAuthHeaders(c)
	if !ok {
		errorResponse(c, ErrUnauthorized)
		return
	}

	uid, err := a.validateAuthorizedUser(c.Request.Context(), token)
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.Set(userIDCtxKey, uid)

	c.Next()
}

// couldBeAuthorizedMiddleware is a middleware that checks if user is authorized and
// if so sets user metadata to context
//
// it is NOT required to be authorized for making the request for specific endpoint
func (a *APIV1) couldBeAuthorizedMiddleware(c *gin.Context) {
	token, ok := getTokenFromAuthHeaders(c)
	if !ok {
		c.Next()
		return
	}

	uid, err := a.validateAuthorizedUser(c.Request.Context(), token)
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.Set(userIDCtxKey, uid)
	c.Next()
}

func (a *APIV1) metricsMiddleware(c *gin.Context) {
	start := time.Now()
	c.Next()
	latency := time.Since(start)

	metrics.RecordLatencyRequestMetric(c.Request.Method, c.Request.RequestURI, latency)

	if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
		metrics.RecordSuccessfulRequestMetric(c.Request.Method, c.Request.RequestURI)
	}

	if c.Writer.Status() >= 400 {
		metrics.RecordFailedRequestMetric(c.Request.Method, c.Request.RequestURI)
	}
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

// getUserId returns userId from the context
// getting user id is only possible if user is authorized
//
// if userID is not set, [uuid.Nil] will be returned.
func (a *APIV1) getUserID(c *gin.Context) uuid.UUID {
	userID, exists := c.Get(userIDCtxKey)
	if !exists {
		return uuid.Nil
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}

	return uid
}

func (a *APIV1) validateAuthorizedUser(ctx context.Context, accessToken string) (uuid.UUID, error) {
	tokenPayload, err := a.authsrv.ParseJWTToken(accessToken)
	if err != nil {
		return uuid.Nil, err
	}

	userID := uuid.Must(uuid.FromString(tokenPayload.UserID))

	ok, err := a.authsrv.CheckIfUserExists(ctx, userID)
	if err != nil {
		return uuid.Nil, err
	}

	if !ok {
		return uuid.Nil, ErrUnauthorized
	}

	ok, err = a.authsrv.CheckIfUserIsActivated(ctx, userID)
	if err != nil {
		return uuid.Nil, err
	}

	if !ok {
		return uuid.Nil, models.ErrUserIsNotActivated
	}

	return userID, nil
}
