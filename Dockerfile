FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.23-alpine AS backend-builder

WORKDIR /app/go-backend
COPY go-backend/go.mod go-backend/go.sum ./
RUN go mod download
COPY go-backend/ ./
COPY --from=frontend-builder /app/go-backend/frontend_dist ./frontend_dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o clawmemory ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=backend-builder /app/go-backend/clawmemory .
COPY --from=frontend-builder /app/go-backend/frontend_dist ./frontend_dist/

RUN mkdir -p /app/data /app/backups /app/logs

ENV HOST=0.0.0.0
ENV PORT=8765
ENV DATA_DIR=/app/data

EXPOSE 8765

VOLUME ["/app/data", "/app/backups", "/app/logs"]

ENTRYPOINT ["./clawmemory"]
