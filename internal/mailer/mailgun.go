package mailer

import (
	"context"
	"log/slog"

	"github.com/mailgun/mailgun-go/v4"
)

var _ Mailer = (*Mailgun)(nil)

type Mailgun struct {
	from string

	mg *mailgun.MailgunImpl
}

func NewMailgun(from, domain, apiKey string) *Mailgun {
	mg := mailgun.NewMailgun(domain, apiKey)
	return &Mailgun{
		from: from,
		mg:   mg,
	}
}

func (m *Mailgun) Send(ctx context.Context, to, subject, content string) error {
	msg := m.mg.NewMessage(m.from, subject, "", to)
	msg.SetHtml(content)

	_, _, err := m.mg.Send(ctx, msg)

	slog.Info("email sent", "to", to)
	slog.Debug("email sent", "subject", subject, "content", content, "err", err)

	return err
}
