FROM golang:1.24.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd cmd
COPY internal internal

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -trimpath -ldflags='-w -s' -o /onasty ./cmd/server

FROM alpine:3.21
COPY --from=builder /onasty /onasty
RUN apk --no-cache add ca-certificates
ENTRYPOINT ["/onasty"]
