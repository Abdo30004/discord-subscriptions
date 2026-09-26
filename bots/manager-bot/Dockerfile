# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

COPY bots/manager-bot/package*.json ./
RUN npm ci

COPY bots/manager-bot/tsconfig.json ./
COPY bots/manager-bot/src/ ./src/
RUN npm run build

# Production runtime stage
FROM node:20-alpine

WORKDIR /app

COPY bots/manager-bot/package*.json ./
RUN npm ci --only=production

COPY --from=builder /app/dist ./dist

CMD ["node", "dist/index.js"]
