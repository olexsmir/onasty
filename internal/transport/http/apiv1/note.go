package apiv1

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/service/notesrv"
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

	slug, err := a.notesrv.Create(c.Request.Context(), dtos.CreateNote{
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

type getNoteBySlugRequest struct {
	Password string `json:"password,omitempty"`
}

type getNoteBySlugResponse struct {
	Content   string    `json:"content,omitempty"`
	ReadAt    time.Time `json:"read_at,omitzero"`
	CratedAt  time.Time `json:"crated_at,omitzero"`
	ExpiresAt time.Time `json:"expires_at,omitzero"`
}

func (a *APIV1) getNoteBySlugHandler(c *gin.Context) {
	var req getNoteBySlugRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		newError(c, http.StatusBadRequest, "invalid request")
		return
	}

	note, err := a.notesrv.GetBySlugAndRemoveIfNeeded(
		c.Request.Context(),
		notesrv.GetNoteBySlugInput{
			Slug:     c.Param("slug"),
			Password: req.Password,
		},
	)
	if err != nil {
		errorResponse(c, err)
		return
	}

	status := http.StatusOK
	if !note.ReadAt.IsZero() {
		status = http.StatusNotFound
	}

	c.JSON(status, getNoteBySlugResponse{
		Content:   note.Content,
		ReadAt:    note.ReadAt,
		CratedAt:  note.CreatedAt,
		ExpiresAt: note.ExpiresAt,
	})
}
