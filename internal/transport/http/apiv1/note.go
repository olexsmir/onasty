package apiv1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olexsmir/onasty/internal/dtos"
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

	// TODO: burn_before_expiration shouldn't be set if user has not set or specified expires_at

	slug, err := a.notesrv.Create(c.Request.Context(), dtos.CreateNote{
		Content:              req.Content,
		UserID:               a.getUserID(c),
		Slug:                 req.Slug,
		Password:             req.Password,
		BurnBeforeExpiration: req.BurnBeforeExpiration,
		CreatedAt:            time.Now(),
		ExpiresAt:            req.ExpiresAt,
	}, a.getUserID(c))
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.JSON(http.StatusCreated, createNoteResponse{slug})
}

type getNoteBySlugResponse struct {
	Content              string    `json:"content"`
	ReadAt               time.Time `json:"read_at,omitzero"`
	BurnBeforeExpiration bool      `json:"burn_before_expiration"`
	CreatedAt            time.Time `json:"created_at"`
	ExpiresAt            time.Time `json:"expires_at,omitzero"`
}

func (a *APIV1) getNoteBySlugHandler(c *gin.Context) {
	note, err := a.notesrv.GetBySlugAndRemoveIfNeeded(
		c.Request.Context(),
		notesrv.GetNoteBySlugInput{
			Slug:     c.Param("slug"),
			Password: "",
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
		Content:              note.Content,
		ReadAt:               note.ReadAt,
		CreatedAt:            note.CreatedAt,
		ExpiresAt:            note.ExpiresAt,
		BurnBeforeExpiration: note.BurnBeforeExpiration,
	})
}

type getNoteBuySlugAndPasswordRequest struct {
	Password string `json:"password"`
}

func (a *APIV1) getNoteBySlugAndPasswordHandler(c *gin.Context) {
	var req getNoteBuySlugAndPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
		Content:              note.Content,
		ReadAt:               note.ReadAt,
		CreatedAt:            note.CreatedAt,
		ExpiresAt:            note.ExpiresAt,
		BurnBeforeExpiration: note.BurnBeforeExpiration,
	})
}

type getNoteMetadataBySlugResponse struct {
	CreatedAt   time.Time `json:"created_at"`
	HasPassword bool      `json:"has_password"`
}

func (a *APIV1) getNoteMetadataByIDHandler(c *gin.Context) {
	meta, err := a.notesrv.GetNoteMetadataBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, getNoteMetadataBySlugResponse{
		CreatedAt:   meta.CreatedAt,
		HasPassword: meta.HasPassword,
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
	notes, err := a.notesrv.GetAllByAuthorID(c.Request.Context(), a.getUserID(c))
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

type updateNoteRequest struct {
	ExpiresAt            *time.Time `json:"expires_at,omitempty"`
	BurnBeforeExpiration *bool      `json:"burn_before_expiration,omitempty"`
}

func (a *APIV1) updateNoteHandler(c *gin.Context) {
	var req updateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		newError(c, http.StatusBadRequest, "invalid request")
		return
	}

	// TODO: burn_before_expiration shouldn't be set if user has not set or specified expires_at

	if err := a.notesrv.UpdateExpirationTimeSettings(
		c.Request.Context(),
		dtos.PatchNote{
			BurnBeforeExpiration: req.BurnBeforeExpiration,
			ExpiresAt:            req.ExpiresAt,
		},
		c.Param("slug"),
		a.getUserID(c),
	); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (a *APIV1) deleteNoteHandler(c *gin.Context) {
	if err := a.notesrv.DeleteBySlug(
		c.Request.Context(),
		c.Param("slug"),
		a.getUserID(c),
	); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

type setNotePasswordRequest struct {
	Password string `json:"password"`
}

func (a *APIV1) setNotePasswordHandler(c *gin.Context) {
	var req setNotePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		newError(c, http.StatusBadRequest, "invalid request")
		return
	}

	if err := a.notesrv.UpdatePassword(
		c.Request.Context(),
		c.Param("slug"),
		req.Password,
		a.getUserID(c),
	); err != nil {
		errorResponse(c, err)
		return
	}

	c.Status(http.StatusOK)
}
