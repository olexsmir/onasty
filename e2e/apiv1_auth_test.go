package e2e_test

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/olexsmir/onasty/internal/hasher"
	"github.com/olexsmir/onasty/internal/models"
)

type apiv1AuthSignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (e *AppTestSuite) TestAuthV1_SignUP() {
	email, password := e.randomEmail(), "password"

	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/signup",
		e.jsonify(apiv1AuthSignUpRequest{
			Email:    email,
			Password: password,
		}),
	)

	dbUser := e.getUserByEmail(email)
	hashedPasswd, err := e.hasher.Hash(password)
	e.require.NoError(err)

	e.Equal(http.StatusCreated, httpResp.Code)
	e.Equal(dbUser.Email, email)
	e.Equal(dbUser.Password, hashedPasswd)
}

func (e *AppTestSuite) TestAuthV1_SignUP_badrequest() {
	tests := []struct {
		name     string
		email    string
		password string
	}{
		{name: "all fields empty", email: "", password: ""},
		{name: "non valid email", email: "email", password: "password"},
		{name: "non valid password", email: e.randomEmail(), password: "12345"},
	}
	for _, t := range tests {
		httpResp := e.httpRequest(
			http.MethodPost,
			"/api/v1/auth/signup",
			e.jsonify(apiv1AuthSignUpRequest{
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
			Email:    email,
			Password: password,
		}),
	)

	e.Equal(http.StatusCreated, httpResp.Code)

	user := e.getLastUserByEmail(email)
	token := e.getVerificationTokenByUserID(user.ID)
	e.Equal(token.Token, mockMailStore[email])

	httpResp = e.httpRequest(http.MethodGet, "/api/v1/auth/verify/"+token.Token, nil)
	e.Equal(http.StatusOK, httpResp.Code)

	user = e.getLastUserByEmail(email)
	e.Equal(user.Activated, true)
}

type apiv1AuthResendVerificationEmailRequest struct {
	Email string `json:"email"`
}

func (e *AppTestSuite) TestAuthV1_ResendVerificationEmail() {
	email, password := e.uuid()+"email@email.com", e.uuid()

	// create test user
	signUpHTTPResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/signup",
		e.jsonify(apiv1AuthSignUpRequest{
			Email:    email,
			Password: password,
		}),
	)

	e.Equal(http.StatusCreated, signUpHTTPResp.Code)

	// handle sending of the email
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/resend-verification-email",
		e.jsonify(apiv1AuthResendVerificationEmailRequest{
			Email: email,
		}),
	)

	e.Equal(http.StatusOK, httpResp.Code)
	e.NotEmpty(mockMailStore[email])
}

func (e *AppTestSuite) TestAuthV1_ResendVerificationEmail_wrong() {
	email, password := e.uuid()+"@"+e.uuid()+".com", "password"
	e.insertUser(email, password, true)

	tests := []struct {
		name         string
		email        string
		expectedCode int
		expectedMsg  string
	}{
		{
			name:         "already verified account",
			email:        email,
			expectedCode: http.StatusBadRequest,
			expectedMsg:  models.ErrUserIsAlreadyVerified.Error(),
		},
		{
			name:         "user not found",
			email:        e.uuid() + "@at.com",
			expectedCode: http.StatusBadRequest,
			expectedMsg:  models.ErrUserNotFound.Error(),
		},
	}

	for _, t := range tests {
		httpResp := e.httpRequest(
			http.MethodPost,
			"/api/v1/auth/resend-verification-email",
			e.jsonify(apiv1AuthResendVerificationEmailRequest{
				Email: t.email,
			}))

		e.Equal(httpResp.Code, t.expectedCode)

		var body errorResponse
		e.readBodyAndUnjsonify(httpResp.Body, &body)
		e.Equal(body.Message, t.expectedMsg)

		e.Empty(mockMailStore[t.email])
	}
}

func (e *AppTestSuite) TestAuthV1_SignIn() {
	email := e.uuid() + "email@email.com"
	password := "qwerty"
	uid := e.insertUser(email, password, true)

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

	session := e.getLastSessionByUserID(uid)
	parsedToken := e.parseJwtToken(body.AccessToken)

	e.Equal(http.StatusOK, httpResp.Code)
	e.Equal(body.RefreshToken, session.RefreshToken)
	e.Equal(parsedToken.UserID, uid.String())
}

