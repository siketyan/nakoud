FROM golang:1.20.3-bullseye AS builder

WORKDIR /app
COPY ./go.mod ./go.sum /app/
RUN go mod download

COPY ./ /app
RUN CGO_ENABLED=0 go build -o /bin/proxy ./cmd/proxy

FROM gcr.io/distroless/static-debian11

COPY --from=builder /bin/proxy /bin/nakoud-proxy
ENTRYPOINT ["/bin/nakoud-proxy", "--bind", "0.0.0.0:8080"]
