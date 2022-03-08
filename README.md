# wiki2book
**wiki2book** is a tool to create good-looking EPUB-eBooks from one or more wikipedia articles.

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

# TODOs

Open tasks of this project:

* [x] Add cover to EPUB file
* [x] Tables
* [x] Caption of tables
* [x] Ordered lists
* [ ] Problematic pages:
  * *(currently no problematic pages are known)*
* [x] ~~Use superscript `<sup>...</sup>` for citations~~ (brackets are used)
* [x] Math rendering
* [x] Save rendered templates like images
* [x] Create a file format (JSON?) to create a book in onw run (multiple articles, style, fonts, cover, ...)
* [ ] Add tests
* [ ] Pretty focused on German articles → support at least English Wikipedia
* [ ] Extend CLI
  * [ ] Parameter for project (`--project ./path/to/project.json`) 
  * [ ] Single wiki articles (`--article foobar`) 
  * [ ] Wikitext file (`--file ./file.wikitext`)
  * [ ] Style file (`--style ./style.css`)
  * [ ] Font (`--font /usr/share/fonts/TTF/SuperPrettyFont.ttf`)