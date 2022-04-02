# wiki2book
**wiki2book** is a tool to create good-looking EPUB-eBooks from one or more Wikipedia articles.

The goal of this converter is to create nearly print-ready eBooks from a couple of Wikipedia articles.
This should make reading Wikipedia articles even more fun and may create a whole new readership for this awesome and imperceptibly large database of knowledge. 

**Why not simply using pandoc?**<br>
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

The current CLI is pretty simple: `wiki2book ./path/to/project.json`

This `project.json` is a configuration for a project and may look like this:

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

**Notice:** Currently only the German Wikipedia is supported.
However, you can specify `en` as `wikipedia-domain` to download articles from the English Wikipedia.
But because a lot of German template-strings are removed while parsing, the English strings remain and result in unwanted stuff in the eBook.

# Development

1. Go into `src` folder
2. Follow instructions of the README.md there

# Long-term goals

* Be independent of the specific Wikipedia instance (#5)
* Create a public API and web app (#7)
* Ask Wikipedia if they want to embed/link to this tool in any way