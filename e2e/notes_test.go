package e2e

import (
	"net/http"

	"github.com/olexsmir/onasty/internal/core/domain"
)

type noteResponse struct {
	Slug string `json:"slug"`
}

func (s *AppTestSuite) TestNote_Create_CantBeWirhoutContent() {
	httpResp := s.httpRequest(http.MethodPost, "/api/v1/note", s.jsonify(map[string]any{}))

	var res errorResponse
	s.readBodyAndUnjsonify(httpResp.Body, &res)

	s.Equal(http.StatusBadRequest, httpResp.Code)
	s.Equal(res.Message, domain.ErrNoteContentIsEmpty.Error())
}

func (s *AppTestSuite) TestNote_Create_RandomSlug() {
	httpResp := s.httpRequest(http.MethodPost, "/api/v1/note", s.jsonify(map[string]any{
		"content": "testing",
	}))

	var note noteResponse
	s.readBodyAndUnjsonify(httpResp.Body, &note)

	dbNote := s.getNoteFromDBBySlug(note.Slug)

	s.Equal(http.StatusCreated, httpResp.Code)
	s.NotEmpty(dbNote)
}

func (s *AppTestSuite) TestNote_Create_ExplicitSlug() {
	slug := "test-slug"
	content := "testing"

	httpResp := s.httpRequest(http.MethodPost, "/api/v1/note", s.jsonify(map[string]any{
		"content": content,
		"slug":    slug,
	}))

	var note noteResponse
	s.readBodyAndUnjsonify(httpResp.Body, &note)

	dbNote := s.getNoteFromDBBySlug(note.Slug)

	s.Equal(http.StatusCreated, httpResp.Code)
	s.Equal(note.Slug, dbNote.Slug)
	s.Equal(content, dbNote.Content)
}
