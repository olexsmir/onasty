package e2e

import (
	"net/http"
)

type noteResponse struct {
	Slug string `json:"slug"`
}

func (s *AppTestSuite) TestNote_Create_RandomSlug() {
	resp := s.httpRequest("POST", "/api/v1/note", s.jsonify(map[string]any{
		"content": "testing",
	}))

	var note noteResponse
	s.readBodyAndUnjsonify(resp.Body, &note)

	dbNote := s.getNoteFromDBBySlug(note.Slug)

	s.Equal(http.StatusCreated, resp.Code)
	s.NotEmpty(dbNote)
}
