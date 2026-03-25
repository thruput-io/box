ARG IDENTITY_CONFIG="DefaultConfig.yaml"

FROM golang:1.22-alpine AS builder
ARG IDENTITY_CONFIG
WORKDIR /app
COPY go.mod ./
COPY *.go *.html *.json ./
RUN go mod tidy
COPY ${IDENTITY_CONFIG} ./Config.yaml
RUN go test -v ./
RUN go build -o server .

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/server /app/server
COPY --from=builder /app/Config.yaml /app/Config.yaml
COPY --from=builder /app/Config.schema.json /app/Config.schema.json
COPY --from=builder /app/login.html /app/login.html
EXPOSE 8080
CMD ["/app/server"]
