package e2e_test

import (
	"net/http"
	"slices"
	"time"
)

func (e *AppTestSuite) TestNoteV1_Create_authorized() {
	uid, toks := e.createAndSingIn(e.uuid()+"@test.com", "password")
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/note",
		e.jsonify(apiv1NoteCreateRequest{ //nolint:exhaustruct
			Content: "sample content for the test",
		}),
		toks.AccessToken,
	)

	var body apiv1NoteCreateResponse
	e.readBodyAndUnjsonify(httpResp.Body, &body)

	dbNote := e.getNoteBySlug(body.Slug)
	dbNoteAuthor := e.getLastNoteAuthorsRecordByAuthorID(uid)

	e.Equal(http.StatusCreated, httpResp.Code)
	e.Equal(dbNote.ID.String(), dbNoteAuthor.noteID.String())
}

func (e *AppTestSuite) TestNoteV1_Delete() {
	_, toks := e.createAndSingIn(e.uuid()+"@test.com", "password")
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/note",
		e.jsonify(apiv1NoteCreateRequest{ //nolint:exhaustruct
			Content: "sample content for the test",
		}),
		toks.AccessToken,
	)

	e.Equal(httpResp.Code, http.StatusCreated)

	var body apiv1NoteCreateResponse
	e.readBodyAndUnjsonify(httpResp.Body, &body)

	dbNote := e.getNoteBySlug(body.Slug)
	e.NotEmpty(dbNote)

	httpResp = e.httpRequest(
		http.MethodDelete,
		"/api/v1/note/"+body.Slug,
		nil,
		toks.AccessToken,
	)
	e.Equal(httpResp.Code, http.StatusNoContent)

	dbNote = e.getNoteBySlug(body.Slug)
	e.Empty(dbNote)
}

type apiV1NotePatchRequest struct {
	ExpiresAt            time.Time `json:"expires_at"`
	KeepBeforeExpiration bool      `json:"keep_before_expiration"`
}

func (e *AppTestSuite) TestNoteV1_updateExpirationTime() {
	_, toks := e.createAndSingIn(e.uuid()+"@test.com", "password")
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/note",
		e.jsonify(apiv1NoteCreateRequest{ //nolint:exhaustruct
			Content:              "sample content for the test",
			ExpiresAt:            time.Now().Add(time.Minute),
			KeepBeforeExpiration: false,
		}),
		toks.AccessToken,
	)

	e.Equal(httpResp.Code, http.StatusCreated)

	var body apiv1NoteCreateResponse
	e.readBodyAndUnjsonify(httpResp.Body, &body)

	patchTime := time.Now().Add(time.Hour)
	httpResp = e.httpRequest(
		http.MethodPatch,
		"/api/v1/note/"+body.Slug+"/expires",
		e.jsonify(apiV1NotePatchRequest{
			ExpiresAt:            patchTime,
			KeepBeforeExpiration: true,
		}),
		toks.AccessToken,
	)

	e.Equal(httpResp.Code, http.StatusOK)

	dbNote := e.getNoteBySlug(body.Slug)
	e.Equal(true, dbNote.KeepBeforeExpiration)
	e.Equal(patchTime.Unix(), dbNote.ExpiresAt.Unix())
}

func (e *AppTestSuite) TestNoteV1_updateExpirationTime_notFound() {
	_, toks := e.createAndSingIn(e.uuid()+"@test.com", "password")
	httpResp := e.httpRequest(
		http.MethodPatch,
		"/api/v1/note/"+e.uuid(),
		e.jsonify(apiV1NotePatchRequest{
			ExpiresAt:            time.Now().Add(time.Hour),
			KeepBeforeExpiration: true,
		}),
		toks.AccessToken,
	)

	e.Equal(httpResp.Code, http.StatusNotFound)
}

type apiV1NoteSetPasswordRequest struct {
	Password string `json:"password"`
}

