Some basic stuff to work with this project.

# Preliminaries

Make sure `go` is installed and then just import the project to your IDE.

# Build project

`go build .`

# Run project

Either `go run .` or (after the previous build step) `./main your/project.json`

## Run test project

Use `go run . test` to compile the test-project `./test/`.
This is just a local wikitext file, so no content is downloaded from Wikipedia which is a) fater and b) doesn't require an internet connection.¹

# Run tests

Normal, without coverage file: `go test ./...`

With coverage: `go test -coverprofile test.out ./... && go tool cover -html=test.out`

---
¹ Especially useful when using trains in Germany. No joke.