This module is so that go tools installed via go get -tool, e.g. `go get -tool github.com/golangci/golangci-lint/cmd/golangci-lint@latest`, do not pollute dependencies in the root module.

## adding a tool

* cd to this direcotry
* `go get -tool tool@version`

## running a tool

You can run one of these tools from the root directory with the -modfile flag, e.g. `go tool -modfile=./tools/go.mod golangci-lint run`
