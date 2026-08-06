# ---- etapa de compilación ----
FROM golang:1.26-alpine AS builder

WORKDIR /build

# Las dependencias se copian primero para aprovechar la caché de capas:
# solo se reinstalan si go.mod o go.sum cambian.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO deshabilitado para producir un binario estático.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /build/api ./cmd/api

# ---- etapa final ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    adduser -D -u 10001 appuser

COPY --from=builder /build/api /usr/local/bin/api

USER appuser

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
    CMD wget -qO- http://127.0.0.1:${PORT:-3000}/health || exit 1

ENTRYPOINT ["/usr/local/bin/api"]
