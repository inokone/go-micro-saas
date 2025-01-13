# Go-micro-SAAS

A template for Micro SAAS applications, with a variety of implemented and planned [features](/docs/featureset.md).

## Environment setup

Before starting backend development the following need to be set up:

```sh
brew install go                                    # Install Go

go install github.com/swaggo/swag/cmd/swag@latest  # OpenAPI spec generator
go install golang.org/x/tools/cmd/goimports@latest # Reformatting tool

brew tap golangci/tap                              # Setting source for brew, then
brew install golangci/tap/golangci-lint            # Static code anlanysis for Go
```

## How to start development

Building the application is not explicitly required for development. The following commands can be used:

```sh
go mod download                            # Download Go dependencies
~/go/bin/swag i -g cmd/app.go -o api       # Generate OpenAPI spec files
go build main.go                           # Build app

golangci-lint run                          # Run static code analysis
```

On OSX if `swag` is not working you might have to add `~/go/bin` to your PATH.

## How to run tests

```sh
go test -v ./...   # Run unit tests
```

## CI

The project has Github actions set up for every push.
Steps included

- [Backend](.github/workflows/backend-build.yaml)
  - OpenAPI re-generation
  - Build
  - Run unit tests
- [Backend Static code analysis](.github/workflows/backend-lint.yml)

## How to run the application on local environment

First, you need a running Postgres database:

```sh
docker run --name postgres --env-file configs/postgres.env -p 5432:5432 -d postgres
```

Then run the application with the following:

```sh
go run cmd/app.go --migrate --config configs/  # Launch app including DB migration
```

### Build docker image

```sh
docker build -t go-micro-saas -f deployments/Dockerfile .  # Build
```

### Run docker image in production

```sh
docker run -d --restart always -p 8080:8080 -v ~/production:/etc/microsaas --mount type=tmpfs,destination=/tmp/files,tmpfs-size=4096 --mount type=bind,source=/etc/ssl/certs,target=/etc/ssl/certs go-micro-saas &   # on compute
```

## Miscellaneous

### How to create new database migration

```sh
migrate create -ext sql -dir db/migrations -seq init_schema
```

### How to run database migrations

```sh
migrate -database "postgresql://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable" -path db/migrations up # Migrate to latest version
migrate -database "postgresql://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable" -path db/migrations down -all
```
