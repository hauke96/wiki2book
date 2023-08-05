# Integration tests

**tl;dr** Execute `run.sh` to run all integration tests.

Next to unit tests (s. `*_test.go` files in the code), this folder contains integration tests to test the implementation with more complex scenarios.
However, some tests look like unit tests but often test the combination of several features (e.g. formatting in references or tables).

## Architecture / How the integration tests work

Real wikitext files are processed and the output is compared to an expected output.
Running a test produces the following files:

* HTML and EPUB file
* downloaded images, evaluated templates and rendered math formula
* one file with a list of all those generated and downloaded files.

The generated HTML and file list are compared to a corresponding expected file.
A test succeeds only if the actually generated and expected files are equal.

These tests require access to the Wikipedia and Wikimedia API in order to work properly (e.g. evaluating templates or downloading content).

## Files

Every test consists of several files, the source mediawiki file, the expected HTML and an expected list of all files.
The name scheme for the files is always this: `test-<name>.<ext>` (where `ext` is `mediawiki`, `html` or `filelist`).

These are the required files for each test:

* `test-<name>.mediawiki`: The actual mediawiki file that should be converted.
* `test-<name>.html`: Expected HTML file.
* `test-<name>.filelist`: List of expected file names. The format of this file is describes below. This file list ensures that all non-HTML-files are also downloaded/created.

### Format of the `.filelist` file

Each `.filelist` file contains one or more file names and a blank line at the very end.
It looks like this:

```
<file-path>
<file-path>
...
<file-path>
<blank-line>
```

The `<file-path>` is the relative path based on the `results` folder, so e.g. `results/test-bold-italic.html`.

Example:

```
results/test-bold-italic/test-bold-italic.epub
results/test-bold-italic/test-bold-italic.filelist
results/test-bold-italic/test-bold-italic.html

```

### Update expected files

Updating files is easy: Either modify them manually or -- what I recommend -- use the `update.sh` script to update the files of a specific test (e.g. `./update.sh headings` for the `test-headings.mediawiki` test file).

## Run tests

Running tests happens with the `run.sh` script.
Use `run.sh -h` for usage information.

Mediawiki files are used as basis for the test.
The generated HTML output and the file list, both written to `results/test-<name>/`, are compared to the expected HTML and file list, which are locates directly here next to each `.mediawiki` file.

If tests failed, the name and file of error are given:
For example: `references [HTML]` says that the HTML output of the `test-references.mediawiki` file had errors, while `references [filelist]` says that the actual and expected list of files differ.

### Run all tests

Just start the `run.sh` script.

### Run specific tests

Using `run.sh <names>` only runs the given names.
For example `run.sh real-article-Erde references` runs the two tests `test-real-article-Erde.mediawiki` and `test-references.mediawiki`.

### Tests failed, what now?

This can have several reasons:

* The caches contain old data not matching the latest test files (clear/reset caches as described below)
* Wikipedia changed something (most likely the way of rendering templates) and HTML files do not match anymore. In this case the test files here are outdated, just update the passages in the HTML files.
* You did something wrong when coding. Best solution is to fix your code ;)
* Bugs in the integration test framework might exist as well, please [open a GitHub issue](https://github.com/hauke96/wiki2book/issues/new) when none of the above helped.

For more information on what went wrong, take a look into the logs (s. below).

### Output and logs

The generated files (HTML, rendered math, images, ...) are in the `results/test-<name>/` folder.

Logs are in `logs/test-<name>.log`.

## Reset test folder

Use the `reset.sh` script to remove all cached files created during test execution.
This can solve failing tests because Wikipedia as well as the test files change over time.

## Create a new test

Creating a new test is easy (also take a look at existing tests):

1. Create a mediawiki file with the name scheme `test-<name>.mediawiki` and fill it with your scenarios.
2. Run the update script `./update.sh -r <some-name>` (running the test and copying the results).
3. Take a look at the `.html` and `.filelist` files to verify their correctness. If they're not correct, adjust these files so that they contain the expected content.

That's it.

### Best practices

* One feature per test (e.g. lust test headings in one test). I know, there is this `generic` test but that's more like a sandbox to me than a serious test.
* Short but descriptive names (e.g. `test-headings.mediawiki` instead of `test-42.mediawiki` or `test-headings-with-formatting-and-links-within-headings.mediawiki`).
* Add line numbers. This just makes it easier to find the failed line in the mediawiki file from an HTML diff.
* Avoid templates since their evaluation might change over time. Especially citation templates should not be used here.