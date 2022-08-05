Some basic stuff to work with this project.

# Preliminaries

Make sure `go` is installed and then just import the project to your IDE.
You also need Pandoc (for the `pandoc` command) and ImageMagick (for the `convert` command) to turn HTML into ePub files.

# Build project

`go build -o wiki2book .`

# Run project

Build project (s.o.) and then `./wiki2book project your/project.json`.
Use `./wiki2book --help` for all available options.

# Run tests

You have three options:

1. Run the unit tests (s. below)
2. Run the integration tests (s. below)
3. Run both with the `tests.sh` script in the root of the repo

## Unit tests

Normal (without creating a coverage file) go into the `src` folder and execute `go test ./...`.

With coverage: Go into `src` folder, use `go test -coverprofile test.out ./...` to run tests and `go tool cover -html=test.out` to view the coverage result.

## Integration tests

In the root of the repo there's an ` integration-test` folder, it contains several standalone `.json`  files.
Take a look at the `README.md` there.