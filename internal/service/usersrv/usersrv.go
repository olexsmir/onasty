package usersrv

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/events/mailermq"
	"github.com/olexsmir/onasty/internal/hasher"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/psql/changeemailrepo"
	"github.com/olexsmir/onasty/internal/store/psql/noterepo"
	"github.com/olexsmir/onasty/internal/store/psql/passwordtokrepo"
	"github.com/olexsmir/onasty/internal/store/psql/userepo"
	"github.com/olexsmir/onasty/internal/store/psql/vertokrepo"
)

type UserServicer interface {
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

	// Verify verifies the user's email using the provided verification key.
	Verify(ctx context.Context, verificationKey string) error

	// ResendVerificationEmail resends the verification email to the user.
	ResendVerificationEmail(ctx context.Context, inp dtos.ResendVerificationEmail) error
}

var _ UserServicer = (*UserSrv)(nil)

type UserSrv struct {
	userstore       userepo.UserStorer
	vertokrepo      vertokrepo.VerificationTokenStorer
	pwdtokrepo      passwordtokrepo.PasswordResetTokenStorer
	changeemailrepo changeemailrepo.ChangeEmailStorer
	notestore       noterepo.NoteStorer

	hasher   hasher.Hasher
	mailermq mailermq.Mailer

	verificationTokenTTL  time.Duration
	resetPasswordTokenTTL time.Duration
	changeEmailTokenTTL   time.Duration
}

func New(
	userstore userepo.UserStorer,
	vertokrepo vertokrepo.VerificationTokenStorer,
	pwdtokrepo passwordtokrepo.PasswordResetTokenStorer,
	changeemailrepo changeemailrepo.ChangeEmailStorer,
	notestore noterepo.NoteStorer,
	hasher hasher.Hasher,
	mailermq mailermq.Mailer,
	verificationTokenTTL, resetPasswordTokenTTL, changeEmailTokenTTL time.Duration,
) *UserSrv {
	return &UserSrv{
		userstore:             userstore,
		vertokrepo:            vertokrepo,
		pwdtokrepo:            pwdtokrepo,
		changeemailrepo:       changeemailrepo,
		notestore:             notestore,
		hasher:                hasher,
		mailermq:              mailermq,
		verificationTokenTTL:  verificationTokenTTL,
		resetPasswordTokenTTL: resetPasswordTokenTTL,
		changeEmailTokenTTL:   changeEmailTokenTTL,
	}
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
		return models.ErrUserNotFound
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
