package web

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olexsmir/onasty/internal/core/domain"
)

func (h *Handler) bindV1Note(r *gin.RouterGroup) {
	note := r.Group("/note")
	{
		note.GET("/:slug", h.v1GetNoteBySlug)
		note.POST("", h.v1CreateNote)
	}
}

type v1CreateNoteRequest struct {
	Content   string    `json:"content"    binding:"required"`
	Slug      string    `json:"slug"`
	ExpiresAt time.Time `json:"expires_at"`
}

type v1CreateNoteResponse struct {
	ID string `json:"id"`
}

func (h *Handler) v1CreateNote(c *gin.Context) {
	var req v1CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		newRespones(c, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.noteServce.Create(c.Request.Context(), domain.Note{
		Content:   req.Content,
		Slug:      req.Slug,
		CreatedAt: time.Now(),
		ExpiresAt: req.ExpiresAt,
	})
	if err != nil {
		// TODO: handle app specific errors

		newRespones(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, v1CreateNoteResponse{
		ID: res.String(),
	})
}

func (h *Handler) v1GetNoteBySlug(c *gin.Context) {
}
