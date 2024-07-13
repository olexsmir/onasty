package e2e

import "net/http"

type apiv1AuthSignUpRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (e *AppTestSuite) TestAuthV1_SignUP() {
	username := "test" + e.uuid()
	email := e.uuid() + "test@test.com"
	password := "password"

	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/signup",
		e.jsonify(apiv1AuthSignUpRequest{
			Username: username,
			Email:    email,
			Password: password,
		}),
	)

	dbUser := e.getUserFromDBByUsername(username)
	hashedPasswd, err := e.hasher.Hash(password)
	e.require.NoError(err)

	e.Equal(http.StatusCreated, httpResp.Code)
	e.Equal(dbUser.Email, email)
	e.Equal(dbUser.Password, hashedPasswd)
}

func (e *AppTestSuite) TestAuthV1_SignUP_badrequest() {
	tests := []struct {
		name     string
		username string
		email    string
		password string
	}{
		{name: "all fiels empty", email: "", password: "", username: ""},
		{
			name:     "non valid email",
			email:    "email",
			password: "password",
		},
		{
			name:     "non valid password",
			email:    "test@test.com",
			password: "12345",
			username: "test",
		},
	}
	for _, t := range tests {
		httpResp := e.httpRequest(
			http.MethodPost,
			"/api/v1/auth/signup",
			e.jsonify(apiv1AuthSignUpRequest{
				Username: t.username,
				Email:    t.email,
				Password: t.password,
			}),
		)

		e.Equal(http.StatusBadRequest, httpResp.Code)
	}
}

type apiv1AuthSignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type apiv1AuthSignInResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (e *AppTestSuite) TestAuthV1_SignIn() {
	email := e.uuid() + "email@email.com"
	password := "qwerty"

	uid := e.insertUserIntoDB("test", email, password)

	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/signin",
		e.jsonify(apiv1AuthSignInRequest{
			Email:    email,
			Password: password,
		}),
	)

	var body apiv1AuthSignInResponse
	e.readBodyAndUnjsonify(httpResp.Body, &body)

	session := e.getLastUserSessionByUserId(uid)
	parsedToken, err := e.jwtTokenizer.Parse(body.AccessToken)
	e.require.NoError(err)

	e.Equal(http.StatusOK, httpResp.Code)
	e.Equal(body.RefreshToken, session.RefreshToken)
	e.Equal(parsedToken.UserID, uid.String())
}
