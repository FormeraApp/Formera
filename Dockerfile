# Build Frontend (Nuxt)
FROM node:20-alpine AS frontend-builder

RUN corepack enable && corepack prepare yarn@4.9.4 --activate

WORKDIR /app

COPY frontend/package.json frontend/yarn.lock frontend/.yarnrc.yml ./
RUN yarn install --immutable

COPY frontend/ .

ARG BASE_URL
ENV BASE_URL=$BASE_URL

RUN yarn build

# Build Backend
FROM golang:1.24-alpine AS backend-builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .

RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o server ./cmd/server

# Final Image
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata nginx

WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /app/server .

# Copy frontend build (Nuxt generates to .output/public for static)
COPY --from=frontend-builder /app/.output/public /usr/share/nginx/html

# Copy nginx config
COPY docker/nginx.conf /etc/nginx/http.d/default.conf

# Create data directory
RUN mkdir -p /app/data

# Create startup script
RUN echo '#!/bin/sh' > /app/start.sh && \
    echo 'nginx' >> /app/start.sh && \
    echo 'exec ./server' >> /app/start.sh && \
    chmod +x /app/start.sh

# Environment variables
ENV PORT=8080
ENV DB_PATH=/app/data/formera.db
ENV JWT_SECRET=change-me-in-production
# BASE_URL should be set at runtime (e.g., https://forms.example.com)
# CORS_ORIGIN is optional - defaults to BASE_URL if not set

EXPOSE 80

VOLUME ["/app/data"]

CMD ["/app/start.sh"]
