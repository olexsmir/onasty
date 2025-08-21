package usersrv

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/events/mailermq"
	"github.com/olexsmir/onasty/internal/hasher"
	"github.com/olexsmir/onasty/internal/jwtutil"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/oauth"
	"github.com/olexsmir/onasty/internal/store/psql/changeemailrepo"
	"github.com/olexsmir/onasty/internal/store/psql/noterepo"
	"github.com/olexsmir/onasty/internal/store/psql/passwordtokrepo"
	"github.com/olexsmir/onasty/internal/store/psql/sessionrepo"
	"github.com/olexsmir/onasty/internal/store/psql/userepo"
	"github.com/olexsmir/onasty/internal/store/psql/vertokrepo"
	"github.com/olexsmir/onasty/internal/store/rdb/usercache"
)

type UserServicer interface {
	// SignUp creates a new user and sends verification email.
	SignUp(ctx context.Context, inp dtos.SignUp) (uuid.UUID, error)

	// SignIn authenticates a user and returns access and refresh tokens.
	SignIn(ctx context.Context, inp dtos.SignIn) (dtos.Tokens, error)

	// RefreshTokens refreshes the access and refresh tokens using the provided refresh token.
	RefreshTokens(ctx context.Context, refreshToken string) (dtos.Tokens, error)

	// Logout logs out a user by deleting the session associated with the provided refresh token.
	Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error

	// LogoutAll logs out a user by deleting all sessions associated with the user ID.
	LogoutAll(ctx context.Context, userID uuid.UUID) error

	// GetUserInfo retrieves user information by user ID.
	GetUserInfo(ctx context.Context, userID uuid.UUID) (dtos.UserInfo, error)

	// ChangePassword changes the user's password.
	ChangePassword(ctx context.Context, userID uuid.UUID, inp dtos.ChangeUserPassword) error

	// RequestPasswordReset initiates a password reset process by sending a reset email.
	RequestPasswordReset(ctx context.Context, inp dtos.RequestResetPassword) error

	// ResetPassword resets the user's password using the provided reset token.
	ResetPassword(ctx context.Context, inp dtos.ResetPassword) error

	RequestEmailChange(ctx context.Context, userID uuid.UUID, inp dtos.ChangeEmail) error

	ChangeEmail(ctx context.Context, token string) error

	// GetOAuthURL retrieves the OAuth URL for the specified provider.
	GetOAuthURL(providerName string) (dtos.OAuthRedirect, error)

	// HandleOAuthLogin handles the OAuth login process by exchanging the code for tokens.
	HandleOAuthLogin(ctx context.Context, providerName, code string) (dtos.Tokens, error)

	// Verify verifies the user's email using the provided verification key.
	Verify(ctx context.Context, verificationKey string) error

	// ResendVerificationEmail resends the verification email to the user.
	ResendVerificationEmail(ctx context.Context, inp dtos.ResendVerificationEmail) error

	// ParseJWTToken parses the JWT token and returns the payload.
	ParseJWTToken(token string) (jwtutil.Payload, error)

	// CheckIfUserExists checks if a user exists by user ID.
	CheckIfUserExists(ctx context.Context, userID uuid.UUID) (bool, error)

	// CheckIfUserIsActivated checks if a user is activated by user ID.
	CheckIfUserIsActivated(ctx context.Context, userID uuid.UUID) (bool, error)
}

var _ UserServicer = (*UserSrv)(nil)

type UserSrv struct {
	userstore       userepo.UserStorer
	sessionstore    sessionrepo.SessionStorer
	vertokrepo      vertokrepo.VerificationTokenStorer
	pwdtokrepo      passwordtokrepo.PasswordResetTokenStorer
	changeemailrepo changeemailrepo.ChangeEmailStorer
	notestore       noterepo.NoteStorer
	cache           usercache.UserCacheer

	hasher       hasher.Hasher
	jwtTokenizer jwtutil.JWTTokenizer
	mailermq     mailermq.Mailer

	googleOauth oauth.Provider
	githubOauth oauth.Provider

	refreshTokenTTL       time.Duration
	verificationTokenTTL  time.Duration
	resetPasswordTokenTTL time.Duration
	changeEmailTokenTTL   time.Duration
}