func (e *AppTestSuite) TestAuthV1_SignIn_wrong() {
	email, unactivatedEmail, password := e.randomEmail(), e.randomEmail(), e.uuid()

	e.insertUser(email, password, true)
	e.insertUser(unactivatedEmail, password, false)

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
			name:         "inactivated user",
			email:        unactivatedEmail,
			password:     password,
			expectedCode: http.StatusBadRequest,
			expectMsg:    true,
			expectedMsg:  models.ErrUserIsNotActivated.Error(),
		},
		{
			name:         "wrong email",
			email:        "wrong@email.com",
			password:     e.uuid(),
			expectedCode: http.StatusBadRequest,
			expectedMsg:  models.ErrUserWrongCredentials.Error(),
		},
		{
			name:         "wrong password",
			email:        email,
			password:     "wrong-wrong",
			expectedCode: http.StatusBadRequest,
			expectedMsg:  models.ErrUserWrongCredentials.Error(),
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
	uid, toks := e.createAndSingIn(e.randomEmail(), "password")
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/refresh-tokens",
		e.jsonify(apiv1AuthRefreshTokensRequest{
			RefreshToken: toks.RefreshToken,
		}),
	)

	var body apiv1AuthSignInResponse
	e.readBodyAndUnjsonify(httpResp.Body, &body)

	sessionDB := e.getLastSessionByUserID(uid)
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

type apiV1AuthLogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (e *AppTestSuite) TestAuthV1_Logout() {
	uid, toks := e.createAndSingIn(e.randomEmail(), "password")

	sessionDB := e.getLastSessionByUserID(uid)
	e.NotEmpty(sessionDB.RefreshToken)

	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/logout",
		e.jsonify(apiV1AuthLogoutRequest{
			RefreshToken: toks.RefreshToken,
		}),
		toks.AccessToken,
	)
	e.Equal(httpResp.Code, http.StatusNoContent)

	sessionDB = e.getLastSessionByUserID(uid)
	e.Empty(sessionDB.RefreshToken)
}

func (e *AppTestSuite) TestAuthV1_LogoutAll() {
	uid, toks := e.createAndSingIn(e.randomEmail(), "password")

	var res int
	query := "select count(*) from sessions where user_id = $1"

	err := e.postgresDB.QueryRow(e.ctx, query, uid).Scan(&res)
	e.require.NoError(err)
	e.NotZero(res)

	httpResp := e.httpRequest(http.MethodPost, "/api/v1/auth/logout/all", nil, toks.AccessToken)
	e.Equal(httpResp.Code, http.StatusNoContent)

	err = e.postgresDB.QueryRow(e.ctx, query, uid).Scan(&res)
	e.require.NoError(err)
	e.Zero(res)
}

type apiv1AuthChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (e *AppTestSuite) TestAuthV1_ChangePassword() {
	email, oldPassword, newPassword := e.randomEmail(), e.uuid(), e.uuid()
	_, toks := e.createAndSingIn(email, oldPassword)

	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/change-password",
		e.jsonify(apiv1AuthChangePasswordRequest{
			CurrentPassword: oldPassword,
			NewPassword:     newPassword,
		}),
		toks.AccessToken,
	)

	e.Equal(httpResp.Code, http.StatusOK)

	userDB := e.getUserByEmail(email)
	e.NoError(e.hasher.Compare(userDB.Password, newPassword))
}

func (e *AppTestSuite) TestAuthV1_ChangePassword_wrongPassword() {
	email, oldPassword, newPassword := e.randomEmail(), e.uuid(), e.uuid()
	_, toks := e.createAndSingIn(email, oldPassword)

	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/change-password",
		e.jsonify(apiv1AuthChangePasswordRequest{
			CurrentPassword: e.uuid(),
			NewPassword:     newPassword,
		}),
		toks.AccessToken,
	)

	e.Equal(http.StatusBadRequest, httpResp.Code)

	var body errorResponse
	e.readBodyAndUnjsonify(httpResp.Body, &body)
	e.Equal(models.ErrUserWrongCredentials.Error(), body.Message)

	userDB := e.getUserByEmail(email)

	err := e.hasher.Compare(userDB.Password, newPassword)
	e.ErrorIs(err, hasher.ErrMismatchedHashes)
}

type (
	apiV1AuthResetPasswordRequest struct {
		Email string `json:"email"`
	}
	apiV1AuthSetPasswordRequest struct {
		Password string `json:"password"`
	}
)

