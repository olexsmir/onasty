package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_getVisitor(t *testing.T) {
	limiter := newLimiter(10, 20, time.Second)
	ip := visitorIP("127.0.0.1")

	visitor := limiter.getVisitor(ip)
	assert.NotNil(t, visitor)

	visitorAgain := limiter.getVisitor(ip)
	assert.Equal(t, visitor, visitorAgain)

	assert.Len(t, limiter.visitors, 1)
}

// TODO: rewrite to use [testing/synctest] when it gets merged
func TestRateLimiter_cleanupVisitors(t *testing.T) {
	limiter := newLimiter(10, 20, time.Second/2)
	limiter.getVisitor("192.168.9.1")
	assert.Len(t, limiter.visitors, 1)

	time.Sleep(time.Second)
	limiter.cleanUpVisitors()
	assert.Empty(t, limiter.visitors)
}

func TestMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := map[string]struct {
		config       Config
		requests     int
		expectedCode int
	}{
		"allows requests with in limit": {
			config: Config{
				RPS:   2,
				Burst: 2,
				TTL:   time.Minute,
			},
			requests:     1,
			expectedCode: http.StatusOK,
		},
		"blocks requests over limit": {
			config: Config{
				RPS:   1,
				Burst: 1,
				TTL:   time.Minute,
			},
			requests:     2,
			expectedCode: http.StatusTooManyRequests,
		},
		"allows burst requests": {
			config: Config{
				RPS:   1,
				Burst: 3,
				TTL:   time.Minute,
			},
			requests:     3,
			expectedCode: http.StatusOK,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			handler := MiddlewareWithConfig(tt.config)
			var lastCode int

			for range tt.requests {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

				handler(c)
				lastCode = w.Code
			}

			assert.Equal(t, tt.expectedCode, lastCode)
		})
	}
}
