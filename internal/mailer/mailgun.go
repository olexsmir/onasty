package mailer

import (
	"context"

	"github.com/mailgun/mailgun-go/v4"
)

var _ Mailer = (*Mailgun)(nil)

type Mailgun struct {
	from string

	mg *mailgun.MailgunImpl
}

func NewMailgun(from, domain, apiKey string) *Mailgun {
	mg := mailgun.NewMailgun(domain, apiKey)
	mg.SetAPIBase(mailgun.APIBaseEU)
	return &Mailgun{
		from: from,
		mg:   mg,
	}
}

func (m *Mailgun) Send(ctx context.Context, to, subject, content string) error {
	msg := m.mg.NewMessage(m.from, subject, "", to)
	msg.SetHtml(content)

	_, _, err := m.mg.Send(ctx, msg)
	return err
}
