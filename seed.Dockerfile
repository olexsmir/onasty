FROM onasty:builder AS builder

WORKDIR /app

COPY internal internal
COPY cmd/seed cmd/seed

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN --mount=type=cache,target=/root/.cache/go-build,id=onasty-go-build \
    --mount=type=cache,target=/go/pkg/mod,id=onasty-go-mod \
    go build -trimpath -ldflags='-w -s' -o /seed ./cmd/seed

FROM onasty:runtime
COPY --from=builder /seed /seed
ENTRYPOINT ["/seed"]
