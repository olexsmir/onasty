package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
)

// jsonify marshalls v into json and returns it as []byte
func (s *AppTestSuite) jsonify(v any) []byte {
	r, err := json.Marshal(v)
	s.require.NoError(err)
	return r
}

// readBodyAndUnjsonify reads body of `httptest.ResponseRecorder` and unmarshalls it into res
//
// Example:
//
//	var res struct { message string `json:"message"` }
//	readBodyAndUnjsonify(httpResp.Body, &res)
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
