package e2e

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/olexsmir/onasty/internal/core/domain"
)

type v1createNoteResponse struct {
	Slug string `json:"slug"`
}

type v1getNoteResponse struct {
	Content  string    `json:"content"`
	CratedAt time.Time `json:"crated_at"`
}

func (s *AppTestSuite) TestNoteV1_Create_AllOpts() {
	content := "testing"
	slug := "some-semi-random-slug"
	burnBeforeExpiration := true
	expireAt := time.Now().Add(7 * time.Minute)

	httpResp := s.httpRequest(http.MethodPost, "/api/v1/note", s.jsonify(map[string]any{
		"content":                content,
		"slug":                   slug,
		"burn_before_expiration": burnBeforeExpiration,
		"expires_at":             expireAt,
	}))

	var res v1createNoteResponse
	s.readBodyAndUnjsonify(httpResp.Body, &res)

	dbNote := s.getNoteFromDBBySlug(res.Slug)

	s.Equal(http.StatusCreated, httpResp.Code)
	s.Equal(slug, dbNote.Slug)
	s.Equal(content, dbNote.Content)
	s.Equal(burnBeforeExpiration, dbNote.BurnBeforeExpiration)
	s.Equal(expireAt.Unix(), dbNote.ExpiresAt.Unix())
}

func (s *AppTestSuite) TestNoteV1_Create_CantBeWirhoutContent() {
	httpResp := s.httpRequest(http.MethodPost, "/api/v1/note", s.jsonify(map[string]any{}))

	var res errorResponse
	s.readBodyAndUnjsonify(httpResp.Body, &res)

	s.Equal(http.StatusBadRequest, httpResp.Code)
	s.Equal(res.Message, domain.ErrNoteContentIsEmpty.Error())
}

func (s *AppTestSuite) TestNoteV1_Create_RandomSlug() {
	httpResp := s.httpRequest(http.MethodPost, "/api/v1/note", s.jsonify(map[string]any{
		"content": "testing",
	}))

	var res v1createNoteResponse
	s.readBodyAndUnjsonify(httpResp.Body, &res)

	dbNote := s.getNoteFromDBBySlug(res.Slug)

	s.Equal(http.StatusCreated, httpResp.Code)
	s.NotEmpty(dbNote)
}

func (s *AppTestSuite) TestNoteV1_Create_ExplicitSlug() {
	slug := "test-slug"
	content := "testing"

	httpResp := s.httpRequest(http.MethodPost, "/api/v1/note", s.jsonify(map[string]any{
		"content": content,
		"slug":    slug,
	}))

	var res v1createNoteResponse
	s.readBodyAndUnjsonify(httpResp.Body, &res)

	dbNote := s.getNoteFromDBBySlug(res.Slug)

	s.Equal(http.StatusCreated, httpResp.Code)
	s.Equal(res.Slug, dbNote.Slug)
	s.Equal(content, dbNote.Content)
}

func (s *AppTestSuite) TestNoteV1_Create_SlugAlreadyInUse() {
	note := domain.Note{
		ID:      uuid.New(),
		Content: "content",
		Slug:    uuid.New().String(),
	}
	s.insertNote(note)

	httpResp := s.httpRequest(http.MethodPost, "/api/v1/note", s.jsonify(map[string]any{
		"content": "testing",
		"slug":    note.Slug,
	}))

	var res errorResponse
	s.readBodyAndUnjsonify(httpResp.Body, &res)

	s.Equal(http.StatusBadRequest, httpResp.Code)
	s.Equal(res.Message, domain.ErrNoteSlugIsAlreadyInUse.Error())
}

func (s *AppTestSuite) TestNoteV1_Create_ExpiresAtInThePast() {
	httpResp := s.httpRequest(http.MethodPost, "/api/v1/note", s.jsonify(map[string]any{
		"content":    "testing",
		"expires_at": time.Now().Add(-1 * time.Minute),
	}))

	var res errorResponse
	s.readBodyAndUnjsonify(httpResp.Body, &res)

	s.Equal(http.StatusBadRequest, httpResp.Code)
	s.Equal(res.Message, domain.ErrNoteExpired.Error())
}

