# go-realworld

* Golang implementation of the [realworld api spec](https://github.com/realworld-apps/realworld)
* The project structure is inspired by [polylith](https://polylith.gitbook.io/polylith/)
	* `pkg/` contains [components](https://polylith.gitbook.io/polylith/architecture/2.3.-component)
	* `cmd/` contains [bases](https://polylith.gitbook.io/polylith/architecture/2.2.-base)
	* `playground/` contains the [development project](https://polylith.gitbook.io/polylith/architecture/2.4.-development)

## Table of Contents

1. [Stack](#stack)
1. [Developer Guide](#developer-guide)
	1. [Dependencies](#dependencies)
    1. [Setting up the dev environment](#setting-up-dev-environment)
    1. [Running the http server](#running-the-http-server)
    1. [Database migrations](#database-migrations)
    1. [Openapi code generation](#openapi-code-generation)
    1. [Tests](#tests)
    1. [Playground](#playground)
    1. [Git hooks](#git-hooks)

## Stack

* [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen/) - openapi spec to golang code generator
* [fx](https://github.com/uber-go/fx) - dependency injection
* [bob](https://github.com/stephenafamo/bob) - sql query builder
* [goose](https://github.com/pressly/goose) - sql migrations

## Developer Guide

### Dependencies

* go 1.26.0 - https://go.dev/doc/install
* docker - https://www.docker.com/products/docker-desktop/

### Setting up dev environment

* run `bin/go-mod-tidy.sh` to download first time go deps
* run `docker compose up -d` to start docker services in background
* run database migrations, see [migrations](#database-migrations)

### Running the http server

* run `bin/run-local-server.sh`
* this will start the API on `http://localhost:8080`
* you can access a swagger web ui deployed via docker, at `http://localhost:8081`

> For auth, use the login or create user route, and grab the token from the response. Then you can click the Authorize button in the swagger ui to set the token `Token <from-response>` for future requests

### Database migrations

* for simplicity, to run all local migrations available, run `./bin/migrate-local.sh`
* full migration runner instructions can be seen by running the command line tool `go run cmd/migrations/main.go`

> View migration files at `pkg/database/migrations`

### Openapi code generation

* go code is generated from the open api spec `pkg/api_gen/api.yaml`
* run `bin/oapi-codegen.sh`
* the generated code is outputted to `pkg/api_gen`

### Tests

* run `go test ./...`

> Go makes use of its build cache to skip running tests for packages that haven't had changes. You can force all tests with `go test -count=1 ./...` 

#### Parallel tests

* Just about all sub tests in this project are written as parallel tests
* To mark a test case as safe for parallelization, call `t.Parallel()` within the test case
* Work was done to allow for parallization of integration tests that use the test postgres docker container
	* tests use `pkg/test_utils/db_config_provider`, which is safe for concurrent use, and each caller gets an isolated version of each test database, with all migrations applied

#### Mocks

* this project uses https://vektra.github.io/mockery
* to mark an interface for mock generation, use the comment `//mockery:generate: true`
* to generate mocks, run `bin/generate-mocks.sh`
* mocks will be generated in a `mocks/` subfolder in the package of the interface, e.g. `pkg/user/types/mocks`

## Playground

* inspired by the polylith [development project](https://polylith.gitbook.io/polylith/architecture/2.4.-development)
* the `playground/` folder sets up a dependency tree using the local config, and gives each developer a playground like experience for tinkering with the system
* e.g. run `go run playground/nimaeskandary/main.go`
* something like this could be extended if desired to load configs for real enviroments like staging or production, for use cases such as manual QA or client support. E.g. execute real code paths of the system using client data

## Git hooks

> Note, you can always skip hooks with the flag `--no-verify`, e.g. `git commit --no-verify`, `git push --no-verify`

### .git/hooks/pre-commit

use this hook to lint/fmt staged files

```bash
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT_ROOT_DIR="${SCRIPT_DIR}/../.."

GO_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

if [[ -z "$GO_FILES" ]]; then
  exit 0
fi

cd "${PROJECT_ROOT_DIR}"
./bin/lint-new.sh

git add $GO_FILES
```
