# wiki2book

Create an eBook (EPUB) from wikipedia pages.

## Preliminaries

You need the following tools and fonts:

1. ImageMagick (to have the `convert` command)
2. Pandoc (to have the `pandoc` command)
3. DejaVu fonts in `/usr/share/fonts/TTF/DejaVuSans*.ttf`

# Create own eBook

// TODO

# TODOs

Open tasks of this project:

* [x] Add cover to EPUB file
* [x] Tables
* [ ] Ordered lists
* [ ] Problematic pages:
  * Universum (table at the beginning)
  * Interstellarer_Staub (list with one element at the end)
  * Riesenstern (weird top of page)
  * Sternentstehung#Molekülwolkenstruktur (table with caption)
  * Exoplanet#Erste Entdeckungen von Exoplaneten (ref that's not moved to bottom)
  * Exoplanet#Bekannte Projekte und Instrumente zum Nachweis von Exoplaneten (not rendered table)
  * Exoplanet#Zahl der bekannten Exoplaneten (not rendered table)
  * Exoplanet#Masse und Radius der entdeckten Planeten (not evaluated template)
  * Sonne (not rendered table at beginning)
* [ ] Use superscript `<sup>...</sup>` for citations
* [ ] Math rendering
* [ ] Save rendered templates like images
* [x] Create a file format (JSON?) to create a book in onw run (multiple articles, style, fonts, cover, ...)
* [ ] Add tests
