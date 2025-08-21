# mailer service

All templates could be found *[here](./template.go)*

## endpoints
### `mailer.ping`
This endpoint always returns pong message

Response:
```json
{
  "message": "pong"
}
```

### `mailer.send`

Input
- `request_id` : *string* - (optional) the request id, needed to keep consistency across services
- `receiver` : *string* - the email receiver
- `template_name` : *string* - the template that's going to be used
- `options` : *Map<string, string>* - template specific options


Example input
```json
{
  "request_id": "hello_world",
  "receiver": "onasty@example.com",
  "template_name": "email_verification",
  "options": {
    "token": "the_verification_token"
  }
}
```

#### Template specific options
- `email_verification`
  - `token` the token that is used in verification link
- `reset_password`
  - `token` the token that is used in password reset link
- `confirm_email_change`
  - `email` the email user want to set as new
  - `token` the token that is used in confirm link
