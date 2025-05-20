package e2e_test

import "net/http"

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
