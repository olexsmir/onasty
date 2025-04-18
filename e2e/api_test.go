package e2e_test

import "net/http"

type apiPingResponse struct {
	Message string `json:"message"`
}

func (e *AppTestSuite) TestPing() {
	httpResp := e.httpRequest(http.MethodGet, "/api/ping", nil)

	var body apiPingResponse
	e.readBodyAndUnjsonify(httpResp.Body, &body)

	e.Equal(http.StatusOK, httpResp.Code)
	e.Equal(body.Message, "pong")
}
