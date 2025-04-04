package main

import (
	"context"
	"log/slog"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/olexsmir/onasty/internal/transport/http/reqid"
)

type Mailer interface {
	Send(ctx context.Context, to, subject, content string) error
}

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
	msg := mailgun.NewMessage(m.from, subject, "", to)
	msg.SetHTML(content)

	slog.InfoContext(ctx, "email sent", "to", to)

	_, _, err := m.mg.Send(ctx, msg)
	if err != nil {
		RecordEmailFailed(reqid.GetContext(ctx))
		return err
	}

	slog.DebugContext(ctx, "email sent", "subject", subject, "content", content, "err", err)
	slog.InfoContext(ctx, "email sent", "to", to)

	RecordEmailSent()

	return nil
}
