# wiki2book

Create an eBook (EPUB) from wikipedia pages.

## Preliminaries

You need the following tools and fonts:

1. ImageMagick (to have the `convert` command)
2. Pandoc (to have the `pandoc` command)
3. DejaVu fonts in `/usr/share/fonts/TTF/DejaVuSans*.ttf`

# Run wiki2book

1. Go into `src` folder
2. Follow instructions of the README.md there

# Create own eBook-project

// TODO

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
