package reqid

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:gochecknoinits
func init() {
	gin.SetMode(gin.TestMode)
}

func testHandler(c *gin.Context) {
	c.Status(http.StatusOK)
}

func TestMiddleware(t *testing.T) {
	r := gin.New()
	r.Use(Middleware())
	r.GET("/", testHandler)

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get(headerRequestID))
}

func BenchmarkMiddleware(b *testing.B) {
	r := gin.New()
	r.Use(Middleware())
	r.GET("/", testHandler)

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(b, err)

	for b.Loop() {
		r.ServeHTTP(w, req)
	}
}