func (s *AppTestSuite) TestNoteV1_Get_ShouldAndReturnNoteAndRemoveIt() {
	note := domain.Note{
		ID:        uuid.New(),
		Content:   "content",
		Slug:      uuid.New().String(),
		CreatedAt: time.Now(),
	}
	s.insertNote(note)

	httpResp := s.httpRequest(
		http.MethodGet,
		"/api/v1/note/"+note.Slug,
		s.jsonify(map[string]any{}),
	)

	var res v1getNoteResponse
	s.readBodyAndUnjsonify(httpResp.Body, &res)

	dbNote := s.getNoteFromDBBySlug(note.Slug)

	s.Equal(http.StatusOK, httpResp.Code)
	s.Equal(note.Content, res.Content)
	s.Equal(note.CreatedAt.Unix(), res.CratedAt.Unix())
	s.Empty(dbNote)
}

func (s *AppTestSuite) TestNoteV1_Get_TheresNoSuchNote() {
	httpResp := s.httpRequest(
		http.MethodGet,
		"/api/v1/note/"+uuid.New().String(),
		s.jsonify(map[string]any{}),
	)

	var res errorResponse
	s.readBodyAndUnjsonify(httpResp.Body, &res)

	s.Equal(http.StatusNotFound, httpResp.Code)
}

func (s *AppTestSuite) TestNoteV1_Get_Expired() {
	note := domain.Note{
		ID:        uuid.New(),
		Content:   "content",
		Slug:      uuid.New().String(),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	}
	s.insertNote(note)

	httpResp := s.httpRequest(
		http.MethodGet,
		"/api/v1/note/"+note.Slug,
		s.jsonify(map[string]any{}),
	)

	var res errorResponse
	s.readBodyAndUnjsonify(httpResp.Body, &res)

	s.Equal(http.StatusNotFound, httpResp.Code)
	s.Equal(res.Message, domain.ErrNoteExpired.Error())
}

func (s *AppTestSuite) TestNoteV1_Get_RespectBurnBeforeExpirationOption_setToTrue() {
	note := domain.Note{
		ID:                   uuid.New(),
		Content:              "content",
		Slug:                 uuid.New().String(),
		BurnBeforeExpiration: true,
		CreatedAt:            time.Now(),
		ExpiresAt:            time.Now().Add(1 * time.Minute),
	}
	s.insertNote(note)

	httpResp := s.httpRequest(
		http.MethodGet,
		"/api/v1/note/"+note.Slug,
		s.jsonify(map[string]any{}),
	)

	var res v1getNoteResponse
	s.readBodyAndUnjsonify(httpResp.Body, &res)

	dbNote := s.getNoteFromDBBySlug(note.Slug)

	s.Empty(dbNote)
	s.Equal(http.StatusOK, httpResp.Code)
	s.Equal(note.Content, res.Content)
	s.Equal(note.CreatedAt.Unix(), res.CratedAt.Unix())
}

func (s *AppTestSuite) TestNoteV1_Get_RespectBurnBeforeExpirationOption_setToFalse() {
	note := domain.Note{
		ID:                   uuid.New(),
		Content:              "content",
		Slug:                 uuid.New().String(),
		BurnBeforeExpiration: false,
		CreatedAt:            time.Now(),
		ExpiresAt:            time.Now().Add(1 * time.Minute),
	}
	s.insertNote(note)

	httpResp := s.httpRequest(
		http.MethodGet,
		"/api/v1/note/"+note.Slug,
		s.jsonify(map[string]any{}),
	)

	var res v1getNoteResponse
	s.readBodyAndUnjsonify(httpResp.Body, &res)

	dbNote := s.getNoteFromDBBySlug(note.Slug)

	s.NotEmpty(dbNote)
	s.Equal(http.StatusOK, httpResp.Code)
	s.Equal(note.Content, res.Content)
	s.Equal(note.CreatedAt.Unix(), res.CratedAt.Unix())
}

func (s *AppTestSuite) TestNoteV1_Get_RespectBurnBeforeExpirationOption_getAfterExpiration() {
	note := domain.Note{
		ID:                   uuid.New(),
		Content:              "content",
		Slug:                 uuid.New().String(),
		BurnBeforeExpiration: false,
		CreatedAt:            time.Now().Add(-time.Minute),
		ExpiresAt:            time.Now(),
	}
	s.insertNote(note)

	httpResp := s.httpRequest(
		http.MethodGet,
		"/api/v1/note/"+note.Slug,
		s.jsonify(map[string]any{}),
	)

	var res errorResponse
	s.readBodyAndUnjsonify(httpResp.Body, &res)

	s.Equal(http.StatusNotFound, httpResp.Code)
	s.Equal(res.Message, domain.ErrNoteExpired.Error())
}
