package events

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

const (
	natsHeaderErrorCode = "Nats-Service-Error-Code"
	natsHeaderErrorMsg  = "Nats-Service-Error"
)

var _ error = (*Error)(nil)

type Error struct {
	Code    string
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("code: %s; msg: %s", e.Code, e.Message)
}

func CheckRespForError(resp *nats.Msg) error {
	code := resp.Header.Get(natsHeaderErrorCode)
	msg := resp.Header.Get(natsHeaderErrorMsg)
	if code == "" && msg == "" {
		return nil
	}

	return &Error{
		Code:    code,
		Message: msg,
	}
}
