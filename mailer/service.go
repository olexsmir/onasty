package main

import (
	"context"
)

type Service struct {
	appURL string
	mg     *Mailgun
}

func NewService(appURL string, mg *Mailgun) *Service {
	return &Service{
		appURL: appURL,
		mg:     mg,
	}
}

type sendInput struct {
	Receiver     string
	TemplateName string
	Options      map[string]string
}

func (s Service) Send(ctx context.Context, inp sendInput) error {
	tmpl, err := getTemplate(s.appURL, inp.TemplateName)
	if err != nil {
		return err
	}

	t := tmpl(inp.Options)

	// prob better to create sep context with time out and use it here

	if err := s.mg.Send(ctx, inp.Receiver, t.Subject, t.Body); err != nil {
		return err
	}

	return nil
}
