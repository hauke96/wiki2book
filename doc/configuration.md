It's possible to configure all kinds of things in wiki2book.
This page describes each possible setting and config entry.

# Overview

There are three different places for configurations:
Configuration file, project file and CLI arguments.

They have a large overlap in settings and are evaluated in this order.
This means a settings from the configuration file can be overwritten by the same settings of a project file and such a setting can be overwritten by a CLI argument.

# Configuration file

A configuration file is a JSON file and might look like this:

```json
{
  "cache-dir": "./path/to/cache/",
  "wikipedia-instance": "de",
  "output-type": "epub3",
  "output-driver": "internal"
}
```

Take a look at the `configs/de.json` file for an example file.
The `src/config/config.go` file contains all fields including documentation and default values.
The CLI command `wiki2book --help` will also show information and default values for the general configuration entries.

# Project files

A project file is also a JSON file and can contain all the properties of a configuration file plus some additional entries, which are the following ones:

* A `"metadata": {...}` object containing the following entries:
  * `"title"`: The title of the book.
  * `"language"`: The language of the book.
  * `"author"`: The author of the book.
  * `"license"`: The license of the book, which should be based on the Wikipedia articles licenses.
  * `"date"`: The date of the article.
* The `"output-file"`, which is a path to the output EPUB file.
* The `"articles": [...]` array, which is a list of article names, that should be included into this book.

The following example contains project-specific entries at the top and general config entries at the bottom.

```json
{
  "metadata": {
    "title": "My great book",
    "author": "Wikipedia contributors",
    "license": "Creative Commons Non-Commercial Share Alike 3.0",
    "language": "de-DE",
    "date": "2021-12-27"
  },
  "output-file": "my-book.epub",
  "articles": [
    "Hamburg",
    "Hamburger",
    "Pannfisch"
  ],
  
  "wikipedia-instance": "de",
  "cache-dir": "./path/to/cache/",
  "output-type": "epub3",
  "output-driver": "internal",
  "cover-image": "cover.png",
  "style-file": "style.css",
  "pandoc-data-dir": "./pandoc/data",
  "font-files": [
    "/path/to/font.ttf",
    "/path/to/fontBold.ttf",
    "/path/to/fontItalic.ttf"
  ]
}
```

Take a look at the `projects` folder for further examples.

# Use a different Wikipedia instance

Per default, the english wikipedia (`en`) is used.
However, you can change the `wikipedia-instance` entry in your projects or config file (s. above).
Notice, that you also have to adjust the list of ignore templates and all other language-specific configurations.
Take a look at the [German config file](../configs/de.json) and some [German project files](../projects/de/), to get an idea of a switch to a different Wikipedia instance.

# Configure external tool calls

Call to external tools can be configured using simple command template strings.
Use the `--help` flag to find all command templates, they all start with `command-template-...`.

The [`config.go`](../src/config/config.go) file contains the default values for different commands.

Writing your own command template is simple, here are some examples for your config or project file.

Using inkscape to convert SVGs to PNGs:
```json
"command-template-svg-to-png": "inkscape {INPUT} -o {OUTPUT}",
```

Using a custom script to convert SVGs to PNGs:
```json
"command-template-svg-to-png": "/path/to/your/scripts/svg-to-png.sh -i {INPUT} -o {OUTPUT} --some --other --params",
```

Using the CLI args works in the same way:
`wiki2book --command-template-svg-to-png "inkscape {INPUT} -o {OUTPUT}"`