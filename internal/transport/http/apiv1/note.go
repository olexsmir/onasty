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
	KeepBeforeExpiration bool      `json:"keep_before_expiration"`
	ExpiresAt            time.Time `json:"expires_at"`
}

type createNoteResponse struct {
	Slug string `json:"slug"`
}

func (a *APIV1) createNoteHandler(c *gin.Context) {
	var req createNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		invalidRequest(c)
		return
	}

	slug, err := a.notesrv.Create(c.Request.Context(), dtos.CreateNote{
		Content:              req.Content,
		UserID:               a.getUserID(c),
		Slug:                 req.Slug,
		Password:             req.Password,
		KeepBeforeExpiration: req.KeepBeforeExpiration,
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
	KeepBeforeExpiration bool      `json:"keep_before_expiration"`
	CreatedAt            time.Time `json:"created_at"`
	ExpiresAt            time.Time `json:"expires_at,omitzero"`
}

func (a *APIV1) getNoteBySlugHandler(c *gin.Context) {
	note, err := a.notesrv.GetBySlugAndRemoveIfNeeded(
		c.Request.Context(),
		notesrv.GetNoteBySlugInput{
			Slug:     c.Param("slug"),
			Password: notesrv.EmptyPassword,
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
		KeepBeforeExpiration: note.KeepBeforeExpiration,
	})
}

type getNoteBuySlugAndPasswordRequest struct {
	Password string `json:"password"`
}

func (a *APIV1) getNoteBySlugAndPasswordHandler(c *gin.Context) {
	var req getNoteBuySlugAndPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		invalidRequest(c)
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
		KeepBeforeExpiration: note.KeepBeforeExpiration,
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
	KeepBeforeExpiration bool      `json:"keep_before_expiration"`
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

	c.JSON(http.StatusOK, mapNotesDTOToResponse(notes))
}

func (a *APIV1) getReadNotesHandler(c *gin.Context) {
	notes, err := a.notesrv.GetAllReadByAuthorID(c.Request.Context(), a.getUserID(c))
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, mapNotesDTOToResponse(notes))
}

func (a *APIV1) getUnReadNotesHandler(c *gin.Context) {
	notes, err := a.notesrv.GetAllUnreadByAuthorID(c.Request.Context(), a.getUserID(c))
	if err != nil {
		errorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, mapNotesDTOToResponse(notes))
}

type updateNoteRequest struct {
	ExpiresAt            *time.Time `json:"expires_at,omitempty"`
	KeepBeforeExpiration *bool      `json:"keep_before_expiration,omitempty"`
}

func (a *APIV1) updateNoteHandler(c *gin.Context) {
	var req updateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		invalidRequest(c)
		return
	}

	if err := a.notesrv.UpdateExpirationTimeSettings(
		c.Request.Context(),
		dtos.PatchNote{
			KeepBeforeExpiration: req.KeepBeforeExpiration,
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
		invalidRequest(c)
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

func mapNotesDTOToResponse(notes []dtos.NoteDetailed) []getNotesResponse {
	var response []getNotesResponse
	for _, note := range notes {
		response = append(response, getNotesResponse{
			Content:              note.Content,
			Slug:                 note.Slug,
			KeepBeforeExpiration: note.KeepBeforeExpiration,
			HasPassword:          note.HasPassword,
			CreatedAt:            note.CreatedAt,
			ExpiresAt:            note.ExpiresAt,
			ReadAt:               note.ReadAt,
		})
	}

	return response
}
