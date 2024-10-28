package apiv1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/models"
)

type createNoteRequest struct {
	Content              string    `json:"content"`
	Slug                 string    `json:"slug"`
	Password             string    `json:"password"`
	BurnBeforeExpiration bool      `json:"burn_before_expiration"`
	ExpiresAt            time.Time `json:"expires_at"`
}

type createNoteResponse struct {
	Slug string `json:"slug"`
}

func (a *APIV1) createNoteHandler(c *gin.Context) {
	var req createNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		newError(c, http.StatusBadRequest, "invalid request")
		return
	}

	note := models.Note{ //nolint:exhaustruct
		Content:              req.Content,
		Slug:                 req.Slug,
		BurnBeforeExpiration: req.BurnBeforeExpiration,
		CreatedAt:            time.Now(),
		Password:             req.Password,
		ExpiresAt:            req.ExpiresAt,
	}

	if err := note.Validate(); err != nil {
		newErrorStatus(c, http.StatusBadRequest, err.Error())
		return
	}

	slug, err := a.notesrv.Create(c.Request.Context(), dtos.CreateNoteDTO{
		Content:              note.Content,
		UserID:               a.getUserID(c),
		Slug:                 note.Slug,
		Password:             note.Password,
		BurnBeforeExpiration: note.BurnBeforeExpiration,
		CreatedAt:            note.CreatedAt,
		ExpiresAt:            note.ExpiresAt,
	}, a.getUserID(c))
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.JSON(http.StatusCreated, createNoteResponse{slug})
}

type getNoteBySlugResponse struct {
	Content   string    `json:"content"`
	CratedAt  time.Time `json:"crated_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (a *APIV1) getNoteBySlugHandler(c *gin.Context) {
	slug := c.Param("slug")
	note, err := a.notesrv.GetBySlugAndRemoveIfNeeded(c.Request.Context(), slug)
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, getNoteBySlugResponse{
		Content:   note.Content,
		CratedAt:  note.CreatedAt,
		ExpiresAt: note.ExpiresAt,
	})
}
