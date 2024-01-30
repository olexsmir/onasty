package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

func (s *AppTestSuite) jsonify(v map[string]any) []byte {
	r, _ := json.Marshal(v)
	return (r)
}

func (s *AppTestSuite) request(method, url string, body []byte) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(body))
	req.Header.Set("Content-type", "application/json")

	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	return resp
}
