package main

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/invopop/jsonschema"
	"github.com/nats-io/nats.go/micro"
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
	if err := m.AddEndpoint("ping",
		micro.HandlerFunc(h.pingHandler),
		micro.WithEndpointMetadata(map[string]string{
			"format":          "application/json",
			"response_schema": schemaFor(&pingResponse{}), //nolint:exhaustruct
		}),
	); err != nil {
		return err
	}

	if err := m.AddEndpoint("send",
		micro.HandlerFunc(h.sendHandler),
		micro.WithEndpointMetadata(map[string]string{
			"format":         "application/json",
			"request_schema": schemaFor(&sendRequest{}), //nolint:exhaustruct
		}),
	); err != nil {
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
	Receiver     string            `json:"receiver"`
	TemplateName string            `json:"template_name"`
	Options      map[string]string `json:"options"`
}

func (h Handlers) sendHandler(req micro.Request) {
	var inp sendRequest
	if err := json.Unmarshal(req.Data(), &inp); err != nil {
		slog.Error("failed to unmarshal input data", "err", err)
		return
	}

	if err := h.service.Send(context.TODO(), sendInput{
		Receiver:     inp.Receiver,
		TemplateName: inp.TemplateName,
		Options:      inp.Options,
	}); err != nil {
		_ = req.Error("500", err.Error(), nil)
	}
}

func schemaFor(t any) string {
	schema := jsonschema.Reflect(t)
	data, _ := schema.MarshalJSON()
	return string(data)
}
