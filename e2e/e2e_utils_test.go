package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gofrs/uuid/v5"
)

// jsonify marshalls v into json and returns it as []byte
func (e *AppTestSuite) jsonify(v any) []byte {
	r, err := json.Marshal(v)
	e.require.NoError(err)
	return r
}

// readBodyAndUnjsonify reads body of `httptest.ResponseRecorder` and unmarshalls it into res
//
// Example:
//
//	var res struct { message string `json:"message"` }
//	readBodyAndUnjsonify(httpResp.Body, &res)
func (e *AppTestSuite) readBodyAndUnjsonify(b *bytes.Buffer, res any) {
	respData, err := io.ReadAll(b)
	e.require.NoError(err)

	err = json.Unmarshal(respData, &res)
	e.require.NoError(err)
}

func (e *AppTestSuite) httpRequest(method, url string, body []byte) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	e.require.NoError(err)

	req.Header.Set("Content-type", "application/json")

	resp := httptest.NewRecorder()
	e.router.ServeHTTP(resp, req)

	return resp
}

func (e *AppTestSuite) uuid() string {
	u, err := uuid.NewV4()
	e.require.NoError(err)
	return u.String()
}
