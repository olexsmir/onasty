FROM golang:1.23.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY internal internal
COPY mailer mailer

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -trimpath -ldflags='-w -s' -o /mailer ./mailer


FROM alpine:3.21
COPY --from=builder /mailer /mailer
RUN apk --no-cache add ca-certificates
ENTRYPOINT ["/mailer"]
