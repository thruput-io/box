const http = require("http");
const { execSync } = require("child_process");
const fs = require("fs");
const path = require("path");

const port = Number(process.env.PORT || 3000);

const SERVICES = [
  { name: "portal", url: "https://portal.web.internal", container: "box.portal" },
  { name: "browser", url: "https://browser.web.internal", container: "box.browser" },
  { name: "entra", url: "https://identity.web.internal", container: "box.entra" },
  { name: "entry (traefik)", url: null, container: "box.entry" },
  { name: "dns (coredns)", url: null, container: "box.dns" },
  { name: "postgres", url: null, container: "box.postgres" },
  { name: "servicebus", url: null, container: "box.servicebus" },
  { name: "sqledge", url: null, container: "box.sqledge" },
  { name: "msal-client", url: "https://msal-client.web.internal", container: "box.msal-client" },
];

function getContainerInfo() {
  try {
    const raw = execSync(
      'docker ps -a --format \'{"name":"{{.Names}}","status":"{{.Status}}","image":"{{.Image}}","created":"{{.CreatedAt}}","ports":"{{.Ports}}"}\' 2>/dev/null',
      { encoding: "utf-8", timeout: 5000 }
    ).trim();
    if (!raw) return [];
    return raw.split("\n").map((line) => {
      try { return JSON.parse(line); } catch { return null; }
    }).filter(Boolean);
  } catch {
    return [];
  }
}

function getCertInfo() {
  const certDir = "/certs";
  const files = ["tls-cert.pem", "tls-key.pem", "dev-root-ca.crt", "identity-signing.key"];
  const result = [];
  for (const f of files) {
    const fp = path.join(certDir, f);
    try {
      const stat = fs.statSync(fp);
      let extra = "";
      if (f.endsWith(".pem") || f.endsWith(".crt")) {
        try {
          const out = execSync(`openssl x509 -in ${fp} -noout -dates 2>/dev/null`, { encoding: "utf-8", timeout: 3000 }).trim();
          extra = out;
        } catch { /* ignore */ }
      }
      result.push({ file: f, exists: true, modified: stat.mtime.toISOString(), size: stat.size, details: extra });
    } catch {
      result.push({ file: f, exists: false, modified: null, size: 0, details: "" });
    }
  }
  return result;
}

function getEnvVars() {
  const vars = {};
  for (const key of Object.keys(process.env).sort()) {
    if (key.startsWith("BOX_") || key.startsWith("COMPOSE_") || key === "PLATFORM" || key.startsWith("INFRA_")) {
      vars[key] = process.env[key];
    }
  }
  return vars;
}

function getBuildInfo() {
  try {
    const raw = execSync(
      'docker images --format \'{"repo":"{{.Repository}}","tag":"{{.Tag}}","created":"{{.CreatedAt}}","size":"{{.Size}}"}\' 2>/dev/null',
      { encoding: "utf-8", timeout: 5000 }
    ).trim();
    if (!raw) return [];
    return raw.split("\n").map((line) => {
      try { return JSON.parse(line); } catch { return null; }
    }).filter(Boolean).filter((img) => img.repo.includes("hostingcompose") || img.repo.includes("infra"));
  } catch {
    return [];
  }
}

function apiData() {
  const containers = getContainerInfo();
  const containerMap = {};
  for (const c of containers) containerMap[c.name] = c;

  const services = SERVICES.map((svc) => {
    const c = containerMap[svc.container];
    return {
      name: svc.name,
      url: svc.url,
      container: svc.container,
      status: c ? c.status : "not running",
      image: c ? c.image : "—",
      running: c ? c.status.startsWith("Up") : false,
    };
  });

  return {
    services,
    certs: getCertInfo(),
    env: getEnvVars(),
    images: getBuildInfo(),
    timestamp: new Date().toISOString(),
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
  </style>
</head>
<body>
  <h1>&#x1F4E6; Developer Portal</h1>
  <p class="subtitle">portal.web.internal &mdash; <span id="ts"></span></p>

  <h2>Services</h2>
  <div class="grid" id="services"></div>

  <h2>Certificates</h2>
  <table id="certs"><thead><tr><th>File</th><th>Status</th><th>Modified</th><th>Details</th></tr></thead><tbody></tbody></table>

  <h2>Docker Images (build times)</h2>
  <table id="images"><thead><tr><th>Repository</th><th>Tag</th><th>Created</th><th>Size</th></tr></thead><tbody></tbody></table>

  <h2>Environment Variables</h2>
  <table id="envvars"><thead><tr><th>Variable</th><th>Value</th></tr></thead><tbody></tbody></table>

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
      } catch (e) {
        console.error("Failed to load status", e);
      }
    }
    load();
    setInterval(load, 30000);
  </script>
</body>
</html>`;

const server = http.createServer((req, res) => {
  if (req.url === "/health") {
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ status: "ok" }));
    return;
  }

  if (req.url === "/api/status") {
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(JSON.stringify(apiData()));
    return;
  }

  res.writeHead(200, { "Content-Type": "text/html; charset=utf-8" });
  res.end(html);
});

server.listen(port, "0.0.0.0", () => {
  console.log(`Developer portal listening on 0.0.0.0:${port}`);
});
