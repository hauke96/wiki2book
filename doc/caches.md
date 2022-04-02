This is a description of the caching this tool uses.

# Goal

The goal is the following: Running this tool twice on the same inputs should only download the content once.

To achieve this, several caches for articles, rendered templates, images and rendered math code are used.

# How it works

Quick facts:

* A cache is just a folder containing files.
* Each file (except article files) has a SHA1 hash as name and a file ending (e.g. `.svg`).
* Each file can contain direct content or another hash pointing to another file in the cache (like a symlink).

## Filling the cache

The cache is filles from the `api` package: Every request gets stored right into the cache.
When the cache already contains an item (e.g. an image), no request is made.

## Clearing the cache

The cache is not automatically cleared but feel free to remove a folder or just one file.
All missing content will be downloaded and saved again.

# Caches

The are caches for the following things:

* [Articles](#articles)
* [Images](#Images)
* [Rendered math](#math)
* [Templates](#Templates)

The cache folders are right next to the project file.

## Articles

* Folder: `articles`.
* Filenames: Just the article name but spaces are replaced by underscores (`_`).

This contains the json response of the Wikipedia API, for example: `{"parse":{"title":"Interstellares Eis","pageid":10700436,"wikitext":{"*":"'''Interstellares Eis''', oder auch ........."}}}`.

## Images

* Folder: `images`.
* Filenames: Just the image name but spaces are replaced by underscores (`_`).

This contains the image that should be used in the eBook.
When the image got downloaded by wiki2book, it's scaled and turned into a grayscale image.
This is done to save space and normal eBook readers can't represent colors anyway.

There might be a lot of files with hash values as names and `.svg` as well as `.png` extensions, they contain rendered math.

## Math

* Folder: `math`.
* Filenames: SHA1 hash of the url-encoded math string.

Each file containing a hash value and files with that exact hash value as filename exist in the `images` cache.
The files from the `math` cache are therefore pointing to files in the `images` cache.

**Example:**<br>
Having `<math>\sqrt{x} + x</math>` the resulting url-encoded string would be `%5Csqrt%7Bx%7D+%2B+x` (see [golang doc](https://pkg.go.dev/net/url#QueryEscape) for details).
The `math` cache folder will then contain a file with he SHA1 hash value of that encoded string (in this case `44fdead768517de73cbc9ce9c9e4300c060b6a84`) as filename.
This file then contains the resource token (another SHA1 hash) received by the Wikipedia API (see [rendering math documentation](./rendering-math.md) for details).
The `images` cache folder will contain two image files (an `.svg` and `.png` file) with exact that resource token as filename so that the `math` cache files always point to these image files.

The file structure for the above example would look like this:
```
|– your-project/
   |– images/
      |– 5bbe82a3c29d695afc67eb99a18ed8453e28f12f.png
      |– 5bbe82a3c29d695afc67eb99a18ed8453e28f12f.svg
   |– math/
      |– 44fdead768517de73cbc9ce9c9e4300c060b6a84
```
The file `math/44fdead768517de73cbc9ce9c9e4300c060b6a84` has as only content the string `5bbe82a3c29d695afc67eb99a18ed8453e28f12f` which was received by the Wikipedia API.

## Templates

* Folder: `templates`.
* Filenames: SHA1 hash of the template string.

Each template file contains the rendered content for a given template.
This probably is a mixture of HTML and Wikitext.

**Example:**<br>
The template string `{{foobar}}` results -- after removing the braces `{{` and `}}` -- in the SHA1 hash `8843d7f92416211de9ebb963ff4ce28125932878`.
The file `templates/8843d7f92416211de9ebb963ff4ce28125932878` contains the rendered template as received by the Wikipedia API.