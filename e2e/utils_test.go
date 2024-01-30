package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/henvic/pgq"
	"github.com/olexsmir/onasty/internal/core/domain"
)

func (s *AppTestSuite) jsonify(v map[string]any) []byte {
	r, err := json.Marshal(v)
	s.NoError(err)

	return r
}

func (s *AppTestSuite) readBodyAndUnjsonify(b *bytes.Buffer, res any) {
	respData, err := io.ReadAll(b)
	s.NoError(err)

	err = json.Unmarshal(respData, &res)
	s.NoError(err)
}

func (s *AppTestSuite) httpRequest(method, url string, body []byte) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")

	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	return resp
}

func (s *AppTestSuite) getNoteFromDBBySlug(slug string) domain.Note {
	query, args, err := pgq.
		Select("id", "content", "slug", "burn_before_expiration", "created_at", "expires_at").
		From("notes").
		Where("slug = ?", slug).
		SQL()
	s.NoError(err)

	var res domain.Note
	err = s.db.QueryRow(s.ctx, query, args...).
		Scan(&res.ID, &res.Content, &res.Slug, &res.BurnBeforeExpiration, &res.CreatedAt, &res.ExpiresAt)
	s.NoError(err)

	return res
}
