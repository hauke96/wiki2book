# Integration tests

## Run tests

Just start the `run.sh` script.

It basically takes all `.mediawiki` files and tests if their HTML outcome matches an expected HTML file. More details below.

Of a test fails, its HTML outcome and the expected HTML outcome are compared and the difference is shown. 
Also a list of failed tests is given.

### Output and logs

The generated files are in the `results/test-<name>/` folder.

Logs are in `logs/test-<name>.log`.

## Idea

Each mediawiki-file will be turned into an HTML and ePUB file.
Of course the embedded images, templates and math parts are downloaded and stored to disk.

The HTML file is then compared against an expected HTML file.
All other files (except the ePUB one) will be hashed and also compared against an expected list of file hashes.

The last step might fail when e.g. an image on Wikipedia changes and therefore results in a different hash value.
However this probably happens not that often.
See the below step of [updating expected files](update-expected-files) below.

## Files

Every test consists of several files, the source mediawiki file, the expedted HTML and an expedted list of all files with their hashes.
The name scheme for the files is always this: `test-<name>.<ext>` (where `ext` is `mediawiki`, `html` or `filelist`).

These are the required files for each test:

* `test-<name>.mediawiki`: The actual mediawiki file that should be converted.
* `test-<name>.html`: Expected HTML file.
* `test-<name>.filelist`: Expected list of files with their SHA256 hash values. The format of this file is describes below. This file list ensures that all non-HTML-files are also matching the expected file, but without storing all the images and stuff in this repo.

### Format of the `.filelist` file

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

Updating files is easy: Either modify them manually or -- what I recommend -- just copy the actual output file from the `results` folder and edit it if necessary.

## Create a new test

Creating a new test is easy (also take a look on existing tests):

1. Create a mediawiki file with the name scheme `test-<some-name>.mediawiki`
2. Run the tests using `./run.sh`
3. Copy the resulting `test-<some-name>.html` and `test-<some-name>.filelist` files from `results/test-<some-name>/` to this folder. I *can't* recommend creating the expected files yourself ... just copy and edit them :D

That's it.

### Best practices

* One feature per test (e.g. lust test headings in one test). I know, there is this `generic` test but thats more like a sandbox to me than a serious test.
* Short but descriptive names (e.g. `test-headings.mediawiki` instead of `test-42.mediawiki` or `test-headings-with-formatting-and-links-within-headings.mediawiki`).
* Add line numbers. This just makes it easier to find the failed line in the mediawiki file from an HTML diff.