func (e *AppTestSuite) TestNoteV1_UpdatePassword() {
	_, toks := e.createAndSingIn(e.uuid()+"@test.com", "password")
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/note",
		e.jsonify(apiv1NoteCreateRequest{ //nolint:exhaustruct
			Content: "content",
		}),
		toks.AccessToken,
	)

	e.Equal(httpResp.Code, http.StatusCreated)

	var body apiv1NoteCreateResponse
	e.readBodyAndUnjsonify(httpResp.Body, &body)

	dbNoteOriginal := e.getNoteBySlug(body.Slug)
	e.Empty(dbNoteOriginal.Password)

	passwd := "new-password"
	httpResp = e.httpRequest(
		http.MethodPatch,
		"/api/v1/note/"+body.Slug+"/password",
		e.jsonify(apiV1NoteSetPasswordRequest{
			Password: passwd,
		}),
		toks.AccessToken,
	)

	e.Equal(httpResp.Code, http.StatusOK)

	dbNote := e.getNoteBySlug(body.Slug)
	e.NotEmpty(dbNote.Password)

	err := e.hasher.Compare(dbNote.Password, passwd)
	e.require.NoError(err)
}

func (e *AppTestSuite) TestNoteV1_UpdatePassword_notFound() {
	_, toks := e.createAndSingIn(e.uuid()+"@test.com", "password")
	httpResp := e.httpRequest(
		http.MethodPatch,
		"/api/v1/note/"+e.uuid()+"/password",
		e.jsonify(apiV1NoteSetPasswordRequest{
			Password: "passwd",
		}),
		toks.AccessToken,
	)

	e.Equal(httpResp.Code, http.StatusNotFound)
}

func (e *AppTestSuite) TestNoteV1_UpdatePassword_passwordNotProvided() {
	_, toks := e.createAndSingIn(e.uuid()+"@test.com", "password")
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/note",
		e.jsonify(apiv1NoteCreateRequest{ //nolint:exhaustruct
			Content: "content",
		}),
		toks.AccessToken,
	)

	e.Equal(httpResp.Code, http.StatusCreated)

	var body apiv1NoteCreateResponse
	e.readBodyAndUnjsonify(httpResp.Body, &body)

	dbNoteOriginal := e.getNoteBySlug(body.Slug)
	e.Empty(dbNoteOriginal.Password)

	httpResp = e.httpRequest(
		http.MethodPatch,
		"/api/v1/note/"+body.Slug+"/password",
		e.jsonify(apiV1NoteSetPasswordRequest{
			Password: "",
		}),
		toks.AccessToken,
	)

	e.Equal(httpResp.Code, http.StatusBadRequest)
}

type apiv1NoteGetAllResponse struct {
	Content              string    `json:"content"`
	Slug                 string    `json:"slug"`
	KeepBeforeExpiration bool      `json:"keep_before_expiration"`
	HasPassword          bool      `json:"has_password"`
	CreatedAt            time.Time `json:"created_at"`
	ExpiresAt            time.Time `json:"expires_at"`
	ReadAt               time.Time `json:"read_at"`
}

func (e *AppTestSuite) TestNoteV1_GetAll() {
	notesInfo := []struct {
		slug    string
		content string
		read    bool
	}{
		{slug: e.uuid(), content: e.uuid(), read: true},
		{slug: e.uuid(), content: e.uuid(), read: true},
		{slug: e.uuid(), content: e.uuid(), read: false},
		{slug: e.uuid(), content: e.uuid(), read: false},
		{slug: e.uuid(), content: e.uuid(), read: false},
	}

	_, toks := e.createAndSingIn(e.uuid()+"@test.com", "password")

	// create notes
	for _, ni := range notesInfo {
		httpResp := e.httpRequest(
			http.MethodPost,
			"/api/v1/note",
			e.jsonify(apiv1NoteCreateRequest{ //nolint:exhaustruct
				Content: ni.content,
				Slug:    ni.slug,
			}),
			toks.AccessToken)

		e.Equal(http.StatusCreated, httpResp.Code)
	}

	// read notes
	for _, ni := range notesInfo {
		if ni.read {
			httpResp := e.httpRequest(http.MethodGet, "/api/v1/note/"+ni.slug, nil)
			e.Equal(http.StatusOK, httpResp.Code)
		}
	}

	httpResp := e.httpRequest(http.MethodGet, "/api/v1/note", nil, toks.AccessToken)

	var body []apiv1NoteGetAllResponse
	e.readBodyAndUnjsonify(httpResp.Body, &body)

	e.Equal(http.StatusOK, httpResp.Code)
	e.Len(body, len(notesInfo))
}

