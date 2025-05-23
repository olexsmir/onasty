package http

import (
	"log/slog"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (t *Transport) corsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     t.corsAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           t.corsMaxAge,
	})
}

func (t *Transport) loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()
		latency := time.Since(start)

		if raw != "" {
			path = path + "?" + raw
		}

		lvl := slog.LevelInfo
		if c.Writer.Status() >= 400 {
			lvl = slog.LevelError
		}

		slog.LogAttrs(
			c.Request.Context(),
			lvl,
			c.Errors.ByType(gin.ErrorTypePrivate).String(),
			slog.String("latency", latency.String()),
			slog.String("method", c.Request.Method),
			slog.Int("status_code", c.Writer.Status()),
			slog.String("path", path),
			slog.String("client_ip", c.ClientIP()),
			slog.Int("body_size", c.Writer.Size()),
		)
	}
}
