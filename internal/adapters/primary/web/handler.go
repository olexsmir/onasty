package web

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/olexsmir/onasty/internal/ports"
)

type HandlerDeps struct {
	NoteService ports.NoteServicer
	UserService ports.UserServicer
}

type Handler struct {
	noteService ports.NoteServicer
	userService ports.UserServicer
}

func NewHandler(deps HandlerDeps) *Handler {
	return &Handler{
		noteService: deps.NoteService,
		userService: deps.UserService,
	}
}

func (h *Handler) InitRoutes() http.Handler {
	r := gin.New()
	r.Use(
		gin.Recovery(),
		h.logger(),
	)

	r.NoRoute(h.notFoundHandler)
	r.NoMethod(h.notFoundHandler)

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

func (h *Handler) notFoundHandler(c *gin.Context) {
	slog.Info("not found")
	c.AbortWithStatus(http.StatusNotFound)
}
