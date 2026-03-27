const http = require("http");
const fs = require("fs");
const path = require("path");
const { execFileSync } = require("child_process");

const port = Number(process.env.PORT || 3000);

const portalDir = __dirname;
const dataDir = path.join(portalDir, "data");
const staticJsonPath = path.join(dataDir, "static.json");
const runtimeJsonPath = path.join(dataDir, "runtime.json");
const collectDir = path.join(portalDir, "collect");

function escapeHtml(text) {
  return String(text)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function getJunieMcpServerConfig() {
  return {
    transport: "sse",
    url: "https://tools.web.internal/sse",
  };
}

function getJunieMcpServerEntrySnippet() {
  // Snippet intended to be pasted under "mcpServers": { ... }
  // i.e. it includes only the single server entry.
  const snippet = {
    tools: getJunieMcpServerConfig(),
  };
  const json = JSON.stringify(snippet, null, 2);
  // Drop the outer braces so the user gets just:
  //   "tools": { ... }
  return json.replace(/^\{\n/, "").replace(/\n\}$/, "");
}

function getJunieMcpFullSnippet() {
  const snippet = {
    mcpServers: {
      tools: getJunieMcpServerConfig(),
    },
  };
  return JSON.stringify(snippet, null, 2);
}

function renderJunieSnippetHtml() {
  // Display the full shape for context, but grey-out the wrapper so the
  // user focuses on the single server entry.
  const grey = (s) => `<span style="color:#64748b;">${escapeHtml(s)}</span>`;
  const normal = (s) => escapeHtml(s);

  const entryLines = getJunieMcpServerEntrySnippet().trimEnd().split("\n");
  const indented = entryLines.map((l) => (l ? `    ${l}` : l));

  return (
    grey("{") +
    "\n" +
    grey('  "mcpServers": {') +
    "\n" +
    indented.map((l) => normal(l)).join("\n") +
    "\n" +
    grey("  }") +
    "\n" +
    grey("}")
  );
}

function runCollector(scriptFile) {
  const scriptPath = path.join(collectDir, scriptFile);
  execFileSync("sh", [scriptPath], { timeout: 15000, stdio: ["ignore", "ignore", "pipe"] });
}

function safeReadJson(filePath) {
  try {
    const raw = fs.readFileSync(filePath, "utf-8");
    return { data: JSON.parse(raw), error: null };
  } catch (e) {
    return { data: null, error: e instanceof Error ? e.message : String(e) };
  }
}

function buildApiData(startupErrors) {
  const errors = [...startupErrors];

  const staticRead = safeReadJson(staticJsonPath);
  if (staticRead.error) errors.push(`static.json: ${staticRead.error}`);
  const runtimeRead = safeReadJson(runtimeJsonPath);
  if (runtimeRead.error) errors.push(`runtime.json: ${runtimeRead.error}`);

  const staticData = staticRead.data || {};
  const runtimeData = runtimeRead.data || {};

  const containers = (runtimeData.docker && Array.isArray(runtimeData.docker.containers) && runtimeData.docker.containers) || [];
  const containerMap = {};
  for (const c of containers) {
    if (c && typeof c.name === "string") containerMap[c.name] = c;
  }

  const servicesSrc = Array.isArray(staticData.services) ? staticData.services : [];
  const services = servicesSrc.map((svc) => {
    const containerName = svc && typeof svc.container === "string" ? svc.container : "";
    const c = containerName ? containerMap[containerName] : null;
    const status = c && typeof c.status === "string" ? c.status : "not running";
    return {
      name: svc && typeof svc.name === "string" ? svc.name : "(missing name)",
      url: svc && typeof svc.url === "string" ? svc.url : null,
      container: containerName || null,
      status,
      image: c && typeof c.image === "string" ? c.image : "—",
      running: typeof status === "string" ? status.startsWith("Up") : false,
    };
  });

  const images = runtimeData.docker && Array.isArray(runtimeData.docker.images) ? runtimeData.docker.images : [];

  return {
    timestamp: runtimeData.generatedAt || staticData.generatedAt || new Date().toISOString(),
    urls: staticData.urls || {},
    services,
    certs: Array.isArray(runtimeData.certs) ? runtimeData.certs : [],
    env: runtimeData.env || {},
    images,
    errors,
  };
}

const html = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Developer Portal</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #0f172a; color: #e2e8f0; padding: 1.5rem; line-height: 1.5; }
    h1 { font-size: 1.6rem; margin-bottom: 0.3rem; color: #38bdf8; }
    h2 { font-size: 1.1rem; margin: 1.5rem 0 0.5rem; color: #94a3b8; border-bottom: 1px solid #1e293b; padding-bottom: 0.3rem; }
    .subtitle { color: #64748b; font-size: 0.85rem; margin-bottom: 1rem; }
    .grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 0.75rem; }
    .card { background: #1e293b; border-radius: 8px; padding: 0.75rem 1rem; border-left: 3px solid #334155; }
    .card.up { border-left-color: #22c55e; }
    .card.down { border-left-color: #ef4444; }
    .card .svc-name { font-weight: 600; font-size: 0.95rem; }
    .card .svc-name a { color: #38bdf8; text-decoration: none; }
    .card .svc-name a:hover { text-decoration: underline; }
    .card .svc-status { font-size: 0.8rem; color: #94a3b8; margin-top: 0.2rem; }
    .card .svc-image { font-size: 0.75rem; color: #64748b; }
    .dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; margin-right: 6px; }
    .dot.up { background: #22c55e; }
    .dot.down { background: #ef4444; }
    table { width: 100%; border-collapse: collapse; font-size: 0.85rem; }
    table th, table td { text-align: left; padding: 0.4rem 0.6rem; border-bottom: 1px solid #1e293b; }
    table th { color: #64748b; font-weight: 500; }
    table td { color: #cbd5e1; }
    .tag { display: inline-block; padding: 0.1rem 0.5rem; border-radius: 4px; font-size: 0.75rem; font-weight: 600; }
    .tag.ok { background: #14532d; color: #4ade80; }
    .tag.missing { background: #7f1d1d; color: #fca5a5; }
    code { background: #334155; padding: 0.15rem 0.4rem; border-radius: 3px; font-size: 0.8rem; }
    .actions { margin-top: 1.5rem; }
    .btn { display: inline-block; padding: 0.5rem 1.2rem; background: #2563eb; color: #fff; border: none; border-radius: 6px; font-size: 0.85rem; cursor: pointer; text-decoration: none; }
    .btn:hover { background: #1d4ed8; }
    .btn:disabled { background: #334155; color: #64748b; cursor: not-allowed; }
    #test-output { margin-top: 0.5rem; font-size: 0.8rem; color: #94a3b8; white-space: pre-wrap; max-height: 200px; overflow-y: auto; background: #0f172a; padding: 0.5rem; border-radius: 4px; display: none; }
    .refresh-note { font-size: 0.75rem; color: #475569; margin-top: 1rem; }
    .alert { background:#7f1d1d; color:#fecaca; border:1px solid #991b1b; padding:0.6rem 0.8rem; border-radius:6px; margin:0.75rem 0 1rem; display:none; }
    .alert code { background:#991b1b; }
  </style>
</head>
<body>
  <h1>&#x1F4E6; Developer Portal</h1>
  <p class="subtitle">portal.web.internal &mdash; <span id="ts"></span></p>

  <div id="errors" class="alert"></div>

  <div class="actions" style="margin-top:0.75rem;">
    <button id="refresh" class="btn">Refresh data</button>
    <span style="font-size:0.8rem;color:#94a3b8;margin-left:0.5rem;">Runs <code>collect/static.sh</code> and <code>collect/runtime.sh</code> inside the portal container.</span>
  </div>

  <h2>Links</h2>
  <table id="urls"><thead><tr><th>Name</th><th>URL</th></tr></thead><tbody></tbody></table>

  <h2>Services</h2>
  <div class="grid" id="services"></div>

  <h2>Certificates</h2>
  <table id="certs"><thead><tr><th>File</th><th>Status</th><th>Modified</th><th>Details</th></tr></thead><tbody></tbody></table>

  <h2>Docker Images (build times)</h2>
  <table id="images"><thead><tr><th>Repository</th><th>Tag</th><th>Created</th><th>Size</th></tr></thead><tbody></tbody></table>

  <h2>Environment Variables</h2>
  <table id="envvars"><thead><tr><th>Variable</th><th>Value</th></tr></thead><tbody></tbody></table>

  <h2>Junie MCP (tools)</h2>
  <p style="font-size:0.85rem;color:#94a3b8;margin-bottom:0.5rem;">
    The <code>tools</code> service is available over HTTPS inside <code>box</code>.
    Use the snippet below as a starting point.
  </p>
  <div style="display:flex;align-items:flex-start;gap:0.5rem;">
    <pre id="junie-mcp" data-copy="${escapeHtml(getJunieMcpServerEntrySnippet())}" style="flex:1;font-size:0.8rem;color:#cbd5e1;white-space:pre;overflow:auto;background:#0f172a;padding:0.75rem;border-radius:6px;border:1px solid #1e293b;">${renderJunieSnippetHtml()}</pre>
    <div style="display:flex;flex-direction:column;gap:0.5rem;">
      <button id="copy-junie-mcp" class="btn" title="Copy server snippet" style="padding:0.5rem 0.7rem;min-width:2.5rem;">
        <span aria-hidden="true">⧉</span>
      </button>
      <button id="copy-junie-mcp-full" class="btn" title="Copy full mcp.json" style="padding:0.5rem 0.7rem;min-width:2.5rem;background:#334155;">
        <span aria-hidden="true">{…}</span>
      </button>
    </div>
  </div>

  <div class="actions">
    <h2>Self-Test</h2>
    <p style="font-size:0.85rem;color:#94a3b8;margin-bottom:0.5rem;">Run <code>box self-test</code> from your host terminal to execute the full test suite.</p>
  </div>

  <p class="refresh-note">Auto-refreshes every 30s. <a href="/" style="color:#38bdf8;">Refresh now</a></p>

  <script>
    async function load() {
      try {
        const res = await fetch("/api/status");
        const data = await res.json();
        document.getElementById("ts").textContent = new Date(data.timestamp).toLocaleString();

        // Errors
        const errEl = document.getElementById("errors");
        const errs = Array.isArray(data.errors) ? data.errors : [];
        if (errs.length) {
          errEl.style.display = "block";
          errEl.innerHTML = "<strong>Data issues:</strong><br>" + errs.map(e => "<code>" + String(e).replace(/</g,"&lt;") + "</code>").join("<br>");
        } else {
          errEl.style.display = "none";
          errEl.textContent = "";
        }

        // Links
        const utb = document.querySelector("#urls tbody");
        const urls = data.urls && typeof data.urls === "object" ? data.urls : {};
        utb.innerHTML = Object.entries(urls).map(([k, v]) => {
          const val = v ? '<a href="' + v + '" style="color:#38bdf8;">' + v + "</a>" : "—";
          return "<tr><td><code>" + k + "</code></td><td>" + val + "</td></tr>";
        }).join("");

        // Services
        const grid = document.getElementById("services");
        grid.innerHTML = data.services.map(s => {
          const cls = s.running ? "up" : "down";
          const nameHtml = s.url ? '<a href="' + s.url + '">' + s.name + '</a>' : s.name;
          return '<div class="card ' + cls + '">' +
            '<div class="svc-name"><span class="dot ' + cls + '"></span>' + nameHtml + '</div>' +
            '<div class="svc-status">' + s.status + '</div>' +
            '<div class="svc-image">' + s.image + '</div>' +
            '</div>';
        }).join("");

        // Certs
        const ctb = document.querySelector("#certs tbody");
        ctb.innerHTML = data.certs.map(c => {
          const tag = c.exists ? '<span class="tag ok">OK</span>' : '<span class="tag missing">MISSING</span>';
          const mod = c.modified ? new Date(c.modified).toLocaleString() : "—";
          const det = c.details ? c.details.replace(/\\n/g, "<br>") : "—";
          return "<tr><td><code>" + c.file + "</code></td><td>" + tag + "</td><td>" + mod + "</td><td>" + det + "</td></tr>";
        }).join("");

        // Images
        const itb = document.querySelector("#images tbody");
        itb.innerHTML = data.images.map(i =>
          "<tr><td>" + i.repo + "</td><td>" + i.tag + "</td><td>" + i.created + "</td><td>" + i.size + "</td></tr>"
        ).join("");

        // Env
        const etb = document.querySelector("#envvars tbody");
        etb.innerHTML = Object.entries(data.env).map(([k, v]) =>
          "<tr><td><code>" + k + "</code></td><td>" + v + "</td></tr>"
        ).join("");

        // Junie MCP snippet is server-rendered so it is always available to copy.
        // Leave it unchanged here.
      } catch (e) {
        console.error("Failed to load status", e);
      }
    }
    load();
    setInterval(load, 30000);

    document.getElementById("refresh")?.addEventListener("click", async () => {
      const btn = document.getElementById("refresh");
      if (btn) btn.disabled = true;
      try {
        const res = await fetch("/api/refresh", { method: "POST" });
        const data = await res.json().catch(() => ({}));
        if (!res.ok) {
          console.error("Refresh failed", data);
        }
      } finally {
        if (btn) btn.disabled = false;
        await load();
      }
    });

    async function copyText(text) {
      try {
        if (navigator.clipboard && navigator.clipboard.writeText) {
          await navigator.clipboard.writeText(text);
          return;
        }
      } catch {
        // ignore and fallback
      }

      const ta = document.createElement("textarea");
      ta.value = text;
      ta.style.position = "fixed";
      ta.style.left = "-1000px";
      ta.style.top = "-1000px";
      document.body.appendChild(ta);
      ta.focus();
      ta.select();
      try { document.execCommand("copy"); } catch { /* ignore */ }
      document.body.removeChild(ta);
    }

    document.getElementById("copy-junie-mcp")?.addEventListener("click", async () => {
      const el = document.getElementById("junie-mcp");
      const text = el?.getAttribute("data-copy") || "";
      await copyText(text);
    });

    document.getElementById("copy-junie-mcp-full")?.addEventListener("click", async () => {
      await copyText(${JSON.stringify(getJunieMcpFullSnippet())});
    });
  </script>
</body>
</html>`;

const startupErrors = [];
try {
  runCollector("static.sh");
} catch (e) {
  startupErrors.push(`collect/static.sh: ${e instanceof Error ? e.message : String(e)}`);
}
try {
  runCollector("runtime.sh");
} catch (e) {
  startupErrors.push(`collect/runtime.sh: ${e instanceof Error ? e.message : String(e)}`);
}

const server = http.createServer((req, res) => {
  if (req.url === "/health") {
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ status: "ok" }));
    return;
  }

  if (req.url === "/api/refresh" && req.method === "POST") {
    const errors = [];
    try {
      runCollector("static.sh");
    } catch (e) {
      errors.push(`collect/static.sh: ${e instanceof Error ? e.message : String(e)}`);
    }
    try {
      runCollector("runtime.sh");
    } catch (e) {
      errors.push(`collect/runtime.sh: ${e instanceof Error ? e.message : String(e)}`);
    }

    const ok = errors.length === 0;
    res.writeHead(ok ? 200 : 500, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ ok, errors }));
    return;
  }

  if (req.url === "/api/status") {
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(JSON.stringify(buildApiData(startupErrors)));
    return;
  }

  res.writeHead(200, { "Content-Type": "text/html; charset=utf-8" });
  res.end(html);
});

server.listen(port, "0.0.0.0", () => {
  console.log(`Developer portal listening on 0.0.0.0:${port}`);
});