func New(
	userstore userepo.UserStorer,
	sessionstore sessionrepo.SessionStorer,
	vertokrepo vertokrepo.VerificationTokenStorer,
	pwdtokrepo passwordtokrepo.PasswordResetTokenStorer,
	changeemailrepo changeemailrepo.ChangeEmailStorer,
	notestore noterepo.NoteStorer,
	hasher hasher.Hasher,
	jwtTokenizer jwtutil.JWTTokenizer,
	mailermq mailermq.Mailer,
	cache usercache.UserCacheer,
	googleOauth, githubOauth oauth.Provider,
	refreshTokenTTL, verificationTokenTTL, resetPasswordTokenTTL, changeEmailTokenTTL time.Duration,
) *UserSrv {
	return &UserSrv{
		userstore:             userstore,
		sessionstore:          sessionstore,
		vertokrepo:            vertokrepo,
		pwdtokrepo:            pwdtokrepo,
		changeemailrepo:       changeemailrepo,
		notestore:             notestore,
		cache:                 cache,
		hasher:                hasher,
		jwtTokenizer:          jwtTokenizer,
		mailermq:              mailermq,
		googleOauth:           googleOauth,
		githubOauth:           githubOauth,
		refreshTokenTTL:       refreshTokenTTL,
		verificationTokenTTL:  verificationTokenTTL,
		resetPasswordTokenTTL: resetPasswordTokenTTL,
		changeEmailTokenTTL:   changeEmailTokenTTL,
	}
}

func (u *UserSrv) SignUp(ctx context.Context, inp dtos.SignUp) (uuid.UUID, error) {
	user := models.User{
		ID:          uuid.Nil, // nil, because it does not get used here
		Email:       inp.Email,
		Activated:   false,
		Password:    inp.Password,
		CreatedAt:   inp.CreatedAt,
		LastLoginAt: inp.LastLoginAt,
	}
	if err := user.Validate(); err != nil {
		return uuid.Nil, err
	}

	hashedPassword, err := u.hasher.Hash(inp.Password)
	if err != nil {
		return uuid.UUID{}, err
	}

	user.Password = hashedPassword

	userID, err := u.userstore.Create(ctx, user)
	if err != nil {
		return uuid.Nil, err
	}

	verificationToken := uuid.Must(uuid.NewV4()).String()
	if err := u.vertokrepo.Create(ctx, models.VerificationToken{
		UserID:    userID,
		Token:     verificationToken,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(u.verificationTokenTTL),
	}); err != nil {
		return uuid.Nil, err
	}

	if err := u.mailermq.SendVerificationEmail(ctx, mailermq.SendVerificationEmailRequest{
		Receiver: inp.Email,
		Token:    verificationToken,
	}); err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func (u *UserSrv) SignIn(ctx context.Context, inp dtos.SignIn) (dtos.Tokens, error) {
	user, err := u.userstore.GetByEmail(ctx, inp.Email)
	if err != nil {
		return dtos.Tokens{}, err
	}

	if err = u.hasher.Compare(user.Password, inp.Password); err != nil {
		if errors.Is(err, hasher.ErrMismatchedHashes) {
			return dtos.Tokens{}, models.ErrUserWrongCredentials
		}
		return dtos.Tokens{}, err
	}

	if !user.IsActivated() {
		return dtos.Tokens{}, models.ErrUserIsNotActivated
	}

	tokens, err := u.issueTokens(ctx, user.ID)
	return tokens, err
}

func (u *UserSrv) Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	return u.sessionstore.Delete(ctx, userID, refreshToken)
}

func (u *UserSrv) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return u.sessionstore.DeleteAllByUserID(ctx, userID)
}

func (u *UserSrv) GetUserInfo(ctx context.Context, userID uuid.UUID) (dtos.UserInfo, error) {
	user, err := u.userstore.GetByID(ctx, userID)
	if err != nil {
		return dtos.UserInfo{}, err
	}

	count, err := u.notestore.GetCountOfNotesByAuthorID(ctx, userID)
	if err != nil {
		return dtos.UserInfo{}, err
	}

	return dtos.UserInfo{
		Email:        user.Email,
		CreatedAt:    user.CreatedAt,
		LastLoginAt:  user.LastLoginAt,
		NotesCreated: int(count),
	}, nil
}

