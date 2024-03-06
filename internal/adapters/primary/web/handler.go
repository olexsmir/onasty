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

	r.NoRoute(h.respondWithMsgAndStatus("not found", http.StatusNotFound))
	r.NoMethod(h.respondWithMsgAndStatus("method not allowed", http.StatusMethodNotAllowed))

	api := r.Group("/api")
	api.GET("/ping", h.pingHandler)
	h.bindV1Routes(api.Group("/v1"))

	return r.Handler()
}

func (h *Handler) bindV1Routes(r *gin.RouterGroup) {
	h.bindV1Note(r)
	h.bindV1Auth(r)
}

func (h *Handler) pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func (Handler) respondWithMsgAndStatus(msg string, code int) gin.HandlerFunc {
	return func(c *gin.Context) {
		slog.Info(msg)
		c.AbortWithStatus(code)
	}
}
