# Build stage for React frontend
FROM node:20-alpine AS frontend-build

WORKDIR /app/web

COPY web/package*.json ./
RUN npm ci --prefer-offline --no-audit

COPY web/ ./
RUN npm run build

# Build stage for Go backend
FROM golang:1.24-alpine AS backend-build

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY cmd/ ./cmd/
COPY api/ ./api/
COPY internal/ ./internal/
COPY config/ ./config/
COPY constants/ ./constants/
COPY repo/ ./repo/
COPY templates/ ./templates/
COPY .env .

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -v -p 4 -ldflags="-s -w" -o server ./cmd/main

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=backend-build /app/server .
COPY --from=backend-build /app/.env .
COPY --from=backend-build /app/templates ./templates
COPY --from=frontend-build /app/web/dist ./web/dist

EXPOSE 8016

CMD ["./server"]