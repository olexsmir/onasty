package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go/micro"
	"github.com/olexsmir/onasty/internal/transport/http/reqid"
)

type Handlers struct {
	service *Service
}

func NewHandlers(service *Service) *Handlers {
	return &Handlers{
		service: service,
	}
}

func (h Handlers) RegisterAll(svc micro.Service) error {
	m := svc.AddGroup("mailer")
	if err := m.AddEndpoint("ping", micro.HandlerFunc(h.pingHandler)); err != nil {
		return err
	}

	if err := m.AddEndpoint("send", micro.HandlerFunc(h.sendHandler)); err != nil {
		return err
	}

	return nil
}

type pingResponse struct {
	Message string `json:"message"`
}

func (h Handlers) pingHandler(req micro.Request) {
	_ = req.RespondJSON(pingResponse{
		Message: "pong",
	})
}

type sendRequest struct {
	RequestID string `json:"request_id"`

	Receiver     string            `json:"receiver"`
	TemplateName string            `json:"template_name"`
	Options      map[string]string `json:"options"`
}

func (h Handlers) sendHandler(req micro.Request) {
	// NOTE: the input probably should be validated
	var inp sendRequest
	if err := json.Unmarshal(req.Data(), &inp); err != nil {
		slog.Error("failed to unmarshal input data", "err", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	ctx = reqid.SetContext(ctx, inp.RequestID)

	if err := h.service.Send(ctx, cancel, inp.Receiver, inp.TemplateName, inp.Options); err != nil {
		_ = req.Error("500", err.Error(), nil)
	}

	_ = req.Respond(nil)
}
