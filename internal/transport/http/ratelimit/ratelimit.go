// thanks to https://www.alexedwards.net/blog/how-to-rate-limit-http-requests

package ratelimit

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type (
	rateLimiter struct {
		mu sync.RWMutex

		visitors map[visitorIP]*visitor

		// limit is the maximum number of requests per second
		limit rate.Limit

		// ttl is the time after which a visitor is forgotten
		ttl time.Duration

		// burst is the maximum number of requests that can be made in a short amount of time
		burst int
	}

	visitorIP string
	visitor   struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
)

func newLimiter(rps, burst int, ttl time.Duration) *rateLimiter {
	return &rateLimiter{ //nolint:exhaustruct
		visitors: make(map[visitorIP]*visitor),
		limit:    rate.Limit(rps),
		burst:    burst,
		ttl:      ttl,
	}
}

// Retrieve and return the rate limiter for the current visitor if it
// already exists. Otherwise create a new rate limiter and add it to
// the visitors map, using the IP address as the key.
func (r *rateLimiter) getVisitor(ip visitorIP) *rate.Limiter {
	r.mu.RLock()
	v, exists := r.visitors[ip]
	r.mu.RUnlock()

	if !exists {
		limit := rate.NewLimiter(r.limit, r.burst)

		r.mu.Lock()
		r.visitors[ip] = &visitor{
			limiter:  limit,
			lastSeen: time.Now(),
		}
		r.mu.Unlock()

		return limit
	}

	r.mu.Lock()
	v.lastSeen = time.Now()
	r.mu.Unlock()

	return v.limiter
}

// Every minute check the map for visitors that haven't been seen for
// more than 3 minutes and delete the entries.
func (r *rateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		r.mu.Lock()
		for ip, v := range r.visitors {
			if time.Since(v.lastSeen) > r.ttl {
				delete(r.visitors, ip)
			}
		}
		r.mu.Unlock()
	}
}

type Config struct {
	// RPS is the maximum number of requests per second
	RPS int

	// TTL is the time after which a visitor is forgotten
	TTL time.Duration

	// Burst is the maximum number of requests that can be made in a short amount of time
	Burst int
}

// MiddlewareWithConfig returns a new rate limiting middleware with the given config
func MiddlewareWithConfig(c Config) gin.HandlerFunc {
	lmt := newLimiter(c.RPS, c.Burst, c.TTL)
	go lmt.cleanupVisitors()

	return func(c *gin.Context) {
		visitor := lmt.getVisitor(visitorIP(c.ClientIP()))
		if visitor == nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if !visitor.Allow() {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		c.Next()
	}
}
