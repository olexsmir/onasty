package mailer

import "context"

var _ Mailer = (*TestMailer)(nil)

type TestMailer struct {
	emails map[string]string
}

func NewTestMailer() *TestMailer {
	return &TestMailer{
		emails: make(map[string]string),
	}
}

func (t *TestMailer) Send(_ context.Context, to, _, content string) error {
	t.emails[to] = content
	return nil
}

func (t *TestMailer) GetLastSentEmailToEmail(email string) string {
	return t.emails[email]
}
