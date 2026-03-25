FROM alpine:3.21

# Install openssl and wget
RUN apk add --no-cache openssl wget libcap

# Download mkcert binary matching build architecture
ARG TARGETARCH=amd64
RUN wget --connect-timeout=1 -q -O /usr/local/bin/mkcert https://github.com/FiloSottile/mkcert/releases/download/v1.4.4/mkcert-v1.4.4-linux-${TARGETARCH} && \
    chmod +x /usr/local/bin/mkcert

WORKDIR /app
COPY generate.sh .
RUN chmod +x generate.sh

CMD ["./generate.sh"]
