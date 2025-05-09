package main

import (
	"errors"
	"fmt"
)

type Template struct {
	Subject string
	Body    string
}

type TemplateFunc func(args map[string]string) Template

func getTemplate(appURL string, templateName string) (TemplateFunc, error) {
	switch templateName {
	case "email_verification":
		return emailVerificationTemplate(appURL), nil
	case "reset_password":
		return passwordResetTemplete(appURL), nil
	default:
		return nil, errors.New("failed to get template") //nolint:err113
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

func passwordResetTemplete(appURL string) TemplateFunc {
	return func(opts map[string]string) Template {
		return Template{
			Subject: "Onasty: reset your password",
			// TODO: change the link after making frontend
			Body: fmt.Sprintf(`To reset your password, use this api:
<a href="%[1]s/api/v1/auth/reset-password/%[2]s">%[1]s/api/v1/auth/reset-password/%[2]s</a>
<br />
<br />
This link will expire after an hour.`, appURL, opts["token"]),
		}
	}
}