func (u *UserSrv) RefreshTokens(ctx context.Context, rtoken string) (dtos.Tokens, error) {
	userID, err := u.sessionstore.GetUserIDByRefreshToken(ctx, rtoken)
	if err != nil {
		return dtos.Tokens{}, err
	}

	tokens, err := u.createTokens(userID)
	if err != nil {
		return dtos.Tokens{}, err
	}

	if err := u.sessionstore.Update(ctx, userID, rtoken, tokens.Refresh); err != nil {
		return dtos.Tokens{}, err
	}

	return dtos.Tokens{
		Access:  tokens.Access,
		Refresh: tokens.Refresh,
	}, nil
}

func (u *UserSrv) ChangePassword(
	ctx context.Context,
	userID uuid.UUID,
	inp dtos.ChangeUserPassword,
) error {
	//nolint:exhaustruct
	if err := (models.User{Password: inp.NewPassword}).ValidatePassword(); err != nil {
		return err
	}

	user, err := u.userstore.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err = u.hasher.Compare(user.Password, inp.CurrentPassword); err != nil {
		return errors.Join(err, models.ErrUserInvalidPassword)
	}

	newPass, err := u.hasher.Hash(inp.NewPassword)
	if err != nil {
		return err
	}

	if err := u.userstore.ChangePassword(ctx, userID, newPass); err != nil {
		return err
	}

	return nil
}

func (u *UserSrv) RequestPasswordReset(ctx context.Context, inp dtos.RequestResetPassword) error {
	user, err := u.userstore.GetByEmail(ctx, inp.Email)
	if err != nil {
		return err
	}

	token := uuid.Must(uuid.NewV4()).String()
	if err := u.pwdtokrepo.Create(ctx, models.ResetPasswordToken{
		UserID:    user.ID,
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(u.resetPasswordTokenTTL),
	}); err != nil {
		return err
	}

	if err := u.mailermq.SendPasswordResetEmail(ctx, mailermq.SendPasswordResetEmailRequest{
		Receiver: inp.Email,
		Token:    token,
	}); err != nil {
		return err
	}

	return nil
}

func (u *UserSrv) ResetPassword(ctx context.Context, inp dtos.ResetPassword) error {
	//nolint:exhaustruct
	if err := (models.User{Password: inp.NewPassword}).ValidatePassword(); err != nil {
		return err
	}

	uid, err := u.pwdtokrepo.GetUserIDByTokenAndMarkAsUsed(ctx, inp.Token, time.Now())
	if err != nil {
		return err
	}

	hashedPassword, err := u.hasher.Hash(inp.NewPassword)
	if err != nil {
		return err
	}

	return u.userstore.SetPassword(ctx, uid, hashedPassword)
}

func (u *UserSrv) RequestEmailChange(
	ctx context.Context,
	userID uuid.UUID,
	inp dtos.ChangeEmail,
) error {
	user, err := u.userstore.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.Email == inp.NewEmail {
		return models.ErrUserEmailIsAlreadyInUse
	}

	token := uuid.Must(uuid.NewV4()).String()
	changeEmailInput := models.ChangeEmailToken{
		UserID:    userID,
		Token:     token,
		NewEmail:  inp.NewEmail,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(u.changeEmailTokenTTL),
	}
	if err := changeEmailInput.Validate(); err != nil {
		return err
	}

	if err := u.changeemailrepo.Create(ctx, changeEmailInput); err != nil {
		return err
	}

	if err := u.mailermq.SendChangeEmailConfirmation(ctx, mailermq.SendChangeEmailConfirmationRequest{
		Receiver: user.Email,
		Token:    token,
		NewEmail: inp.NewEmail,
	}); err != nil {
		return err
	}

	return nil
}

