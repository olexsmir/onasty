package mailermq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/olexsmir/onasty/internal/transport/http/reqid"
)

type MailerMQ struct {
	nc *nats.Conn
}

const (
	sendMailSubject = "mailer.send"
	sendMailTimeout = 5 * time.Second
)

func New(nc *nats.Conn) *MailerMQ {
	return &MailerMQ{
		nc: nc,
	}
}

type SendRequest struct {
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
	req, err := json.Marshal(SendRequest{
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

	_, err = m.nc.Request(sendMailSubject, req, sendMailTimeout)
	if err != nil {
		return err
	}

	// TODO: handle error that service might return

	return err
}
