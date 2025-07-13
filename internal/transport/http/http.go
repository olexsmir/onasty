package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olexsmir/onasty/internal/config"
	"github.com/olexsmir/onasty/internal/service/notesrv"
	"github.com/olexsmir/onasty/internal/service/usersrv"
	"github.com/olexsmir/onasty/internal/transport/http/apiv1"
	"github.com/olexsmir/onasty/internal/transport/http/ratelimit"
	"github.com/olexsmir/onasty/internal/transport/http/reqid"
)

type Transport struct {
	usersrv usersrv.UserServicer
	notesrv notesrv.NoteServicer

	env    config.Environment
	domain string

	corsAllowedOrigins []string
	corsMaxAge         time.Duration
	ratelimitCfg       ratelimit.Config
	slowRatelimitCfg   ratelimit.Config
}

func NewTransport(
	us usersrv.UserServicer,
	ns notesrv.NoteServicer,
	env config.Environment,
	domain string,
	corsAllowedOrigins []string,
	corsMaxAge time.Duration,
	ratelimitCfg ratelimit.Config,
	slowRatelimitCfg ratelimit.Config,
) *Transport {
	return &Transport{
		usersrv:            us,
		notesrv:            ns,
		env:                env,
		domain:             domain,
		corsAllowedOrigins: corsAllowedOrigins,
		corsMaxAge:         corsMaxAge,
		ratelimitCfg:       ratelimitCfg,
		slowRatelimitCfg:   slowRatelimitCfg,
	}
}

func (t *Transport) Handler() http.Handler {
	r := gin.New()
	r.Use(
		gin.Recovery(),
		reqid.Middleware(),
		t.loggerMiddleware(),
		t.corsMiddleware(),
		ratelimit.MiddlewareWithConfig(t.ratelimitCfg),
	)

	api := r.Group("/api")
	{
		api.GET("/ping", t.pingHandler)
		apiv1.
			NewAPIV1(t.usersrv, t.notesrv, t.slowRatelimitCfg, t.env, t.domain).
			Routes(api.Group("/v1"))
	}

	return r.Handler()
}

func (*Transport) pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
