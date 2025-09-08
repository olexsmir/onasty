FROM onasty:builder AS builder

COPY cmd cmd
COPY internal internal

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN --mount=type=cache,target=/root/.cache/go-build,id=onasty-go-build \
    --mount=type=cache,target=/go/pkg/mod,id=onasty-go-mod \
    go build -trimpath -ldflags='-w -s' -o /onasty ./cmd/api

FROM onasty:runtime
COPY --from=builder /onasty /onasty
ENTRYPOINT ["/onasty"]
