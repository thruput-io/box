## Portal as a living README

The portal page should be a **living README**: the single place developers visit to find URLs and the current state of the local box.

Important principles:
- **URLs are first-class** (no localhost).
- **Fail fast and loud**: if data cannot be collected or parsed, the portal must show the error.
- **Entra explains itself**: the portal should only link to Entra’s own index page, not duplicate explanations.

## Structure of data

Inside `portal/collect` there are scripts that generate JSON contracts consumed by `portal/server.js`.

- `collect/static.sh` → `data/static.json`
- `collect/runtime.sh` → `data/runtime.json`

The portal:
- runs both collectors on startup (errors are shown in the UI)
- exposes `POST /api/refresh` that re-runs both collectors

### `data/static.json`

Collects data that is known at image/runtime startup and does not depend on what is currently running.

Good examples:
- canonical URLs (portal, browser, identity/entra, msal-client, tools MCP SSE)
- service names and container names
- DNS zone and resolver IP
- cert filenames and important filesystem paths (derived from `BOX_ROOT`)

### `data/runtime.json`

Collects data that is expected to change while the box is running.

Good examples:
- docker version and container status
- docker images (build timestamps)
- certificate presence and validity dates (from `/certs`)
- environment variables relevant for the box

## Minimal contract (example shape)

`static.json` and `runtime.json` are ordinary JSON files.

```json
{
  "generatedAt": "2026-03-27T23:37:00.000Z",
  "urls": {
    "portal": "https://portal.web.internal",
    "toolsMcpSse": "https://tools.web.internal/sse"
  }
}
```



