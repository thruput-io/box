FROM mcr.microsoft.com/playwright:v1.58.2-noble

WORKDIR /test

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libnss3-tools \
    && rm -rf /var/lib/apt/lists/*

# Copy package files and install dependencies
COPY tests/msal-client/package*.json ./
RUN npm install

# Copy source code (explicitly; avoid bringing host node_modules/dist)
COPY tests/msal-client/index.html ./
COPY tests/msal-client/playwright.config.ts ./
COPY tests/msal-client/tsconfig.json ./
COPY tests/msal-client/vite.config.ts ./
COPY tests/msal-client/src ./src
COPY tests/msal-client/tests ./tests

# Import CA cert into system trust and Chromium NSS store at build time
COPY certs/dev-root-ca.crt /usr/local/share/ca-certificates/mkcert.crt
RUN update-ca-certificates >/dev/null 2>&1 && \
    mkdir -p /root/.pki/nssdb && \
    certutil -d sql:/root/.pki/nssdb -N --empty-password && \
    certutil -d sql:/root/.pki/nssdb -A -t "CT,C,C" -n "it-local-ca" -i /usr/local/share/ca-certificates/mkcert.crt

# Build the application
RUN npm run build

EXPOSE 8080

# Run playwright tests
CMD ["npm", "test"]
