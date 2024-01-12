package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/olexsmir/onasty/internal/ports"
)

type HandlerDeps struct {
	NoteService ports.NoteServicer
}

type Handler struct {
	noteServce ports.NoteServicer
}

func NewHandler(deps HandlerDeps) *Handler {
	return &Handler{
		noteServce: deps.NoteService,
	}
}

func (h *Handler) InitRoutes() http.Handler {
	r := gin.Default()
	r.Use(
		gin.Recovery(),
		gin.Logger(),
	)

	api := r.Group("/api")
	api.GET("/ping", h.pingHandler)
	h.bindV1Routes(api.Group("/v1"))

	return r.Handler()
}

func (h *Handler) bindV1Routes(r *gin.RouterGroup) {
	h.bindV1Note(r)
}

func (h *Handler) pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
