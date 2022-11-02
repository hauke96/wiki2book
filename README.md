# wiki2book

**wiki2book** is a tool to create good-looking EPUB-eBooks from one or more Wikipedia articles.

The goal of this converter is to create nearly print-ready eBooks from a couple of Wikipedia articles.
This should make reading Wikipedia articles even more fun and may create a whole new readership for this awesome and imperceptibly large database of knowledge. 

<p align="center">
<img src="photo.JPG" alt="eBook of the German article about astronomy."/>
</p>

### Why not simply using pandoc?

Good question.
Pandoc (and a lot of other tools as well) is great and yes, it can convert mediawiki to ePUB.
In fact, this converter relies heavily on pandoc because turning HTML into ePUB works perfectly.
However, there are a lot of things missing, for example rendering math but more importantly downloading images and evaluating templates.
Also pandoc doesn't do any eBook specific assumptions, e.g. ignoring ebook-unsuitable styles or not evaluating Wikipedia oriented templates.
A lot of existing tools are furthermore tied to their implementation as non-specific but rather general purpose tool, which is not beneficial when converting Wikipedia articles to eBooks. 

# Usage

## Preliminaries

You need the following tools and fonts:

1. ImageMagick (to have the `convert` command)
2. Pandoc (to have the `pandoc` command)
3. DejaVu fonts in `/usr/share/fonts/TTF/DejaVuSans*.ttf` (currently hard-coded, s. TODOs below)

## CLI

The current CLI is pretty simple: `wiki2book project ./path/to/project.json`

You can also use a single article name (`wiki2book article "The article name"`) or a local mediawiki file (`wiki2book standalone ./the/mediawiki/file.txt`).

When using a project, the mentioned `project.json` is a configuration for a project and may look like this:

```json
{
  "metadata": {
    "title": "My great book",
    "author": "Wikipedia contributors",
    "license": "Creative Commons Non-Commercial Share Alike 3.0",
    "language": "de-DE",
    "date": "2021-12-27"
  },
  "caches": {
    "images": "images",
    "templates": "templates",
    "math": "math",
    "articles": "articles"
  },
  "wikipedia-domain": "de",
  "output-file": "my-book.epub",
  "cover": "cover.png",
  "style": "style.css",
  "articles": [
    "Hamburg",
    "Hamburger",
    "Pannfisch"
  ]
}
```

The `caches` object is completely optional and in this example the default values are shown.
All values are folders, which don't need to exist, they will be created.

**Notice:** Currently only the German Wikipedia is supported.
However, you can specify `en` as `wikipedia-domain` to download articles from the English Wikipedia.
But because a lot of German template-strings are removed while parsing, the English strings remain and result in unwanted stuff in the eBook.

### Example

Use the following command to build the German project about astronomy:

`./wiki2book project ./projects/de/astronomie/astronomie.json`

Or this command to build one file from the integration tests. The `-s` parameter specifies an existing style sheet
file, `-o` the output folder (will be created if it doesn't exist) and the last value specifies the mediawiki file that
should be turned into an eBook.

`./wiki2book standalone -s ./integration-tests/style.css -o ./another-example-book ./integration-tests/test-real-article-Erde.mediawiki`

# Development

For building, running, testing, etc. take a look at the `src` folder and `src/README.md`.

# Long-term goals

* Be independent of the specific Wikipedia instance (#5)
* Create a public API and web app (#7)
* Ask Wikipedia if they want to embed/link to this tool in any way (that would be super cool :D )
