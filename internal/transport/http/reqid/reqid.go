// reqid provides gin-gonic/gin middleware to generate a requestid for each request
package reqid

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
)

type requestIDKey string

const (
	RequestID requestIDKey = "request_id"

	headerRequestID = "X-Request-ID"
)

// Middleware initializes the request ID
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(headerRequestID)
		if rid == "" {
			rid = uuid.Must(uuid.NewV4()).String()
			c.Request.Header.Add(headerRequestID, rid)
		}

		// set request ID request context
		ctx := context.WithValue(c.Request.Context(), RequestID, rid)
		c.Request = c.Request.WithContext(ctx)

		// ensures that the request ID is in the response
		c.Header(headerRequestID, rid)
		c.Next()
	}
}

// Get returns the request ID
func Get(c *gin.Context) string {
	return c.GetHeader(headerRequestID)
}

// GetContext returns the request ID from context
func GetContext(ctx context.Context) string {
	rid, ok := ctx.Value(RequestID).(string)
	if !ok {
		return ""
	}
	return rid
}
