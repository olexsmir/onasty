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
	if templateName == "email_verification" {
		return emailVerificationTemplate(appURL), nil
	}

	return nil, errors.New("failed to get template") //nolint:err113
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
