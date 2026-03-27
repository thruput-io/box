ARG IDENTITY_CONFIG="DefaultConfig.yaml"

FROM golang:1.26-alpine AS builder
ARG IDENTITY_CONFIG
ARG SKIP_LINT=1
WORKDIR /app
RUN apk add --no-cache curl bc bash
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.11.4

COPY server/go.mod ./
RUN go mod download
COPY server/ .
RUN go mod tidy


# Run linting (temporarily disabled by default)
RUN if [ "$SKIP_LINT" = "1" ]; then echo "Skipping golangci-lint"; else golangci-lint run ./...; fi


RUN chmod +x scripts/coverage.sh && (./scripts/coverage.sh || true)

COPY server/${IDENTITY_CONFIG} ./Config.yaml
RUN go build -o bin/server .

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/Config.yaml /app/Config.yaml
COPY --from=builder /app/Config.schema.json /app/Config.schema.json
RUN printf '%s\n' '<!doctype html><html><body><h1>Sign in</h1></body></html>' > /app/login.html
RUN printf '%s\n' '<!doctype html><html><body><h1>Home</h1></body></html>' > /app/index.html
EXPOSE 8080
CMD ["/app/server"]
