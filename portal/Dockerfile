FROM node:20-alpine

RUN apk add --no-cache docker-cli openssl

WORKDIR /app/portal

COPY portal/ ./

RUN mkdir -p data

ENV PORT=3000

EXPOSE 3000

CMD ["node", "server.js"]
