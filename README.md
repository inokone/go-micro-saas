# Go-micro-SAAS

A template for building Micro-SAAS (Software as a Service) applications in Go. See our [feature list](/docs/featureset.md) for implemented and planned capabilities.

Before getting started, make sure to review and set up the required [third-party services and API keys](PREREQUISITES.md).

## Development Setup

### Prerequisites

Install the following tools:

```sh
# Install Go
brew install go

# Install development tools
go install github.com/swaggo/swag/cmd/swag@latest    # OpenAPI spec generator
go install golang.org/x/tools/cmd/goimports@latest   # Code formatting tool
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest  # Database migration tool

# Install linter
brew tap golangci/tap
brew install golangci/tap/golangci-lint
```

Note: On macOS, if `swag` command is not found, add `~/go/bin` to your PATH.

### Development Workflow

1. Install dependencies:

```sh
go mod download
```

2. Generate OpenAPI specification:

```sh
~/go/bin/swag init -g cmd/app.go -o api
```

3. Run tests:

```sh
go test -v ./...
```

4. Run linter:

```sh
golangci-lint run
```

### Running Locally

1. Start Postgres database:

```sh
docker run --name postgres --env-file configs/postgres.env -p 5432:5432 -d postgres
```

2. Run the application:

```sh
# Run with database migrations
go run cmd/app.go --migrate --config configs/
```

## Database Management

### Creating New Migrations

```sh
migrate create -ext sql -dir db/migrations -seq migration_name
```

### Running Migrations Manually

```sh
# Migrate up to latest version
migrate -database "postgresql://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable" -path db/migrations up

# Rollback all migrations
migrate -database "postgresql://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable" -path db/migrations down -all
```

## Docker Support

### Building Image

```sh
docker build -t go-micro-saas -f deployments/Dockerfile .
```

### Production Deployment

```sh
docker run -d \
  --restart always \
  -p 8080:8080 \
  -v ~/production:/etc/microsaas \
  --mount type=tmpfs,destination=/tmp/files,tmpfs-size=4096 \
  --mount type=bind,source=/etc/ssl/certs,target=/etc/ssl/certs \
  go-micro-saas
```

## Continuous Integration

The project uses GitHub Actions for CI/CD:

- [Backend Pipeline](.github/workflows/backend-build.yaml)
  - OpenAPI specification generation
  - Build verification
  - Unit tests
- [Code Quality](.github/workflows/backend-lint.yml)
  - Static code analysis

## Documentation

- [Feature Set](/docs/featureset.md)
- [Deployment Guide](/docs/deployment.md)
