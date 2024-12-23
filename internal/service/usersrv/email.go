package usersrv

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

var ErrFailedToSendVerifcationEmail = errors.New("failed to send verification email")

const (
	verificationEmailSubject = "Onasty: verify your email"
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
	url string,
) {
	select {
	case <-ctx.Done():
		slog.ErrorContext(ctx, "failed to send verification email", "err", ctx.Err())
	default:
		if err := u.mailer.Send(
			ctx,
			userEmail,
			verificationEmailSubject,
			fmt.Sprintf(verificationEmailBody, url, token),
		); err != nil {
			slog.ErrorContext(ctx, "failed to send verification email", "err", err)
		}
		cancel()
	}
}
