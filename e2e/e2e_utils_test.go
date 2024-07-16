package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gofrs/uuid/v5"
	"github.com/olexsmir/onasty/internal/jwtutil"
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

// httpRequest sends http request to the server and returns `httptest.ResponseRecorder`
// conteny-type always set to application/json
func (e *AppTestSuite) httpRequest(
	method, url string, //nolint:unparam // TODO: fix me later
	body []byte,
	accessToken ...string,
) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	e.require.NoError(err)

	req.Header.Set("Content-type", "application/json")

	if len(accessToken) == 1 {
		req.Header.Set("Authorization", "Bearer "+accessToken[0])
	}

	resp := httptest.NewRecorder()
	e.router.ServeHTTP(resp, req)

	return resp
}

// uuid generates a new UUID and returns it as a string
func (e *AppTestSuite) uuid() string {
	u, err := uuid.NewV4()
	e.require.NoError(err)
	return u.String()
}

// parseJwtToken util func that parses jwt token and returns payload
func (e *AppTestSuite) parseJwtToken(t string) jwtutil.Payload {
	r, err := e.jwtTokenizer.Parse(t)
	e.require.NoError(err)
	return r
}
