# wiki2book

**wiki2book** is a tool to create good-looking eBooks from one or more Wikipedia articles.

The goal is to create eBooks (EPUB files) as beautiful as real books from a given list of Wikipedia articles.
To achieve this, wiki2book contains specific treatments of Wikipedia- and website-specific content of the articles and therefore provides different results than general converters (more on this below).
This should make reading Wikipedia articles even more fun and may create a whole new readership for this awesome and imperceptibly large database of knowledge. 

eBook of the German article about astronomy on a Tolino eBook-reader:
<p align="center">
<img src="photo.JPG"/>
</p>

### Why not simply use pandoc?

Good question.

[Pandoc](https://pandoc.org/epub.html) and other converters, like [wb2pdf](https://mediawiki2latex.wmflabs.org/) or [percollate](https://github.com/danburzo/percollate), are great and yes, they can convert mediawiki to EPUB.
In fact, wiki2book relies by default on pandoc to turn HTML into EPUB because pandoc does this quite well.

However, when converting mediawiki to EPUB, there are always things missing when using these tools.
For example, the correct rendering math code, downloading and embedding images, evaluating templates or a proper handling of tables.

They are also rather general purpose and don't do any eBook-specific assumptions, e.g. ignoring ebook-unsuitable styles or Wikipedia-oriented templates.

Another feature missing in all of these tools: You cannot turn multiple articles into a ready-to-read eBook.
This also includes adding a title mage, table-of-content, custom styles, etc.

Wiki2book is a tool adressing all these issues and nice features to generate beautiful looking eBooks.

# Installation

* Arch Linux: AUR package [`wiki2book`](https://aur.archlinux.org/packages/wiki2book).
  * Default style and configs can be found in `/usr/share/wiki2book`.
* Others: See the [current releases](https://github.com/hauke96/wiki2book/releases) or [build instructions](./src#build-project).

# Usage

Currently only a CLI (_command line interface_) version of wiki2book exists, so nothing with a GUI.
Wiki2book uses configuration files, project files and CLI arguments to be configured.
Use the `--help` flag or the [documentation](./doc/configuration.md) for further information.

## Preliminaries

You need the following tools and fonts when using the default configuration and styles:

* ImageMagick (to have the `convert` command)
* Pandoc (when using the `pandoc` output driver). See notes on pandoc versions 2 and 3 below.
* DejaVu fonts in `/usr/share/fonts/TTF/DejaVuSans*.ttf` (is used by the default style in this repo but can be replaced to any other font).

When enabling the conversion of SVGs to PNGs or when using the math converter "internal", then wiki2book uses the tool `rsvg-convert` by default.

The usage of external tools can be configured, e.g. to use explicit paths to executables or to use a custom script.
See [doc/configuration](./doc/configuration.md#configure-external-tool-calls) for further details.

## CLI

The CLI contains three sub-commands that generate an EPUB file from different sources:

1. Project: `wiki2book project ./path/to/project.json`
2. Article: `wiki2book article "article name"`
3. Standalone: `wiki2book standalone ./path/to/file.mediawiki`

Use `wiki2book -h` for more information and `wiki2book <command> -h` for information on a specific command.

### Configuration

See the [config documentation](./doc/configuration.md).

### Pandoc version 2 and 3

_Only relevant when using the `pandoc` output driver._

Pandoc version 2 might internally use CSS3 parameters by default, such as the `gap` property.
This might cause problems on certain eBook readers (e.g. Tolino ones).
To overcome this, pass the argument `--pandoc-data-dir ./pandoc/data` to wiki2book, which uses a template from this repo without such problematic `gap` parameter.

Alternatively install pandoc 3, which [avoids CSS3 parameters](https://github.com/jgm/pandoc/blob/3.0/data/epub.css#L166:L169).

# Contribute

## Issues, bugs, ideas

Feel free to open [a new issue](https://github.com/hauke96/wiki2book/issues/new/choose) and filling out the issue-template.

Please keep in mind:
1. This is a hobby-project and my time is limited.
2. Things with less or no use for me personally will get a lower priority.

## Development

For building, running, testing, etc. take a look at [`src/README.md`](src/README.md).

# Long-term goals

* Create a public API and web app (#7)
* Ask Wikipedia if they want to embed/link to this tool in any way (that would be super cool :D)