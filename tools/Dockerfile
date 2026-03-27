FROM nicolaka/netshoot:latest

# Install packages.
# - nodejs/npm: needed for the MCP SDK
# - ca-certificates: required to trust our local dev CA (mounted at /certs)
RUN apk add --no-cache nodejs npm ca-certificates

WORKDIR /app

# 2. Initialize and install the MCP SDK
RUN npm init -y && npm install @modelcontextprotocol/sdk

# 3. Trust local dev root CA if mounted (see compose.infra.yaml volume mount)
COPY install-certs.sh /usr/local/bin/install-certs
RUN chmod +x /usr/local/bin/install-certs

# 4. Copy your MCP server logic
COPY mcp-server.js .

# Standard port for MCP over SSE
EXPOSE 3000

CMD ["/bin/sh", "-lc", "/usr/local/bin/install-certs && exec node mcp-server.js"]