func (e *AppTestSuite) TestNoteV1_GetAllRead_inaccesibleForAnUnauthorized() {
	httpResp := e.httpRequest(http.MethodGet, "/api/v1/note/read", nil)
	e.Equal(httpResp.Code, http.StatusUnauthorized)
}

func (e *AppTestSuite) TestNoteV1_GetAllRead() {
	notesInfo := []struct {
		slug    string
		content string
	}{
		{slug: e.uuid(), content: e.uuid()},
		{slug: e.uuid(), content: e.uuid()},
		{slug: e.uuid(), content: e.uuid()},
		{slug: e.uuid(), content: e.uuid()},
	}

	_, toks := e.createAndSingIn(e.uuid()+"@test.com", "password")

	// create few notes
	for _, ni := range notesInfo {
		httpResp := e.httpRequest(
			http.MethodPost,
			"/api/v1/note",
			e.jsonify(apiv1NoteCreateRequest{ //nolint:exhaustruct
				Content: ni.content,
				Slug:    ni.slug,
			}),
			toks.AccessToken)

		e.Equal(http.StatusCreated, httpResp.Code)
	}

	// read those notes
	for _, ni := range notesInfo {
		httpResp := e.httpRequest(http.MethodGet, "/api/v1/note/"+ni.slug, nil)
		e.Equal(http.StatusOK, httpResp.Code)
	}

	// check if all notes are returned
	httpResp := e.httpRequest(http.MethodGet, "/api/v1/note/read", nil, toks.AccessToken)

	var body []apiv1NoteGetAllResponse
	e.readBodyAndUnjsonify(httpResp.Body, &body)

	e.Equal(http.StatusOK, httpResp.Code)
	e.require.Len(body, len(notesInfo))
}

func (e *AppTestSuite) TestNoteV1_GetAllUnread_inaccesibleForAnUnauthorized() {
	httpResp := e.httpRequest(http.MethodGet, "/api/v1/note/unread", nil)
	e.Equal(httpResp.Code, http.StatusUnauthorized)
}

func (e *AppTestSuite) TestNoteV1_GetAllUnread() {
	type notesTestData struct {
		slug    string
		content string
		read    bool
	}

	notesInfo := []notesTestData{
		{slug: e.uuid(), content: e.uuid(), read: true},
		{slug: e.uuid(), content: e.uuid(), read: true},
		{slug: e.uuid(), content: e.uuid(), read: true},
		{slug: e.uuid(), content: e.uuid(), read: false},
		{slug: e.uuid(), content: e.uuid(), read: false},
		{slug: e.uuid(), content: e.uuid(), read: false},
		{slug: e.uuid(), content: e.uuid(), read: false},
		{slug: e.uuid(), content: e.uuid(), read: false},
	}
	unreadNotesTotal := len(
		slices.DeleteFunc(
			slices.Clone(notesInfo),
			func(n notesTestData) bool { return n.read }),
	)

	_, toks := e.createAndSingIn(e.uuid()+"@test.com", "password")

	// create notes
	for _, ni := range notesInfo {
		httpResp := e.httpRequest(
			http.MethodPost,
			"/api/v1/note",
			e.jsonify(apiv1NoteCreateRequest{ //nolint:exhaustruct
				Content: ni.content,
				Slug:    ni.slug,
			}),
			toks.AccessToken)

		e.Equal(http.StatusCreated, httpResp.Code)
	}

	// read notes
	for _, ni := range notesInfo {
		if ni.read {
			httpResp := e.httpRequest(http.MethodGet, "/api/v1/note/"+ni.slug, nil)
			e.Equal(http.StatusOK, httpResp.Code)
		}
	}

	httpResp := e.httpRequest(http.MethodGet, "/api/v1/note/unread", nil, toks.AccessToken)

	var body []apiv1NoteGetAllResponse
	e.readBodyAndUnjsonify(httpResp.Body, &body)

	e.Equal(http.StatusOK, httpResp.Code)
	e.Len(body, unreadNotesTotal)
}
