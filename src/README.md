Some basic stuff to work with this project.

# Preliminaries

Make sure `go` is installed (s. minimal go-version in `go.mod`) and then just import the project to your IDE.
You also need Pandoc (for the `pandoc` command) and ImageMagick (for the `convert` command) to turn HTML into EPUB files and to process images.

# Build project

`go build -o wiki2book .`

Or `build.sh` (use parameter `-h` for help) in root of this repo.

# Run

To test if everything works, run `go run . -c ../configs/de.json article Erde` (it builds an eBook from the German "Erde" article).

Use `go run --help` for all available CLI options.

# Linting

I use `golangci-lint` since it combines multiple linters.

1. Install it `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
2. Use it `golangci-lint run`

As you may see, the logs are spammed with error messages.
Not all of them are relevant, some are false-positives, some do not relate to this project but to dependencies, but in general, the messages give you a hint on what could be improved.

The configuration `.golangci-lint` contains the list of used linters and some additional configuration of them.

# Run tests

You have three options:

1. Run the unit tests (s. below)
2. Run the integration tests (s. below)
3. Run both with the `tests.sh` script in the root of the repo

## Unit tests

Normal (without creating a coverage file) go into the `src` folder and execute `go test ./...`.

With coverage: Go into `src` folder, use `go test -coverprofile test.out ./...` to run tests and `go tool cover -html=test.out` to view the coverage result.

Of course IDEs like Goland provide direct possibility to run the unit tests with and without coverage.

## Integration tests

In the root of the repo there's an ` integration-test` folder, it contains several standalone `.json` files.
Take a look at the `README.md` there.

# Performance measurement

## Profiling

### Method 1: IDE

Use your IDE with the `main_test.go` file, which only exists to allow IntelliJ to profile go-code.

### Method 2: CLI & pprof

* Run wiki2book with the `--diagnostics-profiling` flag to generate a `profiling.prof` file.
* Run `go tool pprof <wiki2book-executable> ./profiling.prof` so that the `pprof` console comes up.
* Enter `web` for a browser or `evince` for a PDF visualization

You can also pass all arguments to the CLI directly.
For me `go tool pprof -nodecount=1000 -call_tree -pdf wiki2book ./profiling.prof` works quite well.

## `time` command tricks

The following command uses the `time` command to print the complete execution time, takes the `real` time and prints the number in milliseconds.
This can easily be used in a loop to do some semi-professional performance measurement.

```
time ./wiki2book article -r "DB Class VT 11.5"; } |& grep real | sed -E 's/[^0-9\.]+//g' | bc
```