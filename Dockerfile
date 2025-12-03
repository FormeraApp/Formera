# Build Frontend (Nuxt)
FROM node:20-alpine AS frontend-builder

RUN corepack enable && corepack prepare yarn@4.9.4 --activate

WORKDIR /app

COPY frontend/package.json frontend/yarn.lock frontend/.yarnrc.yml ./
RUN yarn install --immutable || yarn install

COPY frontend/ .

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
FROM node:20-alpine

RUN apk add --no-cache ca-certificates tzdata nginx

WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /app/server .

# Copy frontend build (Nuxt SSR output)
COPY --from=frontend-builder /app/.output ./.output

# Copy nginx config
COPY docker/nginx.conf /etc/nginx/http.d/default.conf

# Create data directory
RUN mkdir -p /app/data

# Copy startup script
COPY docker/start.sh /app/start.sh
RUN chmod +x /app/start.sh

# Environment variables (internal ports, not configurable)
ENV PORT=8080
ENV NITRO_PORT=3000

EXPOSE 80

VOLUME ["/app/data"]

CMD ["/app/start.sh"]
