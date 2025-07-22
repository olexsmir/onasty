package main

import (
	"errors"
	"fmt"
)

var ErrInvalidTemplate = errors.New("failed to get template")

type Template struct {
	Subject string
	Body    string
}

type TemplateFunc func(args map[string]string) Template

func getTemplate(appURL, frontendURL string, templateName string) (TemplateFunc, error) {
	switch templateName {
	case "email_verification":
		return emailVerificationTemplate(appURL), nil
	case "reset_password":
		return passwordResetTemplate(frontendURL), nil
	default:
		return nil, ErrInvalidTemplate
	}
}

func emailVerificationTemplate(appURL string) TemplateFunc {
	return func(opts map[string]string) Template {
		return Template{
			Subject: "Onasty: verify your email",
			Body: fmt.Sprintf(`To verify your email, please follow this link:
<a href="%[1]s/api/v1/auth/verify/%[2]s">%[1]s/api/v1/auth/verify/%[2]s</a>
<br />
<br />
This link will expire after 24 hours.`, appURL, opts["token"]),
		}
	}
}

func passwordResetTemplate(frontendURL string) TemplateFunc {
	return func(opts map[string]string) Template {
		return Template{
			Subject: "Onasty: reset your password",
			Body: fmt.Sprintf(`To reset your password, use this api:
<a href="%[1]s/auth?token=%[2]s">%[1]s/auth?token=%[2]s</a>
<br />
<br />
This link will expire after an hour.`, frontendURL, opts["token"]),
		}
	}
}
