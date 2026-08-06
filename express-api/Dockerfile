# ---- etapa de compilación ----
FROM node:20-alpine AS builder

WORKDIR /build

# Las dependencias se copian primero para aprovechar la caché de capas.
COPY package*.json ./
RUN npm ci

COPY tsconfig.json ./
COPY src ./src

RUN npm run build

# ---- etapa de dependencias de producción ----
FROM node:20-alpine AS deps

WORKDIR /deps

COPY package*.json ./
RUN npm ci --omit=dev

# ---- etapa final ----
FROM node:20-alpine

RUN addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=deps  /deps/node_modules ./node_modules
COPY --from=builder /build/dist ./dist
COPY package.json ./

USER app

EXPOSE 4000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
    CMD wget -qO- http://127.0.0.1:${PORT:-4000}/health || exit 1

CMD ["node", "dist/server.js"]
