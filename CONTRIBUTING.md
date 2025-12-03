# Contributing to Formera

Thank you for your interest in contributing to Formera!

## Getting Started

1. Fork the repository
2. Clone your fork
3. Create a new branch for your feature or fix

```bash
git clone https://github.com/YOUR_USERNAME/formera.git
cd formera
git checkout -b feature/your-feature-name
```

## Development Setup

### Prerequisites

- Go 1.24+
- Node.js 20+
- Yarn

### Backend

```bash
cd backend
cp ../.env.example ../.env
go mod download
go run ./cmd/server
```

### Frontend

```bash
cd frontend
yarn install
yarn dev
```

## Code Style

### Go

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `golangci-lint` before committing

### TypeScript/Vue

- Use Biome for linting and formatting
- Run `yarn lint` before committing

## Commit Messages

Use clear, descriptive commit messages:

```
feat: add password protection for forms
fix: resolve date parsing issue in submissions
docs: update configuration documentation
```

## Pull Requests

1. Ensure all tests pass
2. Update documentation if needed
3. Keep PRs focused on a single feature or fix
4. Reference related issues in the PR description

## Reporting Issues

- Check existing issues before creating a new one
- Provide clear steps to reproduce bugs
- Include environment details (OS, browser, versions)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
