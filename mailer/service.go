package main

import (
	"context"
	"log/slog"
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

func (s Service) Send(
	ctx context.Context,
	cancel context.CancelFunc,
	receiver, templateName string,
	templateOpts map[string]string,
) error {
	tmpl, err := getTemplate(s.appURL, templateName)
	if err != nil {
		return err
	}

	t := tmpl(templateOpts)

	go func() {
		select {
		case <-ctx.Done():
			return
		default:
			if err := s.mg.Send(ctx, receiver, t.Subject, t.Body); err != nil {
				slog.ErrorContext(ctx, "failed to send email",
					"template_name", templateName,
					"err", err)
			}
			cancel()
		}
	}()

	return nil
}
