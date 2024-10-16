FROM golang:1.23.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd cmd
COPY internal internal

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s' -o /onasty ./cmd/server

FROM alpine:3.20
COPY --from=builder /onasty /onasty
RUN apk --no-cache add ca-certificates
ENTRYPOINT ["/onasty"]
