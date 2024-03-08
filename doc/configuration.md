It's possible to configure all kinds of things in wiki2book.
This page describes each possible setting and config entry.

## Overview

There are three different places for configurations and they sometimes overlap.

* **Configuration file.**
  This file contains very general settings, e.g. the wikipedia instance or a list of templates to ignore. Some things, like the list of ignored templates, are too large to be configured everytime using CLI arguments. Take a look at the [config.go](../src/config/config.go) source file, which contains a lot of documentation on each config entry. See [configs/de.json](../configs/de.json) for an example.
* **CLI arguments.**
  Can be used to configure things that differ from execution to execution. Use the CLI flag `-h` for further information on the available arguments for a specific command.
* **Project file.**
  Is used to configure project-specific things, e.g. the cover image. See [below](#project-file) or [project.go](../src/project/project.go) for more information. See [astronomie.json](../projects/de/astronomie/astronomie.json) for an example.

## Configuration

| Config entry                     | Config file | Project file | CLI arg ¹ | Default ² ³                                                          | Description                                                                                                                                                            |
|----------------------------------|-------------|--------------|-----------|----------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `IgnoredTemplates`               | X           |              |           | `[]`                                                                 | List of template names to be ignored.                                                                                                                                  |
| `TrailingTemplates`              | X           |              |           | `[]`                                                                 | List of templates that should be moved to the end of the document.                                                                                                     |
| `IgnoredImageParams`             | X           |              |           | `[]`                                                                 | List of image parameters to be ignored, e.g. `alt=...`.                                                                                                                |
| `IgnoredMediaTypes`              | X           |              |           | `[ "gif", "mp3", "mp4", "pdf", "oga", "ogg", "ogv", "wav", "webm" ]` | List of media file types to be ignored, e.g. `gif`.                                                                                                                    |
| `WikipediaInstance`              | X           | X            |           | `"en"`                                                               | Subdomain of Wikipedia instance to use, e.g. `de` or `en`.                                                                                                             |
| `WikipediaImageArticleInstances` | X           |              |           | `[ "commons", "en" ]`                                                | Subdomain of Wikipedia to download images from.                                                                                                                        |
| `FilePrefixe`                    | X           |              |           | `[ "file", "image", "media" ]`                                       | Prefixed of links considered to be files, e.g. `file`.                                                                                                                 |
| `AllowedLinkPrefixes`            | X           |              |           | `[ "arxiv", "doi" ]`                                                 | Allowed prefixed of special links, such as `arxiv:foobar`.                                                                                                             |
| `CategoryPrefixes`               | X           |              |           | `[ "category" ]`                                                     | Prefix of categories, e.g. `category`.                                                                                                                                 |
| `OutputFile`                     |             | X            | X         | (no default, must be specified)                                      | EPUB filename, e.g. `my-book.epub`.                                                                                                                                    |
| `OutputType`                     |             | X            | X         | `"epub2"`                                                            | Type of EPUB. Allowed: `epub2`, `epub3`.                                                                                                                               |
| `CacheDir`                       |             | X            | X         | `.wiki2book`                                                         | Folder where everything downloaded will be cached.                                                                                                                     |
| `StyleFile`                      |             | X            | X         |                                                                      | CSS file to style the ebook. Used in the pandoc `--css` argument.                                                                                                      |
| `CoverImage`                     |             | X            | X         |                                                                      | Cover image file to use. Used in the pandoc `--epub-cover-image` argument.                                                                                             |
| `PandocDataDir`                  |             | X            | X         |                                                                      | Folder of additional pandoc configurations. Used in the pandoc `--data-dir` argument.                                                                                  |
| `FontFiles`                      |             | X            | X         |                                                                      | Full path to font files. Used in the pandoc `--epub-embed-font` argumemnt.                                                                                             |
| `ImagesToGrayscale`              |             | X            | X         | `false`                                                              | When `true`, raster images are converted to grayscale. Note: Images are cached, then this command has been `true` in a previous run, then the images remain grayscale. |
| `Metadata.Title`                 |             | X            |           |                                                                      | Title of the ebook. Used in the pandoc `--metadata` arguments.                                                                                                         |
| `Metadata.Language`              |             | X            |           |                                                                      | Language of the book. Used in the pandoc `--metadata` arguments.                                                                                                       |
| `Metadata.Author`                |             | X            |           |                                                                      | Name of the author(s). Used in the pandoc `--metadata` arguments.                                                                                                      |
| `Metadata.License`               |             | X            |           |                                                                      | License of the book. Used in the pandoc `--metadata` arguments.                                                                                                        |
| `Metadata.Date`                  |             | X            |           |                                                                      | Publishing/creation date of the book. Used in the pandoc `--metadata` arguments.                                                                                       |
| `Logging`                        |             |              | X         |                                                                      | Sets the log level. Possible values are `debug` and `trace`.                                                                                                           |
| `DiagnosticsProfiling`           |             |              | X         |                                                                      | Enable profiling and write results to `./profiling.prof`.                                                                                                              |
| `DiagnosticsTrace`               |             |              | X         |                                                                      | Enable tracing to analyse memory usage and write results to `./trace.out`.                                                                                             |
| `ForceRegenerateHtml`            |             |              | X         |                                                                      | Forces wiki2book to recreate HTML files even if they exists from a previous run.                                                                                       |
| `SvgSizeToViewbox`               |             |              | X         |                                                                      | Sets the 'width' and 'height' property of an SimpleSvgAttributes image to its viewbox width and height. This might fix wrong SVG sizes on some eBook-readers.          |

¹ Depending on the used command, there might exist additional arguments. Use `wiki2book <command> -h` for further detail.<br>
² If no entry is given, then no default exists, but the property is not mandatory.<br>
³ The `project` command doesn't have default values for the CLI arguments.

### Precedences

The precedence of properties is as in the table above:
First, the config-file is used, then entries might be overridden by project files, which can be overridden by CLI arguments.

## Project file

When using a project, the above-mentioned project file is a JSON configuration containing e.g. the title, cover image and list of articles, and may look like this:

```json
{
  "metadata": {
    "title": "My great book",
    "author": "Wikipedia contributors",
    "license": "Creative Commons Non-Commercial Share Alike 3.0",
    "language": "de-DE",
    "date": "2021-12-27"
  },
  "cache-dir": "./path/to/cache/",
  "wikipedia-instance": "de",
  "output-file": "my-book.epub",
  "output-type": "epub3",
  "cover-image": "cover.png",
  "style-file": "style.css",
  "pandoc-data-dir": "./pandoc/data",
  "articles": [
    "Hamburg",
    "Hamburger",
    "Pannfisch"
  ],
  "font-files": [
    "/path/to/font.ttf",
    "/path/to/fontBold.ttf",
    "/path/to/fontItalic.ttf"
  ]
}
```

See the table above for mandatory entries, default values and possibilities to override these.

### Use a different Wikipedia instance

Per default, the english wikipedia (`en`) is used.
However, you can change the `wikipedia-instance` entry in your projects or config file (s. above).
Notice, that you also have to adjust the list of ignore templates and all other language-specific configurations.
Take a look at the [German config file](../configs/de.json) and some [German project files](../projects/de/), to get an idea of a switch to a different Wikipedia instance.
