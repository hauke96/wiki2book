# Build project

`go build .`

# Run project

Either `go run .` or (after the previous build step) `./main your/project.json`

## Run test project

Use `test` as project name, so e.g. `go run . test`.

# Run tests

Normal, without coverage file: `go test ./...`

With coverage: `go test -coverprofile test.out ./... && go tool cover -html=test.out`