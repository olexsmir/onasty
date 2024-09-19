// reqid provides gin-gonic/gin middleware to generate a requestid for each request
package reqid

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RequestIDKey string

const (
	RequestID RequestIDKey = "request_id"

	headerRequestID = "X-Request-ID"
)

// Middleware initializes the request ID
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(headerRequestID)
		if rid == "" {
			rid = uuid.New().String()
			c.Request.Header.Add(headerRequestID, rid)
		}

		// set reqeust ID request context
		ctx := context.WithValue(c.Request.Context(), RequestID, rid)
		c.Request = c.Request.WithContext(ctx)

		// ensures that the request ID is in the response
		c.Header(headerRequestID, rid)
		c.Next()
	}
}

// Get returns the request ID
func Get(c *gin.Context) string {
	v, ok := c.Request.Context().Value(RequestID).(string)
	if !ok {
		return ""
	}

	return v
}
