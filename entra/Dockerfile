ARG IDENTITY_CONFIG="DefaultConfig.yaml"

FROM golang:1.26-alpine AS builder
ARG IDENTITY_CONFIG
WORKDIR /app
RUN apk add --no-cache bash make git
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
ENV PATH="/go/bin:${PATH}"

COPY entrago/go.mod ./
COPY entrago/ .

RUN make test build

COPY entrago/${IDENTITY_CONFIG} ./Config.yaml

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/Config.yaml /app/Config.yaml
COPY --from=builder /app/Config.schema.json /app/Config.schema.json
EXPOSE 8080
CMD ["/app/server"]
