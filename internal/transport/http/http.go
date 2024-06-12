package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/olexsmir/onasty/internal/service/usersrv"
	"github.com/olexsmir/onasty/internal/transport/http/apiv1"
)

type Transport struct {
	usersrv usersrv.UserServicer
}

func NewTransport(us usersrv.UserServicer) *Transport {
	return &Transport{
		usersrv: us,
	}
}

func (t *Transport) Handler() http.Handler {
	r := gin.New()
	r.Use(
		gin.Recovery(),
		t.logger(),
	)

	api := r.Group("/api")
	api.GET("/ping", t.pingHandler)
	apiv1.NewAPIV1(t.usersrv).Routes(api.Group("/v1"))

	return r.Handler()
}

func (*Transport) pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
