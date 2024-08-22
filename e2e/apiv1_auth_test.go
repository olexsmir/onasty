package e2e

import (
	"net/http"

	"github.com/gofrs/uuid/v5"
)

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

type (
	apiv1AuthSignInRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	apiv1AuthSignInResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
)

func (e *AppTestSuite) TestAuthV1_VerifyEmail() {
	email := e.uuid() + "email@email.com"
	password := "qwerty"

	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/signup",
		e.jsonify(apiv1AuthSignUpRequest{
			Username: e.uuid(),
			Email:    email,
			Password: password,
		}),
	)

	e.Equal(http.StatusCreated, httpResp.Code)

	user := e.getLastInsertedUserByEmail(email)
	token := e.getVerificationTokenByUserID(user.ID)
	httpResp = e.httpRequest(http.MethodGet, "/api/v1/auth/verify/"+token.Token, nil)
	e.Equal(http.StatusOK, httpResp.Code)

	user = e.getLastInsertedUserByEmail(email)
	e.Equal(user.Activated, true)
}

func (e *AppTestSuite) TestAuthV1_ResendVerificationEmail() {
	e.T().Skip("implement me daddy")
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

	session := e.getLastUserSessionByUserID(uid)
	parsedToken := e.parseJwtToken(body.AccessToken)

	e.Equal(http.StatusOK, httpResp.Code)
	e.Equal(body.RefreshToken, session.RefreshToken)
	e.Equal(parsedToken.UserID, uid.String())
}

func (e *AppTestSuite) TestAuthV1_SignIn_wrong() {
	password := "password"
	email := e.uuid() + "@test.com"
	e.insertUserIntoDB(e.uuid(), email, "password")

	tests := []struct {
		name     string
		email    string
		password string
	}{
		// TODO: unactivated user
		{
			name:     "wrong email",
			email:    "wrong@emai.com",
			password: password,
		},
		{
			name:     "wrong password",
			email:    email,
			password: "wrong-wrong",
		},
	}

	for _, t := range tests {
		httpResp := e.httpRequest(
			http.MethodPost,
			"/api/v1/auth/signin",
			e.jsonify(apiv1AuthSignInRequest{
				Email:    t.email,
				Password: t.password,
			}),
		)

		e.Equal(http.StatusUnauthorized, httpResp.Code)
	}
}

type apiv1AuthRefreshTokensRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (e *AppTestSuite) TestAuthV1_RefreshTokens() {
	uid, toks := e.createAndSingIn(e.uuid()+"@test.com", e.uuid(), "password")
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/refresh-tokens",
		e.jsonify(apiv1AuthRefreshTokensRequest{
			RefreshToken: toks.RefreshToken,
		}),
	)

	var body apiv1AuthSignInResponse
	e.readBodyAndUnjsonify(httpResp.Body, &body)

	sessionDB := e.getLastUserSessionByUserID(uid)
	e.Equal(e.parseJwtToken(body.AccessToken).UserID, uid.String())

	e.Equal(httpResp.Code, http.StatusOK)
	e.NotEqual(toks.RefreshToken, body.RefreshToken)
	e.Equal(body.RefreshToken, sessionDB.RefreshToken)
}

func (e *AppTestSuite) TestAuthV1_RefreshTokens_wrong() {
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/refresh-tokens",
		e.jsonify(apiv1AuthRefreshTokensRequest{
			RefreshToken: e.uuid(),
		}),
	)

	e.Equal(httpResp.Code, http.StatusBadRequest)
}

func (e *AppTestSuite) TestAuthV1_Logout() {
	uid, toks := e.createAndSingIn(e.uuid()+"@test.com", e.uuid(), "password")

	sessionDB := e.getLastUserSessionByUserID(uid)
	e.NotEmpty(sessionDB.RefreshToken)

	httpResp := e.httpRequest(http.MethodPost, "/api/v1/auth/logout", nil, toks.AccessToken)
	e.Equal(httpResp.Code, http.StatusNoContent)

	sessionDB = e.getLastUserSessionByUserID(uid)
	e.Empty(sessionDB.RefreshToken)
}

type apiv1AtuhChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (e *AppTestSuite) TestAuthV1_ChangePassword() {
	password := e.uuid()
	newPassword := e.uuid()
	_, toks := e.createAndSingIn(e.uuid()+"@test.com", e.uuid(), password)

	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/change-password",
		e.jsonify(apiv1AtuhChangePasswordRequest{
			CurrentPassword: password,
			NewPassword:     newPassword,
		}),
		toks.AccessToken,
	)

	e.Equal(httpResp.Code, http.StatusOK)

	// TODO: check if password has been changed in db
}

func (e *AppTestSuite) createAndSingIn(
	email, username, password string,
) (uuid.UUID, apiv1AuthSignInResponse) {
	uid := e.insertUserIntoDB(username, email, password)
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/signin",
		e.jsonify(apiv1AuthSignInRequest{
			Email:    email,
			Password: password,
		}),
	)

	e.Equal(httpResp.Code, http.StatusOK)

	var body apiv1AuthSignInResponse
	e.readBodyAndUnjsonify(httpResp.Body, &body)

	return uid, body
}
