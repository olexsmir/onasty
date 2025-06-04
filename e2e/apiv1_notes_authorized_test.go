package e2e_test

import (
	"net/http"
	"time"
)

func (e *AppTestSuite) TestNoteV1_Create_authorized() {
	uid, toks := e.createAndSingIn(e.uuid()+"@test.com", "password")
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/note",
		e.jsonify(apiv1NoteCreateRequest{ //nolint:exhaustruct
			Content: "some random ass content for the test",
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
			Content: "some random ass content for the test",
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
	BurnBeforeExpiration bool      `json:"burn_before_expiration"`
}

func (e *AppTestSuite) TestNoteV1_updateExpirationTime() {
	_, toks := e.createAndSingIn(e.uuid()+"@test.com", "password")
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/note",
		e.jsonify(apiv1NoteCreateRequest{ //nolint:exhaustruct
			Content:              "some random ass content for the test",
			ExpiresAt:            time.Now().Add(time.Minute),
			BurnBeforeExpiration: false,
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
			BurnBeforeExpiration: true,
		}),
		toks.AccessToken,
	)

	e.Equal(httpResp.Code, http.StatusOK)

	dbNote := e.getNoteBySlug(body.Slug)
	e.Equal(true, dbNote.BurnBeforeExpiration)
	e.Equal(patchTime.Unix(), dbNote.ExpiresAt.Unix())
}

func (e *AppTestSuite) TestNoteV1_updateExpirationTime_notFound() {
	_, toks := e.createAndSingIn(e.uuid()+"@test.com", "password")
	httpResp := e.httpRequest(
		http.MethodPatch,
		"/api/v1/note/"+e.uuid(),
		e.jsonify(apiV1NotePatchRequest{
			ExpiresAt:            time.Now().Add(time.Hour),
			BurnBeforeExpiration: true,
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

func (e *AppTestSuite) TestNoteV1_SetPassword_not_found() {
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
