package usersrv

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

var ErrFailedToSendVerifcationEmail = errors.New("failed to send verification email")

const (
	verificationEmailSubject = "Onasty: verifiy your email"
	verificationEmailBody    = `To verify your email, please follow this link:
<a href="%[1]s/api/v1/auth/verify/%[2]s">%[1]s/api/v1/auth/verify/%[2]s</a>
<br />
<br />
This link will expire after 24 hours.`
)

func (u *UserSrv) sendVerificationEmail(
	ctx context.Context,
	cancel context.CancelFunc,
	userEmail string,
	token string,
) error {
	select {
	case <-ctx.Done():
		slog.Error("failed to send verfication email", "err", ctx.Err())
		return ErrFailedToSendVerifcationEmail
	default:
		if err := u.mailer.Send(
			ctx,
			userEmail,
			verificationEmailSubject,
			// TODO: set proper url
			fmt.Sprintf(verificationEmailBody, "http://localhost:3000", token),
		); err != nil {
			return errors.Join(ErrFailedToSendVerifcationEmail, err)
		}
		cancel()

		slog.Debug("email sent")
	}

	return nil
}
