package e2e

import (
	"net/http"
	"time"

	"github.com/olexsmir/onasty/internal/core/domain"
)

type noteResponse struct {
	Slug string `json:"slug"`
}

func (s *AppTestSuite) TestNote_Create_AllOpts() {
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

	var res noteResponse
	s.readBodyAndUnjsonify(httpResp.Body, &res)

	dbNote := s.getNoteFromDBBySlug(res.Slug)

	s.Equal(http.StatusCreated, httpResp.Code)
	s.Equal(slug, dbNote.Slug)
	s.Equal(content, dbNote.Content)
	s.Equal(burnBeforeExpiration, dbNote.BurnBeforeExpiration)
	s.Equal(expireAt.Unix(), dbNote.ExpiresAt.Unix())
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

	var res noteResponse
	s.readBodyAndUnjsonify(httpResp.Body, &res)

	dbNote := s.getNoteFromDBBySlug(res.Slug)

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

	var res noteResponse
	s.readBodyAndUnjsonify(httpResp.Body, &res)

	dbNote := s.getNoteFromDBBySlug(res.Slug)

	s.Equal(http.StatusCreated, httpResp.Code)
	s.Equal(res.Slug, dbNote.Slug)
	s.Equal(content, dbNote.Content)
}
