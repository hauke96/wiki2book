# Integration tests

## Run tests

Just start the `run.sh` script.

It basically takes all `.mediawiki` files and tests if their outcome matches an expected HTML file. More details below.

Of a test fails, its HTML outcome and the expected HTML outcome are compared and the difference is shown. 
Also a list of failed tests is given at the end.

## Idea

A mediawiki-file will be turned into an HTML and ePUB file.
The HTML file is then compared against an expected HTML file.
All other files (except the ePUB one) will be hashed als also compared against an expected list of file hashes.

The last step might fail when e.g. an image on Wikipedia changes and therefore results in a different hash value. However this probably happens not that often.

## Files

Every test consists of the following files.
The name scheme is always this: `test-<name>.<ext>` (where `ext` is `mediawiki`, `html` or `filelist`).

* `test-<name>.mediawiki`: The actual mediawiki file that should be converted.
* `test-<name>.html`: Expected HTML file.
* `test-<name>.filelist`: Expected list of files with their SHA256 hash values. The format of this file is describes below. This file list ensures that all non-HTML-files are also matching the expected file, but without storing all the images and stuff in this repo.

### Format of file list

Each `.filelist` file contains one or more of the following lines:

```
<sha256-hash-of-file><space><space><file-path>
```

The `<file-path>` is the relative path based on the `results` folder, so e.g. `results/test-bold-italic.html`.

Example:

```
e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  results/test-bold-italic/test-bold-italic.filelist
a43dfb291002d12831921a28cef83987ce18f68953c691c7bff49d1b09ec7a63  results/test-bold-italic/test-bold-italic.html
```

### Update expected files

Updating files is easy: Either modify then manually or just copy the actual output file from the `results` folder.

## Create a new test

Creating a new test is easy (just take a look on existing tests):

1. Create a mediawiki file with the name scheme `test-<some-name>.mediawiki`
2. Run the tests using `./run.sh`
3. Copy the resulting `test-<some-name>.html` and `test-<some-name>.filelist` files from `results/test-<some-name>/` to this folder. I *don't* recommend creating the expected files by hand ... just copy and adjust them if needed :D

That's it.

### Best practices

* Just test one feature in one test (e.g. lust headings in one test)
* Use short but descriptive names (e.g. `test-headings.mediawiki`)
* Add line numbers. This just makes it easier to find the failed line in the mediawiki file from an HTML diff.
