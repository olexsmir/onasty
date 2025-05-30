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
	Password string `json:"password"`
}

type getNoteBySlugResponse struct {
	Content   string    `json:"content"`
	ReadAt    time.Time `json:"read_at,omitzero"`
	CreatedAt time.Time `json:"created_at"`
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
		CreatedAt: note.CreatedAt,
		ExpiresAt: note.ExpiresAt,
	})
}

type getNotesResponse struct {
	Content              string    `json:"content"`
	Slug                 string    `json:"slug"`
	BurnBeforeExpiration bool      `json:"burn_before_expiration"`
	HasPassword          bool      `json:"has_password"`
	CreatedAt            time.Time `json:"created_at"`
	ExpiresAt            time.Time `json:"expires_at,omitzero"`
	ReadAt               time.Time `json:"read_at,omitzero"`
}

func (a *APIV1) getNotesHandler(c *gin.Context) {
	notes, err := a.notesrv.GetAllNotesByAuthorID(c.Request.Context(), a.getUserID(c))
	if err != nil {
		errorResponse(c, err)
		return
	}

	var response []getNotesResponse
	for _, note := range notes {
		response = append(response, getNotesResponse{
			Content:              note.Content,
			Slug:                 note.Slug,
			BurnBeforeExpiration: note.BurnBeforeExpiration,
			HasPassword:          note.HasPassword,
			CreatedAt:            note.CreatedAt,
			ExpiresAt:            note.ExpiresAt,
			ReadAt:               note.ReadAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

func (a *APIV1) updateNoteHandler(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}
