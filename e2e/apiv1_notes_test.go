package e2e_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/olexsmir/onasty/internal/models"
)

type (
	apiv1NoteCreateRequest struct {
		Content              string    `json:"content"`
		Slug                 string    `json:"slug"`
		Password             string    `json:"password"`
		BurnBeforeExpiration bool      `json:"burn_before_expiration"`
		ExpiresAt            time.Time `json:"expires_at"`
	}
	apiv1NoteCreateResponse struct {
		Slug string `json:"slug"`
	}
)

func (e *AppTestSuite) TestNoteV1_Create() {
	tests := []struct {
		name   string
		inp    apiv1NoteCreateRequest
		assert func(*httptest.ResponseRecorder, apiv1NoteCreateRequest)
	}{
		{
			name: "empty request",
			inp:  apiv1NoteCreateRequest{}, //nolint:exhaustruct
			assert: func(r *httptest.ResponseRecorder, _ apiv1NoteCreateRequest) {
				e.Equal(r.Code, http.StatusBadRequest)
			},
		},
		{
			name: "content only",
			inp:  apiv1NoteCreateRequest{Content: e.uuid()}, //nolint:exhaustruct
			assert: func(r *httptest.ResponseRecorder, _ apiv1NoteCreateRequest) {
				e.Equal(r.Code, http.StatusCreated)

				var body apiv1NoteCreateResponse
				e.readBodyAndUnjsonify(r.Body, &body)

				_, err := uuid.FromString(body.Slug)
				e.require.NoError(err)

				dbNote := e.getNoteBySlug(body.Slug)
				e.NotEmpty(dbNote)
			},
		},
		{
			name: "set slug",
			inp: apiv1NoteCreateRequest{ //nolint:exhaustruct
				Slug:    e.uuid() + "fuker",
				Content: e.uuid(),
			},
			assert: func(r *httptest.ResponseRecorder, inp apiv1NoteCreateRequest) {
				e.Equal(r.Code, http.StatusCreated)

				var body apiv1NoteCreateResponse
				e.readBodyAndUnjsonify(r.Body, &body)

				dbNote := e.getNoteBySlug(inp.Slug)
				e.NotEmpty(dbNote)
			},
		},
		{
			name: "invalid slug, with space",
			inp: apiv1NoteCreateRequest{ //nolint:exhaustruct
				Slug:    e.uuid() + "fuker fuker",
				Content: e.uuid(),
			},
			assert: func(r *httptest.ResponseRecorder, _ apiv1NoteCreateRequest) {
				e.Equal(http.StatusBadRequest, r.Code)
			},
		},
		{
			name: "invalid slug, with slash",
			inp: apiv1NoteCreateRequest{ //nolint:exhaustruct
				Slug:    e.uuid() + "fuker/fuker",
				Content: e.uuid(),
			},
			assert: func(r *httptest.ResponseRecorder, _ apiv1NoteCreateRequest) {
				e.Equal(http.StatusBadRequest, r.Code)
			},
		},
		{
			name: "invalid slug, 'read'",
			inp: apiv1NoteCreateRequest{ //nolint:exhaustruct
				Slug:    "read",
				Content: e.uuid(),
			},
			assert: func(r *httptest.ResponseRecorder, _ apiv1NoteCreateRequest) {
				e.Equal(r.Code, http.StatusBadRequest)

				var body errorResponse
				e.readBodyAndUnjsonify(r.Body, &body)

				e.Equal(models.ErrNoteSlugIsAlreadyInUse.Error(), body.Message)
			},
		},
		{
			name: "invalid slug, 'unread'",
			inp: apiv1NoteCreateRequest{ //nolint:exhaustruct
				Slug:    "unread",
				Content: e.uuid(),
			},
			assert: func(r *httptest.ResponseRecorder, _ apiv1NoteCreateRequest) {
				e.Equal(r.Code, http.StatusBadRequest)

				var body errorResponse
				e.readBodyAndUnjsonify(r.Body, &body)

				e.Equal(models.ErrNoteSlugIsAlreadyInUse.Error(), body.Message)
			},
		},
		{
			name: "slug provided but empty",
			inp: apiv1NoteCreateRequest{ //nolint:exhaustruct
				Slug:    "",
				Content: e.uuid(),
			},
			assert: func(r *httptest.ResponseRecorder, inp apiv1NoteCreateRequest) {
				e.Equal(r.Code, http.StatusCreated)

				var body apiv1NoteCreateResponse
				e.readBodyAndUnjsonify(r.Body, &body)

				dbNote := e.getNoteBySlug(body.Slug)
				e.NotEmpty(dbNote)
				e.Equal(inp.Content, dbNote.Content)
			},
		},
		{
			name: "burn before expiration, but without expiration time",
			inp: apiv1NoteCreateRequest{ //nolint:exhaustruct
				Content:              e.uuid(),
				BurnBeforeExpiration: true,
			},
			assert: func(r *httptest.ResponseRecorder, _ apiv1NoteCreateRequest) {
				var body errorResponse
				e.readBodyAndUnjsonify(r.Body, &body)

				e.Equal(models.ErrNoteCannotBeBurnt.Error(), body.Message)
			},
		},
		{
			name: "set password",
			inp: apiv1NoteCreateRequest{ //nolint:exhaustruct
				Content:  e.uuid(),
				Password: e.uuid(),
			},
			assert: func(r *httptest.ResponseRecorder, _ apiv1NoteCreateRequest) {
				e.Equal(http.StatusCreated, r.Code)
			},
		},
		{
			name: "all possible fields",
			inp: apiv1NoteCreateRequest{ //nolint:exhaustruct
				Content:              e.uuid(),
				BurnBeforeExpiration: true,
				ExpiresAt:            time.Now().Add(time.Hour),
			},
			assert: func(r *httptest.ResponseRecorder, inp apiv1NoteCreateRequest) {
				e.Equal(r.Code, http.StatusCreated)

				var body apiv1NoteCreateResponse
				e.readBodyAndUnjsonify(r.Body, &body)

				dbNote := e.getNoteBySlug(body.Slug)
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

type apiv1NoteGetResponse struct {
	Content   string     `json:"content"`
	ReadAt    *time.Time `json:"read_at"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt time.Time  `json:"expires_at"`
}

func (e *AppTestSuite) TestNoteV1_Get() {
	// create note
	content := e.uuid()
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/note",
		e.jsonify(apiv1NoteCreateRequest{ //nolint:exhaustruct
			Content: content,
		}),
	)
	e.Equal(http.StatusCreated, httpResp.Code)

	var bodyCreated apiv1NoteCreateResponse
	e.readBodyAndUnjsonify(httpResp.Body, &bodyCreated)

	// read note
	httpResp2 := e.httpRequest(http.MethodGet, "/api/v1/note/"+bodyCreated.Slug, nil)
	e.Equal(http.StatusOK, httpResp2.Code)

	var body apiv1NoteGetResponse
	e.readBodyAndUnjsonify(httpResp2.Body, &body)

	e.Equal(content, body.Content)

	dbNote := e.getNoteBySlug(bodyCreated.Slug)
	e.Empty(dbNote.Content)
	e.False(dbNote.ReadAt.IsZero())
}

func (e *AppTestSuite) TestNoteV1_Get_alreadyRead() {
	// create note
	content := e.uuid()
	httpRespCreated := e.httpRequest(
		http.MethodPost,
		"/api/v1/note",
		e.jsonify(apiv1NoteCreateRequest{Content: content}), //nolint:exhaustruct
	)
	e.Equal(http.StatusCreated, httpRespCreated.Code)

	var bodyCreated apiv1NoteCreateResponse
	e.readBodyAndUnjsonify(httpRespCreated.Body, &bodyCreated)

	// read note
	httpRespRead := e.httpRequest(http.MethodGet, "/api/v1/note/"+bodyCreated.Slug, nil)
	e.Equal(httpRespRead.Code, http.StatusOK)

	var bodyRead apiv1NoteGetResponse
	e.readBodyAndUnjsonify(httpRespRead.Body, &bodyRead)

	e.Equal(content, bodyRead.Content)

	dbNote := e.getNoteBySlug(bodyCreated.Slug)
	e.Empty(dbNote.Content)
	e.False(dbNote.ReadAt.IsZero())

	// // read not once again
	httpRespRead2 := e.httpRequest(http.MethodGet, "/api/v1/note/"+bodyCreated.Slug, nil)
	e.Equal(http.StatusNotFound, httpRespRead2.Code)

	var bodyRead2 apiv1NoteGetResponse
	e.readBodyAndUnjsonify(httpRespRead2.Body, &bodyRead2)

	dbNote2 := e.getNoteBySlug(bodyCreated.Slug)
	e.Empty(dbNote2.Content)

	e.Empty(bodyRead2.Content)
	e.Equal(dbNote2.ReadAt.Unix(), bodyRead2.ReadAt.Unix())
	e.Equal(dbNote2.CreatedAt.Unix(), bodyRead2.CreatedAt.Unix())
	e.Equal(dbNote2.ExpiresAt.Unix(), bodyRead2.ExpiresAt.Unix())
}

type apiv1NoteGetWithPasswordRequest struct {
	Password string `json:"password"`
}

func (e *AppTestSuite) TestNoteV1_GetWithPassword() {
	content := e.uuid()
	passwd := e.uuid()
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/note",
		e.jsonify(apiv1NoteCreateRequest{ //nolint:exhaustruct
			Content:  content,
			Password: passwd,
		}),
	)
	e.Equal(http.StatusCreated, httpResp.Code)

	var bodyCreated apiv1NoteCreateResponse
	e.readBodyAndUnjsonify(httpResp.Body, &bodyCreated)

	httpResp = e.httpRequest(
		http.MethodPost,
		"/api/v1/note/"+bodyCreated.Slug+"/view",
		e.jsonify(apiv1NoteGetWithPasswordRequest{
			Password: passwd,
		}),
	)
	e.Equal(httpResp.Code, http.StatusOK)

	var body apiv1NoteGetResponse
	e.readBodyAndUnjsonify(httpResp.Body, &body)

	e.Equal(content, body.Content)

	dbNote := e.getNoteBySlug(bodyCreated.Slug)
	e.Equal(dbNote.Content, "")
	e.Equal(dbNote.ReadAt.IsZero(), false)
}

func (e *AppTestSuite) TestNoteV1_GetWithPassword_wrongNoPassword() {
	content := e.uuid()
	passwd := e.uuid()
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/note",
		e.jsonify(apiv1NoteCreateRequest{ //nolint:exhaustruct
			Content:  content,
			Password: passwd,
		}),
	)
	e.Equal(http.StatusCreated, httpResp.Code)

	var bodyCreated apiv1NoteCreateResponse
	e.readBodyAndUnjsonify(httpResp.Body, &bodyCreated)

	httpResp = e.httpRequest(http.MethodGet, "/api/v1/note/"+bodyCreated.Slug, nil)
	e.Equal(httpResp.Code, http.StatusNotFound)
}

func (e *AppTestSuite) TestNoteV1_GetWithPassword_wrong() {
	content := e.uuid()
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/note",
		e.jsonify(apiv1NoteCreateRequest{ //nolint:exhaustruct
			Content:  content,
			Password: e.uuid(),
		}),
	)
	e.Equal(http.StatusCreated, httpResp.Code)

	var bodyCreated apiv1NoteCreateResponse
	e.readBodyAndUnjsonify(httpResp.Body, &bodyCreated)

	httpResp = e.httpRequest(
		http.MethodPost,
		"/api/v1/note/"+bodyCreated.Slug+"/view",
		e.jsonify(apiv1NoteGetWithPasswordRequest{
			Password: e.uuid(),
		}),
	)
	e.Equal(httpResp.Code, http.StatusNotFound)
}

type apiv1NoteMetadataResponse struct {
	CreatedAt   time.Time `json:"created_at"`
	HasPassword bool      `json:"has_password"`
}

func (e *AppTestSuite) TestNoteV1_GetMetadata() {
	// create note
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/note",
		e.jsonify(apiv1NoteCreateRequest{Content: "content"}), //nolint:exhaustruct
	)
	e.Equal(http.StatusCreated, httpResp.Code)

	var bodyCreated apiv1NoteCreateResponse
	e.readBodyAndUnjsonify(httpResp.Body, &bodyCreated)

	// get metadata
	metaResp := e.httpRequest(http.MethodGet, "/api/v1/note/"+bodyCreated.Slug+"/meta", []byte{})
	e.Equal(metaResp.Code, http.StatusOK)

	var metadata apiv1NoteMetadataResponse
	e.readBodyAndUnjsonify(metaResp.Body, &metadata)

	e.False(metadata.HasPassword)
	e.NotEmpty(metadata.CreatedAt)
}

func (e *AppTestSuite) TestNoteV1_GetMetadata_withPassword() {
	// create note
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/note",
		e.jsonify(apiv1NoteCreateRequest{ //nolint:exhaustruct
			Content:  "content",
			Password: "pass",
		}),
	)
	e.Equal(http.StatusCreated, httpResp.Code)

	var bodyCreated apiv1NoteCreateResponse
	e.readBodyAndUnjsonify(httpResp.Body, &bodyCreated)

	// get metadata
	metaResp := e.httpRequest(http.MethodGet, "/api/v1/note/"+bodyCreated.Slug+"/meta", []byte{})
	e.Equal(http.StatusOK, metaResp.Code)

	var metadata apiv1NoteMetadataResponse
	e.readBodyAndUnjsonify(metaResp.Body, &metadata)

	e.True(metadata.HasPassword)
	e.NotEmpty(metadata.CreatedAt)
}

func (e *AppTestSuite) TestNoteV1_GetMetadata_notFound() {
	metaResp := e.httpRequest(http.MethodGet, "/api/v1/note/"+e.uuid()+"/meta", []byte{})
	e.Equal(http.StatusNotFound, metaResp.Code)
}

func (e *AppTestSuite) TestNoteV1_GetMetadata_readNote() {
	// create note
	createdResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/note",
		e.jsonify(apiv1NoteCreateRequest{Content: "content"}), //nolint:exhaustruct
	)
	e.Equal(http.StatusCreated, createdResp.Code)

	var bodyCreated apiv1NoteCreateResponse
	e.readBodyAndUnjsonify(createdResp.Body, &bodyCreated)

	// read note
	readResp := e.httpRequest(http.MethodGet, "/api/v1/note/"+bodyCreated.Slug, nil)
	e.Equal(http.StatusOK, readResp.Code)

	// get metadata
	metaResp := e.httpRequest(http.MethodGet, "/api/v1/note/"+e.uuid()+"/meta", nil)
	e.Equal(http.StatusNotFound, metaResp.Code)
}
