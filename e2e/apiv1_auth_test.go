package e2e

import "net/http"

type apiv1AuthSignUpRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (t *AppTestSuite) TestAuthV1_SignUP() {
	username := "test"
	email := "test@test.com"
	password := "password"

	httpResp := t.httpRequest(
		http.MethodPost,
		"/api/v1/auth/signup",
		t.jsonify(apiv1AuthSignUpRequest{
			Username: username,
			Email:    email,
			Password: password,
		}),
	)

	dbUser := t.getUserFromDBByUsername(username)
	hashedPasswd, err := t.hasher.Hash(password)
	t.require.NoError(err)

	t.Equal(http.StatusCreated, httpResp.Code)
	t.Equal(dbUser.Email, email)
	t.Equal(dbUser.Password, hashedPasswd)
}