func (u *UserSrv) ChangeEmail(ctx context.Context, givenToken string) error {
	token, err := u.changeemailrepo.GetByToken(ctx, givenToken)
	if err != nil {
		return err
	}

	user, err := u.userstore.GetByID(ctx, token.UserID)
	if err != nil {
		return err
	}

	if user.Email == token.NewEmail {
		return models.ErrUserEmailIsAlreadyInUse
	}

	if err := u.userstore.SetEmail(ctx, token.UserID, token.NewEmail); err != nil {
		return err
	}

	if err := u.changeemailrepo.MarkAsUsed(ctx, token.Token, time.Now()); err != nil {
		return err
	}

	return nil
}

func (u *UserSrv) Verify(ctx context.Context, verificationKey string) error {
	uid, err := u.vertokrepo.GetUserIDByTokenAndMarkAsUsed(ctx, verificationKey, time.Now())
	if err != nil {
		return err
	}

	return u.userstore.MarkUserAsActivated(ctx, uid)
}

func (u *UserSrv) ResendVerificationEmail(
	ctx context.Context,
	inp dtos.ResendVerificationEmail,
) error {
	user, err := u.userstore.GetByEmail(ctx, inp.Email)
	if err != nil {
		return err
	}

	if user.Activated {
		return models.ErrUserIsAlreadyVerified
	}

	token, err := u.vertokrepo.GetTokenOrUpdateTokenByUserID(
		ctx,
		user.ID,
		uuid.Must(uuid.NewV4()).String(),
		time.Now().Add(u.verificationTokenTTL))
	if err != nil {
		return err
	}

	if err := u.mailermq.SendVerificationEmail(ctx, mailermq.SendVerificationEmailRequest{
		Receiver: inp.Email,
		Token:    token,
	}); err != nil {
		return err
	}

	return nil
}

func (u *UserSrv) ParseJWTToken(token string) (jwtutil.Payload, error) {
	return u.jwtTokenizer.Parse(token)
}

func (u UserSrv) CheckIfUserExists(ctx context.Context, id uuid.UUID) (bool, error) {
	r, err := u.cache.GetIsExists(ctx, id.String())
	if err == nil {
		return r, nil
	}

	slog.ErrorContext(ctx, "usercache", "err", err)

	isExists, err := u.userstore.CheckIfUserExists(ctx, id)
	if err != nil {
		return false, err
	}

	if err := u.cache.SetIsExists(ctx, id.String(), isExists); err != nil {
		slog.ErrorContext(ctx, "usercache", "err", err)
	}

	return isExists, nil
}

func (u *UserSrv) CheckIfUserIsActivated(ctx context.Context, userID uuid.UUID) (bool, error) {
	r, err := u.cache.GetIsActivated(ctx, userID.String())
	if err == nil {
		return r, nil
	}

	slog.ErrorContext(ctx, "usercache", "err", err)

	isActivated, err := u.userstore.CheckIfUserIsActivated(ctx, userID)
	if err != nil {
		return false, err
	}

	if err := u.cache.SetIsActivated(ctx, userID.String(), isActivated); err != nil {
		slog.ErrorContext(ctx, "usercache", "err", err)
	}

	return isActivated, nil
}

func (u UserSrv) createTokens(userID uuid.UUID) (dtos.Tokens, error) {
	accessToken, err := u.jwtTokenizer.AccessToken(jwtutil.Payload{UserID: userID.String()})
	if err != nil {
		return dtos.Tokens{}, err
	}

	refreshToken, err := u.jwtTokenizer.RefreshToken()
	if err != nil {
		return dtos.Tokens{}, err
	}

	return dtos.Tokens{
		Access:  accessToken,
		Refresh: refreshToken,
	}, err
}

func (u UserSrv) issueTokens(ctx context.Context, userID uuid.UUID) (dtos.Tokens, error) {
	toks, err := u.createTokens(userID)
	if err != nil {
		return dtos.Tokens{}, err
	}

	if err := u.sessionstore.Set(ctx, userID, toks.Refresh, time.Now().Add(u.refreshTokenTTL)); err != nil {
		return dtos.Tokens{}, err
	}

	return toks, nil
}
