# Changelog

All notable changes to this project will be documented in this file.

## [0.0.1] - 2025-12-03

### Initial Release

Self-hosted form builder. Privacy-friendly alternative to Google Forms.

### Features

#### Form Builder
- Drag & drop form builder with live preview
- 15+ field types: Text, Textarea, Number, Email, Select, Multi-Select, Checkbox, Radio, Rating, Scale, Date/Time, File Upload, Signature, Rich Text, Heading, Paragraph, Divider, Section
- Multi-page forms with progress indicator
- Custom form slugs for clean URLs
- Password-protected forms
- Form duplication

#### Responses & Analytics
- Real-time submission tracking
- Response statistics and analytics
- Export to CSV and JSON
- Individual response view

#### Storage
- Local file storage (default)
- S3-compatible storage support (AWS S3, MinIO, etc.)
- Automatic file migration between storage backends
- Orphaned file cleanup

#### Security & Privacy
- JWT authentication
- Self-hosted - your data stays on your server
- No tracking, no analytics sent to third parties

#### Deployment
- Single Docker image with everything included
- Docker Compose support
- Simple configuration via environment variables

#### Internationalization
- English and German language support
- Browser language detection

### Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.24, Gin, GORM |
| Frontend | Nuxt 4, Vue 3, TypeScript, Tailwind CSS 4 |
| Database | SQLite |
| Storage | Local / S3-compatible |
