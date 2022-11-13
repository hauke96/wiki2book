# Integration tests

## Run tests

Just start the `run.sh` script.

It basically takes all `.mediawiki` files and tests if their HTML outcome matches an expected HTML file. More details below.

Of a test fails, its HTML outcome and the expected HTML outcome are compared and the difference is shown. 
Also a list of failed tests is given.

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

## Architecture

Each mediawiki-file will be turned into an HTML and EPUB file.
Of course the embedded images, templates and math parts are downloaded and stored to disk.

The HTML file is then compared against an expected HTML file.
All file names will be then be sorted, stored in a list and compared against an expected list of file names.

See the below step of [updating expected files](update-expected-files) below.

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

## Create a new test

Creating a new test is easy (also take a look at existing tests):

1. Create a mediawiki file with the name scheme `test-<some-name>.mediawiki`
2. Run the update script `./update.sh <some-name>`
3. Take a look at the `.html` and `.filelist` file to check if everything is alright. If not, adjust these files so that they contain the expected content.

That's it.

### Best practices

* One feature per test (e.g. lust test headings in one test). I know, there is this `generic` test but thats more like a sandbox to me than a serious test.
* Short but descriptive names (e.g. `test-headings.mediawiki` instead of `test-42.mediawiki` or `test-headings-with-formatting-and-links-within-headings.mediawiki`).
* Add line numbers. This just makes it easier to find the failed line in the mediawiki file from an HTML diff.
