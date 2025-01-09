package mailermq

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/olexsmir/onasty/internal/events"
	"github.com/olexsmir/onasty/internal/transport/http/reqid"
)

type Mailer interface {
	SendVerificationEmail(ctx context.Context, input SendVerificationEmailRequest) error
}

type MailerMQ struct {
	nc *nats.Conn
}

const sendMailSubject = "mailer.send"

func New(nc *nats.Conn) *MailerMQ {
	return &MailerMQ{
		nc: nc,
	}
}

type sendRequest struct {
	RequestID    string            `json:"request_id"`
	Receiver     string            `json:"receiver"`
	TemplateName string            `json:"template_name"`
	Options      map[string]string `json:"options"`
}

type SendVerificationEmailRequest struct {
	Receiver string
	Token    string
}

func (m MailerMQ) SendVerificationEmail(
	ctx context.Context,
	inp SendVerificationEmailRequest,
) error {
	req, err := json.Marshal(sendRequest{
		RequestID:    reqid.GetContext(ctx),
		Receiver:     inp.Receiver,
		TemplateName: "email_verification",
		Options: map[string]string{
			"token": inp.Token,
		},
	})
	if err != nil {
		return err
	}

	resp, err := m.nc.RequestWithContext(ctx, sendMailSubject, req)
	if err != nil {
		return err
	}

	return events.CheckRespForError(resp)
}
