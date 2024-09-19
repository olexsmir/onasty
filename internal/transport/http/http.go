package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/olexsmir/onasty/internal/service/notesrv"
	"github.com/olexsmir/onasty/internal/service/usersrv"
	"github.com/olexsmir/onasty/internal/transport/http/apiv1"
	"github.com/olexsmir/onasty/internal/transport/http/reqid"
)

type Transport struct {
	usersrv usersrv.UserServicer
	notesrv notesrv.NoteServicer
}

func NewTransport(
	us usersrv.UserServicer,
	ns notesrv.NoteServicer,
) *Transport {
	return &Transport{
		usersrv: us,
		notesrv: ns,
	}
}

func (t *Transport) Handler() http.Handler {
	r := gin.New()
	r.Use(
		gin.Recovery(),
		reqid.Middleware(),
		t.logger(),
	)

	api := r.Group("/api")
	api.GET("/ping", t.pingHandler)
	apiv1.NewAPIV1(t.usersrv, t.notesrv).Routes(api.Group("/v1"))

	return r.Handler()
}

func (*Transport) pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
