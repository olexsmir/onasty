package mailermq

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/olexsmir/onasty/internal/events"
	"github.com/olexsmir/onasty/internal/transport/http/reqid"
)

const sendTopic = "mailer.send"

type Mailer interface {
	// SendVerificationEmail sends an email with a verification token to the user.
	SendVerificationEmail(ctx context.Context, input SendVerificationEmailRequest) error

	// SendPasswordResetEmail sends an email with a password reset token to the user.
	SendPasswordResetEmail(ctx context.Context, input SendPasswordResetEmailRequest) error

	// SendChangeEmailVerification sends an email with a change email verification token to the user.
	SendChangeEmailConfirmation(ctx context.Context, inp SendChangeEmailConfirmationRequest) error
}

type MailerMQ struct {
	nc *nats.Conn
}

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

	resp, err := m.nc.RequestWithContext(ctx, sendTopic, req)
	if err != nil {
		return err
	}

	return events.CheckRespForError(resp)
}

type SendPasswordResetEmailRequest struct {
	Receiver string
	Token    string
}

func (m MailerMQ) SendPasswordResetEmail(
	ctx context.Context,
	inp SendPasswordResetEmailRequest,
) error {
	req, err := json.Marshal(sendRequest{
		RequestID:    reqid.GetContext(ctx),
		Receiver:     inp.Receiver,
		TemplateName: "reset_password",
		Options: map[string]string{
			"token": inp.Token,
		},
	})
	if err != nil {
		return err
	}

	resp, err := m.nc.RequestWithContext(ctx, sendTopic, req)
	if err != nil {
		return err
	}

	return events.CheckRespForError(resp)
}

type SendChangeEmailConfirmationRequest struct {
	Receiver string
	Token    string
	NewEmail string
}

func (m MailerMQ) SendChangeEmailConfirmation(
	ctx context.Context,
	inp SendChangeEmailConfirmationRequest,
) error {
	req, err := json.Marshal(sendRequest{
		RequestID:    reqid.GetContext(ctx),
		Receiver:     inp.Receiver,
		TemplateName: "confirm_email_change",
		Options: map[string]string{
			"token": inp.Token,
			"email": inp.NewEmail,
		},
	})
	if err != nil {
		return err
	}

	resp, err := m.nc.RequestWithContext(ctx, sendTopic, req)
	if err != nil {
		return err
	}

	return events.CheckRespForError(resp)
}
