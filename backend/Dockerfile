# Build Backend
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o server ./cmd/server

# Final Image
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .

RUN mkdir -p /app/data

# Create non-root user and group, set ownership
RUN addgroup -S app && adduser -S app -G app \
  && chown -R app:app /app /app/data

ENV PORT=8080

EXPOSE 8080

VOLUME ["/app/data"]

USER app
CMD ["./server", "serve"]
