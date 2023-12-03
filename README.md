# wiki2book

**wiki2book** is a tool to create good-looking eBooks from one or more Wikipedia articles.

The goal is to create eBooks (EPUB files) as beautiful as real books from a couple of Wikipedia articles.
Therefore, wiki2book is specifically implemented to create such books by implementing awareness for Wikipedia- and website-specific features (more on that below).
This should make reading Wikipedia articles even more fun and may create a whole new readership for this awesome and imperceptibly large database of knowledge. 

<p align="center">
<img src="photo.JPG" alt="eBook of the German article about astronomy on a Tolino eBook-reader."/>
</p>

### Why not simply using pandoc?

Good question.

[Pandoc](https://pandoc.org/epub.html) and others like [wb2pdf](https://mediawiki2latex.wmflabs.org/) or [percollate](https://github.com/danburzo/percollate) as well) are great and yes, they can convert mediawiki to EPUB.
In fact, wiki2book relies on pandoc to turn HTML into EPUB because pandoc is well known and it's a simple program call.

However, there are always things missing in these tools, for example rendering math, downloading images, evaluating templates or a proper handling of tables.
They also don't do any eBook-specific assumptions, e.g. ignoring ebook-unsuitable styles or not evaluating Wikipedia-oriented templates.

Most existing tools are furthermore rather general purpose, which is not beneficial for the very specific task of converting Wikipedia articles to beautiful offline eBooks.

Another feature missing in all of these tools: You cannot turn multiple articles into a ready-to-read eBook.
But wiki2book has exactly this functionality called "projects" as described below.

# Installation

* Arch Linux: AUR package [`wiki2book`](https://aur.archlinux.org/packages/wiki2book).
* Others: See the [build instructions](./src#build-project).

# Usage

Currently only a CLI version of wiki2book exists.

## Preliminaries

You need the following tools and fonts:

1. ImageMagick (to have the `convert` command)
2. Pandoc (to have the `pandoc` command). See notes on pandoc versions 2 and 3 below.
3. *Optional:* DejaVu fonts in `/usr/share/fonts/TTF/DejaVuSans*.ttf`
    * The DejaVuSans font is used by the default style in this repo but can be replaced to any other font.

## CLI

The current CLI is pretty simple and has three sub-commands:

1. Project: `wiki2book project ./path/to/project.json`
2. Article: `wiki2book article "article name"`
3. Standalone: `wiki2book standalone ./path/to/file.mediawiki`

Use `wiki2book -h` for more information and `wiki2book <command> -h` for information on a specific command.

### Configuration

Next to the project file (s. below), the application reads technical, project-independent and basic configurations from a JSON file (e.g. the templates to ignore), which can be specified with `--config, -c <file>`.
See [configs/de.json](configs/de.json) for an example and [src/config/config.go](src/config/config.go) for all technical details on each possible value including their defaults.

Some properties can be configured in both, the project and configuration file (such as the Wikipedia URL).
Entries from the project file are used in case a property is given in both files.

Also take a look at the [config.go](src/config/config.go) source file, which contains a lot of documentation on each config entry.

### Project file

When using a project, the above-mentioned `project.json` is a configuration for this project, containing e.g. the title, cover image and list of articles, and may look like this:

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
  "cover": "cover.png",
  "style": "style.css",
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

There are some optional entries:

* `cache-dir` (has the default value `.wiki2book`)
* `output-type` (has the default value `epub2`)
* `font-files`
* `wikipedia-instance` (this value overrides the general configuration (s. above) when given)

#### Use a different Wikipedia instance

Per default, the english wikipedia (`en`) is used.
However, you can change the `wikipedia-instance` entry in your projects or config file (s. above; project entries take precedence over configuration entries).
Notice, that you also have to adjust the list of ignore templates and all other language-specific configurations.
Take a look at the [German config file](configs/de.json) and some [German project files](projects/de/) to get an idea of a switch to a different Wikipedia instance.

### Pandoc version 2 and 3

Pandoc version 2 might internally use CSS3 parameters by default, such as the `gap` property.
This might cause problems on certain eBook readers (e.g. Tolino ones).
To overcome this, pass the argument `--pandoc-data-dir ./pandoc/data` to wiki2book, which uses a template from this repo without such problematic `gap` parameter.

Alternatively install pandoc 3, which [avoids CSS3 parameters](https://github.com/jgm/pandoc/blob/3.0/data/epub.css#L166:L169).

### Examples

In the following there are working example calls to wiki2book.

The necessary parameters used below (see `./wiki2book -h` for more information):

* `-c`: Wiki2book configuration file
* `-s`: Specifies an existing style sheet file

#### Project

Use the following command to build the German project about astronomy:

`./wiki2book project -c configs/de.json ./projects/de/astronomie/astronomie.json`

#### Single article

Render a single article by using the `article` sub-command:

`./wiki2book article -c configs/de.json -s projects/style.css "Erde"`

#### Standalone

Use the following command to render the file

`./wiki2book standalone -c configs/de.json -s projects/style.css ./integration-tests/test-real-article-Erde.mediawiki`

# Development

For building, running, testing, etc. take a look at the `src` folder and `src/README.md`.

# Long-term goals

* Create a public API and web app (#7)
* Ask Wikipedia if they want to embed/link to this tool in any way (that would be super cool :D)