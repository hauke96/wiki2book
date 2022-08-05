Some basic stuff to work with this project.

# Preliminaries

Make sure `go` is installed and then just import the project to your IDE.
You also need `pandoc` to turn HTML into ePub files.

# Build project

`go build -o wiki2book .`

# Run project

Build project (s.o.) and then `./wiki2book project your/project.json`.
Use `./wiki2book --help` for all available options.

# Run tests

Normal (without creating a coverage file) go into the `src` folder and execute `go test ./...`.

With coverage: Go into `src` folder, use `go test -coverprofile test.out ./...` to run tests and `go tool cover -html=test.out` to view the coverage result.
