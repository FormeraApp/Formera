# Formera

Self-hosted form builder. Privacy-friendly alternative to Google Forms.

## Features

- Drag & Drop form builder
- Multiple field types (text, select, checkboxes, ratings, date/time, file upload, signature)
- Multi-page forms with progress indicator
- Password-protected forms
- Custom slugs for forms
- Responsive design
- Response analytics & statistics
- CSV/JSON export
- i18n support (German, English)
- Docker deployment

## Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go, Gin, GORM |
| Frontend | Nuxt 3, Vue 3, TypeScript |
| Database | SQLite |
| Storage | Local or S3-compatible |

## Quick Start

### Docker

```bash
docker run -d \
  -p 80:80 \
  -v formera-data:/app/data \
  -e BASE_URL=https://forms.example.com \
  -e JWT_SECRET=your-secure-secret-here \
  ghcr.io/formeraapp/formera:latest
```

### Docker Compose

```yaml
services:
  formera:
    image: ghcr.io/formeraapp/formera:latest
    container_name: formera
    restart: unless-stopped
    environment:
      - BASE_URL=https://forms.example.com
      - JWT_SECRET=your-secure-secret-here
      # - CORS_ORIGIN=https://other-domain.com  # Optional: only if frontend is on different domain
    volumes:
      - formera-data:/app/data
    ports:
      - "80:80"

volumes:
  formera-data:
```

```bash
docker compose up -d
```

Access at `http://localhost`. Setup wizard appears on first start.

### Development

```bash
# Backend
cd backend && go run ./cmd/server

# Frontend
cd frontend && yarn install && yarn dev
```

## Configuration

### Main

| Variable | Description | Default |
|----------|-------------|---------|
| `BASE_URL` | Public URL of the application | `http://localhost:8080` |
| `CORS_ORIGIN` | Allowed origin (optional, defaults to BASE_URL) | - |
| `JWT_SECRET` | JWT signing key (change in production!) | - |

### Server

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Backend port | `8080` |
| `DB_PATH` | SQLite database path | `./data/formera.db` |

### Storage

| Variable | Description | Default |
|----------|-------------|---------|
| `STORAGE_TYPE` | Storage type: `local` or `s3` | `local` |
| `STORAGE_LOCAL_PATH` | Local upload directory | `./data/uploads` |
| `STORAGE_LOCAL_URL` | URL prefix for local files | `/uploads` |

### S3 Storage

| Variable | Description | Default |
|----------|-------------|---------|
| `S3_BUCKET` | S3 bucket name | - |
| `S3_REGION` | S3 region | - |
| `S3_ACCESS_KEY_ID` | S3 access key | - |
| `S3_SECRET_ACCESS_KEY` | S3 secret key | - |
| `S3_ENDPOINT` | Custom endpoint (MinIO, etc.) | - |
| `S3_PREFIX` | File path prefix | `uploads/` |
| `S3_PRESIGN_MINUTES` | URL expiration time | `60` |

### Migration

| Variable | Description | Default |
|----------|-------------|---------|
| `STORAGE_MIGRATE_ON_START` | Auto-migrate local files to S3 | `true` |
| `STORAGE_DELETE_AFTER_MIGRATE` | Delete local files after migration | `false` |

### Cleanup

| Variable | Description | Default |
|----------|-------------|---------|
| `CLEANUP_ENABLED` | Enable automatic orphan file cleanup | `true` |
| `CLEANUP_INTERVAL_HOURS` | Cleanup interval | `24` |
| `CLEANUP_MIN_AGE_DAYS` | Minimum file age before deletion | `7` |
| `CLEANUP_DRY_RUN` | Only log deletions, don't execute | `false` |

### SEO

| Variable | Description | Default |
|----------|-------------|---------|
| `NUXT_PUBLIC_INDEXABLE` | Allow search engine indexing | `true` |

## Testing

```bash
# Run all backend tests
cd backend && go test -v ./...

# Run tests with coverage
cd backend && go test -v -cover ./...

# Run specific package tests
cd backend && go test -v ./internal/handlers/...
```

## Security

If you discover a security vulnerability within Formera, please send an e-mail to admin@formera.app

All reports will be promptly addressed and you'll be credited in the fix release notes.

## Contributing

Formera is a free and open source project licensed under the MIT License. You are free to do whatever you want with it, even offering it as a paid service.

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[MIT](LICENSE.md)
