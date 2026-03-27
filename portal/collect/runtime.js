const { execSync } = require("child_process");
const fs = require("fs");
const path = require("path");

function isoNow() {
  return new Date().toISOString();
}

function exec(cmd, timeoutMs) {
  return execSync(cmd, { encoding: "utf-8", timeout: timeoutMs, stdio: ["ignore", "pipe", "pipe"] }).trim();
}

function readJson(filePath) {
  const raw = fs.readFileSync(filePath, "utf-8");
  return JSON.parse(raw);
}

function dockerVersion() {
  return exec("docker -v 2>/dev/null", 3000);
}

function dockerContainers() {
  const raw = exec(
    'docker ps -a --format \'{"name":"{{.Names}}","status":"{{.Status}}","image":"{{.Image}}","created":"{{.CreatedAt}}","ports":"{{.Ports}}"}\' 2>/dev/null',
    5000
  );
  if (!raw) return [];
  return raw
    .split("\n")
    .map((line) => {
      try {
        return JSON.parse(line);
      } catch {
        return null;
      }
    })
    .filter(Boolean);
}

function dockerImages() {
  const raw = exec(
    'docker images --format \'{"repo":"{{.Repository}}","tag":"{{.Tag}}","created":"{{.CreatedAt}}","size":"{{.Size}}"}\' 2>/dev/null',
    5000
  );
  if (!raw) return [];
  return raw
    .split("\n")
    .map((line) => {
      try {
        return JSON.parse(line);
      } catch {
        return null;
      }
    })
    .filter(Boolean);
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

function certInfo(certFiles) {
  const certDir = "/certs";
  return certFiles.map((file) => {
    const fp = path.join(certDir, file);
    try {
      const stat = fs.statSync(fp);
      let extra = "";
      if (file.endsWith(".pem") || file.endsWith(".crt")) {
        try {
          extra = exec(`openssl x509 -in ${fp} -noout -dates 2>/dev/null`, 3000);
        } catch {
          extra = "";
        }
      }

      return { file, exists: true, modified: stat.mtime.toISOString(), size: stat.size, details: extra };
    } catch {
      return { file, exists: false, modified: null, size: 0, details: "" };
    }
  });
}

function main() {
  const portalDir = path.dirname(__dirname);
  const staticPath = path.join(portalDir, "data", "static.json");
  const staticData = readJson(staticPath);

  const images = dockerImages();
  const filteredImages = images.filter(
    (img) => typeof img.repo === "string" && (img.repo.includes("hostingcompose") || img.repo.includes("infra"))
  );

  const runtimeData = {
    generatedAt: isoNow(),
    system: {
      os: exec("uname 2>/dev/null", 1000),
      user: exec("whoami 2>/dev/null", 1000),
    },
    docker: {
      version: dockerVersion(),
      containers: dockerContainers(),
      images: filteredImages,
    },
    certs: certInfo(staticData.certFiles || []),
    env: getEnvVars(),
  };

  process.stdout.write(`${JSON.stringify(runtimeData, null, 2)}\n`);
}

main();