func (e *AppTestSuite) TestAuthV1_ResetPassword() {
	email := e.randomEmail()
	uid, _ := e.createAndSingIn(email, "password")

	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/reset-password",
		e.jsonify(apiV1AuthResetPasswordRequest{
			Email: email,
		}),
	)

	e.Equal(httpResp.Code, http.StatusOK)

	token := e.getResetPasswordTokenByUserID(uid)
	e.Empty(token.UsedAt)
	e.Equal(mockMailStore[email], token.Token)

	// set new password
	password := e.uuid()
	httpResp = e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/reset-password/"+token.Token,
		e.jsonify(apiV1AuthSetPasswordRequest{
			Password: password,
		}),
	)

	dbUser := e.getUserByEmail(email)
	e.Equal(httpResp.Code, http.StatusOK)
	e.NoError(e.hasher.Compare(dbUser.Password, password))

	token = e.getResetPasswordTokenByUserID(uid)
	e.NotEmpty(token.UsedAt)
}

func (e *AppTestSuite) TestAuthV1_ResetPassword_nonExistentUser() {
	_, _ = e.createAndSingIn(e.randomEmail(), "password")
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/reset-password",
		e.jsonify(apiV1AuthResetPasswordRequest{
			Email: e.uuid() + "@testing.com",
		}),
	)

	e.Equal(httpResp.Code, http.StatusBadRequest)
}

type apiv1AuthChangeEmailRequest struct {
	NewEmail string `json:"new_email"`
}

func (e *AppTestSuite) TestAuthV1_ChangeEmail() {
	oldEmail, newEmail := e.randomEmail(), e.randomEmail()
	uid, toks := e.createAndSingIn(oldEmail, e.uuid())

	// request email change
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/change-email",
		e.jsonify(apiv1AuthChangeEmailRequest{
			NewEmail: newEmail,
		}),
		toks.AccessToken,
	)
	e.Equal(http.StatusOK, httpResp.Code)

	token := e.getChangeEmailTokenByUserID(uid)
	e.Empty(token.UsedAt)
	e.Equal(mockMailStore[oldEmail], token.Token)

	// confirm email change
	httpResp = e.httpRequest(http.MethodGet, "/api/v1/auth/change-email/"+token.Token, nil)
	e.Equal(http.StatusOK, httpResp.Code)

	updatedToken := e.getChangeEmailTokenByUserID(uid)
	e.NotEmpty(updatedToken.UsedAt)

	dbUser := e.getUserByEmail(token.Extra)
	e.Equal(dbUser.Email, newEmail)
}

func (e *AppTestSuite) TestAuthV1_ChangeEmail_wrongSameEmail() {
	email := e.randomEmail()
	_, toks := e.createAndSingIn(email, e.uuid())

	// request email change
	httpResp := e.httpRequest(
		http.MethodPost,
		"/api/v1/auth/change-email",
		e.jsonify(apiv1AuthChangeEmailRequest{
			NewEmail: email,
		}),
		toks.AccessToken,
	)
	e.Equal(http.StatusBadRequest, httpResp.Code)

	var body errorResponse
	e.readBodyAndUnjsonify(httpResp.Body, &body)

	e.Equal(body.Message, models.ErrUserEmailIsAlreadyInUse.Error())
}

type getMeResponse struct {
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	LastLoginAt  time.Time `json:"last_login_at"`
	NotesCreated int       `json:"notes_created"`
}

func (e *AppTestSuite) TestApiV1_getMe() {
	email := e.randomEmail()
	uid, toks := e.createAndSingIn(email, "password")

	httpResp := e.httpRequest(http.MethodGet, "/api/v1/me", nil, toks.AccessToken)

	e.Equal(httpResp.Code, http.StatusOK)

	var body getMeResponse
	e.readBodyAndUnjsonify(httpResp.Body, &body)

	e.Equal(email, body.Email)
	e.NotZero(body.CreatedAt)
	e.NotZero(body.LastLoginAt)

	var notesCount int
	err := e.postgresDB.
		QueryRow(e.ctx, "select count(*) from notes_authors where user_id = $1", uid).
		Scan(&notesCount)
	e.require.NoError(err)

	e.Equal(body.NotesCreated, notesCount)
}

// createAndSingIn creates an activated user, logs them in,
// and returns their userID along with access and refresh tokens.
func (e *AppTestSuite) createAndSingIn(
	email, password string,
) (uuid.UUID, apiv1AuthSignInResponse) {
	uid := e.insertUser(email, password, true)
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

func (e *AppTestSuite) randomEmail() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("user-%s@test.local", hex.EncodeToString(b))
}
