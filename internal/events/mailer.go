package events

import "context"

type MailerEventer interface {
	Send(ctx context.Context, inp SendRequest)
}

type MailerEvents struct{}

type SendRequest struct {
	RequestID    string            `json:"request_id"`
	Receiver     string            `json:"receiver"`
	TemplateName string            `json:"template_name"`
	Options      map[string]string `json:"options"`
}
