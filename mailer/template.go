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
	case "confirm_email_change":
		return confirmEmailChangeTemplate(appURL), nil
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

func confirmEmailChangeTemplate(appURL string) TemplateFunc {
	return func(opts map[string]string) Template {
		link := fmt.Sprintf("%[1]s/api/v1/auth/change-email/%[2]s", appURL, opts["token"])

		return Template{
			Subject: "Onasty: confirm your email change",
			Body: fmt.Sprintf(`
It seems like you have changed your email address to %[1]s.
<br>
To confirm this change, please follow this link:
<a href="%[2]s">%[2]s</a>
<br>
<br>
If you did not request email change, you can ignore this message.
<br>
This link will expire after 24 hours.
`, opts["email"], link),
		}
	}
}
