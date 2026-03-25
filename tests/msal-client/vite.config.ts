import { defineConfig } from 'vite';
import fs from 'node:fs';
import path from 'node:path';

export default defineConfig({
  plugins: [
    {
      name: 'strict-404',
      configurePreviewServer(server) {
        server.middlewares.use((req, res, next) => {
          const url = (req.url || '/').split('?')[0];
          // Handle cases like '/' and '/index.html'
          if (url === '/' || url === '/index.html') {
            return next();
          }
          const filePath = path.join(process.cwd(), 'dist', url);
          if (!fs.existsSync(filePath)) {
            res.statusCode = 404;
            res.end(`Not Found: ${url} (Strict 404)`);
            return;
          }
          next();
        });
      },
    },
  ],
  server: {
    port: 8080,
    host: true,
    strictPort: true,
    allowedHosts: ['msal-client.web.internal'],
  },
  preview: {
    port: 8080,
    host: true,
    strictPort: true,
    allowedHosts: ['msal-client.web.internal'],
  }
});
