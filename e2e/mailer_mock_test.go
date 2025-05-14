package e2e_test

import (
	"context"

	"github.com/olexsmir/onasty/internal/events/mailermq"
)

var _ mailermq.Mailer = (*mailerMockService)(nil)

var mockMailStore = make(map[string]string)

type mailerMockService struct{}

func newMailerMockService() *mailerMockService {
	return &mailerMockService{}
}

func (m *mailerMockService) SendVerificationEmail(
	_ context.Context,
	i mailermq.SendVerificationEmailRequest,
) error {
	mockMailStore[i.Receiver] = i.Token
	return nil
}

func (m *mailerMockService) SendPasswordResetEmail(
	_ context.Context,
	i mailermq.SendPasswordResetEmailRequest,
) error {
	mockMailStore[i.Receiver] = i.Token
	return nil
}
