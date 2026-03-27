const path = require("path");

function isoNow() {
  return new Date().toISOString();
}

function normalizeRoot(root) {
  if (!root) return null;
  const s = String(root).replace(/\/$/, "");
  return s || null;
}

function main() {
  const boxRoot = normalizeRoot(process.env.BOX_ROOT);

  const urls = {
    portal: "https://portal.web.internal",
    browser: "https://browser.web.internal",
    identity: "https://identity.web.internal",
    msalClient: "https://msal-client.web.internal",
    toolsMcpSse: "https://tools.web.internal/sse",
  };

  const entraSelfDescriptionUrl = "https://identity.web.internal";

  const services = [
    { name: "portal", container: "box.portal", url: urls.portal },
    { name: "browser", container: "box.browser", url: urls.browser },
    // Entra explains itself; the portal should only link to it.
    { name: "entra", container: "box.entra", url: entraSelfDescriptionUrl },
    { name: "msal-client", container: "box.msal-client", url: urls.msalClient },
    { name: "entry (traefik)", container: "box.entry", url: null },
    { name: "dns (coredns)", container: "box.dns", url: null },
    { name: "postgres", container: "box.postgres", url: null },
    { name: "servicebus", container: "box.servicebus", url: null },
    { name: "sqledge", container: "box.sqledge", url: null },
    { name: "tools (mcp)", container: "box.tools", url: "https://tools.web.internal" },
  ];

  const certFiles = ["tls-cert.pem", "tls-key.pem", "dev-root-ca.crt", "identity-signing.key"];

  const staticData = {
    generatedAt: isoNow(),
    urls,
    entra: {
      selfDescriptionUrl: entraSelfDescriptionUrl,
    },
    dns: {
      zone: "web.internal",
      resolverIp: "172.18.0.253",
    },
    paths: {
      boxRoot,
      certsDir: boxRoot ? path.join(boxRoot, "certs") : null,
    },
    certFiles,
    services,
  };

  process.stdout.write(`${JSON.stringify(staticData, null, 2)}\n`);
}

main();
