FROM golang:1.20.3-bullseye AS builder

WORKDIR /app
COPY ./go.mod ./go.sum /app/
RUN --mount=type=cache,target=/root/go/pkg \
    go mod download

COPY ./ /app
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -o /bin/proxy ./cmd/proxy

FROM gcr.io/distroless/static-debian11

LABEL org.opencontainers.image.source="https://github.com/siketyan/nakoud"
LABEL org.opencontainers.image.description="Access your Docker containers easily without port forwarding"
LABEL org.opencontainers.image.licenses=MIT

COPY --from=builder /bin/proxy /bin/nakoud-proxy
ENTRYPOINT ["/bin/nakoud-proxy", "--bind", "0.0.0.0:8080"]
