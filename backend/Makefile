.PHONY: build run test swagger docs clean

# Build the application
build:
	go build -o bin/server ./cmd/server

# Run the application
run:
	go run ./cmd/server

# Run tests
test:
	go test -v ./...

# Generate Swagger documentation
swagger:
	@which swag > /dev/null 2>&1 || go install github.com/swaggo/swag/cmd/swag@latest
	$(shell go env GOPATH)/bin/swag init -g cmd/server/main.go -o docs --parseInternal --parseDependency
	@echo "Swagger docs generated in docs/"

# Update API docs for GitLab Pages
docs: swagger
	cp docs/swagger.json ../api-docs/
	@echo "API docs updated in ../api-docs/"

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf docs/

# Install development dependencies
deps:
	go mod download
	go install github.com/swaggo/swag/cmd/swag@latest

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run
