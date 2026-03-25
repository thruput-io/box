FROM mcr.microsoft.com/playwright:v1.58.2-noble

WORKDIR /test

COPY tests/browser-playwright/package*.json ./
RUN npm install

COPY tests/browser-playwright/playwright.config.ts ./
COPY tests/browser-playwright/tests ./tests

CMD ["npm", "test"]