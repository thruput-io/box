// MCP-over-SSE server for the diagnostic `tools` container.
//
// NOTE: This endpoint is intentionally unprotected (no mTLS) and is meant to be
// reachable only via the local Traefik entrypoint bound to localhost.

const { McpServer } = require("@modelcontextprotocol/sdk/server/mcp.js");
const { SSEServerTransport } = require("@modelcontextprotocol/sdk/server/sse.js");
const { createMcpExpressApp } = require("@modelcontextprotocol/sdk/server/express.js");
const dns = require("node:dns");
const net = require("node:net");
const { setTimeout: sleep } = require("node:timers/promises");
const z = require("zod/v4");

const port = process.env.PORT ? Number(process.env.PORT) : 3000;

function createServer() {
  const server = new McpServer(
    {
      name: "box-tools",
      version: "1.0.0",
    },
    { capabilities: { logging: {} } },
  );

  server.registerTool(
    "ping",
    {
      description: "Health-check tool. Returns the provided text.",
      inputSchema: {
        text: z.string().default("pong"),
      },
    },
    async ({ text }) => {
      return {
        content: [
          {
            type: "text",
            text,
          },
        ],
      };
    },
  );

  server.registerTool(
    "dns_lookup",
    {
      description: "Resolve a hostname using the container's DNS configuration.",
      inputSchema: {
        hostname: z.string().min(1),
      },
    },
    async ({ hostname }) => {
      const result = {
        hostname,
        addresses: [],
        error: null,
      };

      try {
        // Prefer the Promise API when available, fall back to callback API.
        const resolver = dns.promises?.lookup
          ? dns.promises
          : {
              lookup: (h, opts) =>
                new Promise((resolve, reject) =>
                  dns.lookup(h, opts, (err, address, family) =>
                    err ? reject(err) : resolve({ address, family }),
                  ),
                ),
            };

        const addresses = await resolver.lookup(hostname, { all: true });
        result.addresses = addresses.map((a) => ({ address: a.address, family: a.family }));
      } catch (error) {
        result.error = String(error && error.message ? error.message : error);
      }

      return {
        content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
      };
    },
  );

  server.registerTool(
    "tcp_check",
    {
      description:
        "Attempt a TCP connection to host:port (useful for readiness checks like AMQP 5672).",
      inputSchema: {
        host: z.string().min(1),
        port: z.number().int().min(1).max(65535),
        timeoutMs: z.number().int().min(100).max(60000).default(1500),
      },
    },
    async ({ host, port, timeoutMs }) => {
      const startedAt = Date.now();
      const result = {
        host,
        port,
        ok: false,
        latencyMs: null,
        error: null,
      };

      const socket = new net.Socket();
      try {
        await new Promise((resolve, reject) => {
          const onError = (err) => reject(err);
          socket.setTimeout(timeoutMs, () => reject(new Error(`timeout after ${timeoutMs}ms`)));
          socket.once("error", onError);
          socket.connect(port, host, () => resolve());
        });

        result.ok = true;
      } catch (error) {
        result.error = String(error && error.message ? error.message : error);
      } finally {
        result.latencyMs = Date.now() - startedAt;
        try {
          socket.destroy();
        } catch {
          // ignore
        }
      }

      return {
        content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
      };
    },
  );

  server.registerTool(
    "http_get",
    {
      description:
        "Perform an HTTP GET from inside the tools container (respects system CA trust).",
      inputSchema: {
        url: z.string().url(),
        timeoutMs: z.number().int().min(100).max(60000).default(5000),
        maxBodyBytes: z.number().int().min(0).max(1024 * 1024).default(64 * 1024),
      },
    },
    async ({ url, timeoutMs, maxBodyBytes }) => {
      const startedAt = Date.now();
      const result = {
        url,
        ok: false,
        status: null,
        latencyMs: null,
        headers: {},
        bodyPreview: null,
        error: null,
      };

      const controller = new AbortController();
      const timeout = setTimeout(() => controller.abort(), timeoutMs);

      try {
        // Small delay can help when probing freshly-started services.
        await sleep(10);
        const response = await fetch(url, {
          method: "GET",
          signal: controller.signal,
        });

        result.status = response.status;
        result.ok = response.ok;

        // Headers
        for (const [k, v] of response.headers.entries()) {
          result.headers[k] = v;
        }

        const bodyText = await response.text();
        result.bodyPreview = bodyText.slice(0, maxBodyBytes);
      } catch (error) {
        result.error = String(error && error.message ? error.message : error);
      } finally {
        clearTimeout(timeout);
        result.latencyMs = Date.now() - startedAt;
      }

      return {
        content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
      };
    },
  );

  return server;
}

const app = createMcpExpressApp({
  host: "0.0.0.0",
  allowedHosts: ["tools.web.internal", "localhost", "127.0.0.1"],
});

// Store transports by session ID.
const transports = {};

app.get("/", (req, res) => {
  res.status(200).type("text/plain").send("tools: mcp-server\n");
});

// SSE endpoint for establishing the stream.
// NOTE: This implements the legacy MCP HTTP+SSE transport.
app.get("/sse", async (req, res) => {
  try {
    // The endpoint for POST messages is '/messages'
    const transport = new SSEServerTransport("/messages", res);
    const sessionId = transport.sessionId;
    transports[sessionId] = transport;

    transport.onclose = () => {
      delete transports[sessionId];
    };

    const server = createServer();
    await server.connect(transport);
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error("Error establishing SSE stream:", error);
    if (!res.headersSent) {
      res.status(500).send("Error establishing SSE stream");
    }
  }
});

// Messages endpoint for receiving client JSON-RPC requests.
app.post("/messages", async (req, res) => {
  const sessionId = req.query.sessionId;

  if (!sessionId) {
    res.status(400).send("Missing sessionId parameter");
    return;
  }

  const transport = transports[sessionId];
  if (!transport) {
    res.status(404).send("Session not found");
    return;
  }

  try {
    await transport.handlePostMessage(req, res, req.body);
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error("Error handling /messages:", error);
    if (!res.headersSent) {
      res.status(500).send("Error handling request");
    }
  }
});

app.listen(port, (error) => {
  if (error) {
    // eslint-disable-next-line no-console
    console.error("Failed to start tools MCP server:", error);
    process.exit(1);
  }
  // eslint-disable-next-line no-console
  console.log(`tools mcp server listening on :${port}`);
});

process.on("SIGINT", async () => {
  for (const sessionId in transports) {
    try {
      await transports[sessionId].close();
      delete transports[sessionId];
    } catch {
      // ignore
    }
  }
  process.exit(0);
});
