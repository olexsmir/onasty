package mailer

import (
	"context"
	"sync"
)

var _ Mailer = (*TestMailer)(nil)

type TestMailer struct {
	mu sync.Mutex

	emails map[string]string
}

// NewTestMailer create a mailer for tests
// that implementation of Mailer stores all sent email in memory
// to get the last email sent to a specific email use GetLastSentEmailToEmail
func NewTestMailer() *TestMailer {
	return &TestMailer{ //nolint:exhaustruct
		emails: make(map[string]string),
	}
}

func (t *TestMailer) Send(_ context.Context, to, _, content string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.emails[to] = content

	return nil
}

// GetLastSentEmailToEmail returns the last email sent to a specific email
func (t *TestMailer) GetLastSentEmailToEmail(email string) string {
	t.mu.Lock()
	defer t.mu.Unlock()

	e := t.emails[email]

	return e
}
