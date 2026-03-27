// MCP-over-SSE server for the diagnostic `tools` container.
//
// Security is enforced by Traefik (mTLS). This process only implements MCP
// transport + a tiny tool for validation.

const { McpServer } = require("@modelcontextprotocol/sdk/server/mcp.js");
const { SSEServerTransport } = require("@modelcontextprotocol/sdk/server/sse.js");
const { createMcpExpressApp } = require("@modelcontextprotocol/sdk/server/express.js");
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
