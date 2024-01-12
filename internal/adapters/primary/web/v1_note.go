package web

import "github.com/gin-gonic/gin"

func (h *Handler) bindV1Note(r *gin.RouterGroup) {
	note := r.Group("/note")
	{
		note.GET("/:slug", h.v1GetNoteBySlug)
		note.POST("", h.v1CreateNote)
	}
}

func (h *Handler) v1CreateNote(c *gin.Context) {
}

func (h *Handler) v1GetNoteBySlug(c *gin.Context) {
}
