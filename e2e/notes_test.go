package e2e

import "net/http"

func (s *AppTestSuite) TestNote_Create_RandomSlug() {
	resp := s.request("POST", "/api/v1/note", s.jsonify(map[string]any{
		"content": "testing",
	}))

	s.Require().Equal(http.StatusCreated, resp.Code)
}
