package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/henvic/pgq"
	"github.com/jackc/pgx/v5"
	"github.com/olexsmir/onasty/internal/core/domain"
)

func (s *AppTestSuite) jsonify(v map[string]any) []byte {
	r, err := json.Marshal(v)
	s.require.NoError(err)

	return r
}

func (s *AppTestSuite) readBodyAndUnjsonify(b *bytes.Buffer, res any) {
	respData, err := io.ReadAll(b)
	s.require.NoError(err)

	err = json.Unmarshal(respData, &res)
	s.require.NoError(err)
}

func (s *AppTestSuite) httpRequest(method, url string, body []byte) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	s.require.NoError(err)

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
	s.require.NoError(err)

	var res domain.Note
	err = s.postgresDB.QueryRow(s.ctx, query, args...).
		Scan(&res.ID, &res.Content, &res.Slug, &res.BurnBeforeExpiration, &res.CreatedAt, &res.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Note{}
	}

	s.require.NoError(err)
	return res
}

func (s *AppTestSuite) insertNote(note domain.Note) {
	query, args, err := pgq.
		Insert("notes").
		Columns("id", "content", "slug", "burn_before_expiration ", "created_at", "expires_at").
		Values(note.ID, note.Content, note.Slug, note.BurnBeforeExpiration, note.CreatedAt, note.ExpiresAt).
		Returning("id", "slug").
		SQL()
	s.require.NoError(err)

	_, err = s.postgresDB.Exec(s.ctx, query, args...)
	s.require.NoError(err)
}
