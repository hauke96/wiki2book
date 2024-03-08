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
  * Default style and configs can be found in ` /usr/share/wiki2book`.
* Others: See the [current releases](https://github.com/hauke96/wiki2book/releases) or [build instructions](./src#build-project).

# Usage

Currently only a CLI (_command line interface_) version of wiki2book exists, so nothing with a GUI.
Wiki2book need a configuration file (s. the [configs](./configs) folder), currently only a German config file exists.

## Preliminaries

You need the following tools and fonts:

1. ImageMagick (to have the `convert` command)
2. Pandoc (to have the `pandoc` command). See notes on pandoc versions 2 and 3 below.
3. *Optional:* DejaVu fonts in `/usr/share/fonts/TTF/DejaVuSans*.ttf`
    * The DejaVuSans font is used by the default style in this repo but can be replaced to any other font.

## CLI

The CLI contains three sub-commands that generate an EPUB file from different sources (s. below for examples and details on each sub-command):

1. Project: `wiki2book project ./path/to/project.json`
2. Article: `wiki2book article "article name"`
3. Standalone: `wiki2book standalone ./path/to/file.mediawiki`

Use `wiki2book -h` for more information and `wiki2book <command> -h` for information on a specific command.

### Configuration

See the [config documentation](./doc/configuration.md).

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

# Contribute

## Issues, bugs, ideas

Feel free to open [a new issue](https://github.com/hauke96/wiki2book/issues/new/choose).
But keep in mind:
This is a hobby-project and my time is limited.
Things with less or no use for me personally will get a lower priority.

## Development

For building, running, testing, etc. take a look at [`src/README.md`](src/README.md).

# Long-term goals

* Create a public API and web app (#7)
* Ask Wikipedia if they want to embed/link to this tool in any way (that would be super cool :D)