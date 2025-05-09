package e2e_test

import (
	"context"

	"github.com/olexsmir/onasty/internal/events/mailermq"
)

var _ mailermq.Mailer = (*mailerMockService)(nil)

type mailerMockService struct{}

func newMailerMockService() *mailerMockService {
	return &mailerMockService{}
}

func (m mailerMockService) SendVerificationEmail(
	_ context.Context,
	_ mailermq.SendVerificationEmailRequest,
) error {
	return nil
}

func (m mailerMockService) SendPasswordResetEmail(
	_ context.Context,
	_ mailermq.SendPasswordResetEmailRequest,
) error {
	return nil
}
