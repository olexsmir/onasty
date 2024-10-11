package e2e_test

import (
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/olexsmir/onasty/internal/models"
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
			username: "testing",
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
	email, password := e.uuid()+"email@email.com", e.uuid()

	// create test user
	signUpHTTPResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/signup",
		e.jsonify(apiv1AuthSignUpRequest{
			Username: e.uuid(),
			Email:    email,
			Password: password,
		}),
	)

	e.Equal(http.StatusCreated, signUpHTTPResp.Code)

	// handle sending of the email
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/resend-verification-email",
		e.jsonify(apiv1AuthSignInRequest{
			Email:    email,
			Password: password,
		}),
	)

	e.Equal(http.StatusOK, httpResp.Code)
	e.NotEmpty(e.mailer.GetLastSentEmailToEmail(email))
}

func (e *AppTestSuite) TestAuthV1_ResendVerificationEmail_wrong() {
	email, password := e.uuid()+"@"+e.uuid()+".com", "password"
	e.insertUserIntoDB(e.uuid(), email, password, true)

	tests := []struct {
		name         string
		email        string
		password     string
		expectedCode int
	}{
		{
			name:         "activated account",
			email:        email,
			password:     password,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "wrong credintials",
			email:        email,
			password:     e.uuid(),
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, t := range tests {
		httpResp := e.httpRequest(
			http.MethodPost,
			"/api/v1/auth/resend-verification-email",
			e.jsonify(apiv1AuthSignInRequest{
				Email:    t.email,
				Password: t.password,
			}))

		e.Equal(httpResp.Code, t.expectedCode)

		// no email should be sent
		e.Empty(e.mailer.GetLastSentEmailToEmail(t.email))
	}
}

func (e *AppTestSuite) TestAuthV1_SignIn() {
	email := e.uuid() + "email@email.com"
	password := "qwerty"

	uid := e.insertUserIntoDB("test", email, password, true)

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
	e.insertUserIntoDB(e.uuid(), email, "password", true)

	unactivatedEmail := e.uuid() + "@test.com"
	e.insertUserIntoDB(e.uuid(), unactivatedEmail, password, false)

	//exhaustruct:ignore
	tests := []struct {
		name         string
		email        string
		password     string
		expectedCode int

		expectMsg   bool
		expectedMsg string
	}{
		{
			name:         "unactivated user",
			email:        unactivatedEmail,
			password:     password,
			expectedCode: http.StatusBadRequest,
			expectMsg:    true,
			expectedMsg:  models.ErrUserIsNotActivated.Error(),
		},
		{
			name:         "wrong email",
			email:        "wrong@emai.com",
			password:     password,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "wrong password",
			email:        email,
			password:     "wrong-wrong",
			expectedCode: http.StatusUnauthorized,
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

		if t.expectMsg {
			var body errorResponse
			e.readBodyAndUnjsonify(httpResp.Body, &body)

			e.Equal(body.Message, t.expectedMsg)
		}

		e.Equal(t.expectedCode, httpResp.Code)
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
	// requests a new token pair with a wrong refresh token

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
	username := e.uuid()
	_, toks := e.createAndSingIn(e.uuid()+"@test.com", username, password)

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

	userDB := e.getUserFromDBByUsername(username)
	hashedNewPassword, err := e.hasher.Hash(newPassword)
	e.require.NoError(err)

	e.Equal(userDB.Password, hashedNewPassword)
}

func (e *AppTestSuite) createAndSingIn(
	email, username, password string,
) (uuid.UUID, apiv1AuthSignInResponse) {
	uid := e.insertUserIntoDB(username, email, password, true)
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
