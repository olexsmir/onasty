package e2e

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gofrs/uuid/v5"
)

type apiv1NoteCreateRequest struct {
	Content              string    `json:"content"`
	Slug                 string    `json:"slug"`
	BurnBeforeExpiration bool      `json:"burn_before_expiration"`
	ExpiresAt            time.Time `json:"expires_at"`
}
type apiv1NoteCreateResponse struct {
	Slug string `json:"slug"`
}

func (e *AppTestSuite) TestNoteV1_Create_unauthorized() {
	tests := []struct {
		name   string
		inp    apiv1NoteCreateRequest
		assert func(*httptest.ResponseRecorder, apiv1NoteCreateRequest)
	}{
		{
			name: "empty request",
			inp:  apiv1NoteCreateRequest{},
			assert: func(r *httptest.ResponseRecorder, _ apiv1NoteCreateRequest) {
				e.Equal(r.Code, http.StatusBadRequest)
			},
		},
		{
			name: "content only",
			inp:  apiv1NoteCreateRequest{Content: e.uuid()},
			assert: func(r *httptest.ResponseRecorder, _ apiv1NoteCreateRequest) {
				e.Equal(r.Code, http.StatusCreated)

				var body apiv1NoteCreateResponse
				e.readBodyAndUnjsonify(r.Body, &body)

				_, err := uuid.FromString(body.Slug)
				e.require.NoError(err)

				dbNote := e.getNoteFromDBbySlug(body.Slug)
				e.NotEmpty(dbNote)
			},
		},
		{
			name: "set slug",
			inp: apiv1NoteCreateRequest{
				Slug:    e.uuid() + "fuker",
				Content: e.uuid(),
			},
			assert: func(r *httptest.ResponseRecorder, inp apiv1NoteCreateRequest) {
				e.Equal(r.Code, http.StatusCreated)

				var body apiv1NoteCreateResponse
				e.readBodyAndUnjsonify(r.Body, &body)

				dbNote := e.getNoteFromDBbySlug(inp.Slug)
				e.NotEmpty(dbNote)
			},
		},
		{
			name: "all possible fields",
			inp: apiv1NoteCreateRequest{
				Content:              e.uuid(),
				BurnBeforeExpiration: true,
				ExpiresAt:            time.Now().Add(time.Hour),
			},
			assert: func(r *httptest.ResponseRecorder, inp apiv1NoteCreateRequest) {
				e.Equal(r.Code, http.StatusCreated)

				var body apiv1NoteCreateResponse
				e.readBodyAndUnjsonify(r.Body, &body)

				dbNote := e.getNoteFromDBbySlug(body.Slug)
				e.NotEmpty(dbNote)

				e.Equal(dbNote.Content, inp.Content)
				e.Equal(dbNote.BurnBeforeExpiration, inp.BurnBeforeExpiration)
				e.Equal(dbNote.ExpiresAt.Unix(), inp.ExpiresAt.Unix())
			},
		},
	}

	for _, tt := range tests {
		httpResp := e.httpRequest(http.MethodPost, "/api/v1/note", e.jsonify(tt.inp))
		tt.assert(httpResp, tt.inp)
	}
}

func (e *AppTestSuite) TestNoteV1_Create_authorized() {
	e.T().Skip("TODO: the app logic isn't there so " +
		"i can't be used, please implement this freaking logic")
}

type apiv1NoteGetResponse struct{} //nolint:unused

func (e *AppTestSuite) TestNoteV1_Get() {
	e.T().Skip()
}
