This is a description of the code structure and architecture.

# File and folder Structure

| Folder      | Contains ...                                                                                            |
|-------------|---------------------------------------------------------------------------------------------------------|
| `api`       | code to make HTTP requests to e.g. download images.                                                     |
| `generator` | generators to produce HTML and EPUB files.                                                              |
| `parser`    | basically the tokenizer to replace wikitext features (e.g. a table) into tokens for the HTML generator. |
| `project`   | a simple struct to read the project file.                                                               |
| `test`      | helper and dummy files for the unit tests.                                                              |
| `util`      | all sort of helper functions.                                                                           |

# Architecture

To generate an EPUB eBook the following high-level steps are executed:

1. Read project file
2. For each Wikipedia article in the project, so the following:
   1. Download the wikitext of the article.
   2. The wikitext is tokenized, resulting in the tokenized text and a token map. 
   3. Generate an HTML file using tokenized text and the token map.
3. Generate an EPUB file from all HTML files and metadata provided in the project file.

Only the tokenizer and the HTML generator is actually interesting, all other party are rather small and simple.

## Tokenizer

### Idea

The general idea of the tokenization is to abstract from the wikitext.
Turning all interesting (= wikitext specific) features of the content into abstract tokens enables a generator to produce output (e.g. HTML) specific content.

### Token format

A token in the text then looks like this: `Some text $$TOKEN_<type>_<counter>$$ some further text.`
The `<type>` (e.g. `IMAGE`) is one of the many `TOKEN_...` constants in `tokenizer.go` and describes what type of thing this token represents (e.g. an image).

Because just to know "okay, here's an image" is not that helpful, each token is stored in a map structure (token map) mapping from token to content of the token.
The content of a token may also contain tokens.

**Example:**<br>
A simple image like `[[Datei:img.jpg|100px|My caption]]` would be replaced with `$$TOKEN_IMAGE_3$$`.
However this token consists of the following three parts: `$$TOKEN_IMAGE_FILENAME_0$$ $$TOKEN_IMAGE_CAPTION_2$$ $$TOKEN_IMAGE_SIZE_1$$`
Because the caption is the last part of the original image wiki-tag, its counter is higher than the one of the image-size-token.

### Token vs. marker

There are also things called *markers* representing parts of the text without any content.
A classic example would be the start of a paragraph or a newline.
However, newlines are pretty much ignored at all and the start of a new paragraph (for which a parse function exists) is currently also ignored.

Start and end of bold and italic parts are actually turned into markers.
This is because they can overlap so that normal tokens don't work here: `normal ''' bold '' bold-italic ''' italic '' normal`

### Steps of tokenization

The `Tikenizer.tokenizeContent` function gives a good overview of what's happening.
It receives returns a string.
The received string can be any wikitext (even with existing tokens in it) and the resulting string is then a tokenized version of the input.
The token map is stored in the struct `Tokenizer` which is used the whole time.

The steps when tokenizing a string:

1. Cleanup: Remove unwanted stuff like categories, specific templates and existing HTML
2. Tokenize headings
3. Handle references
4. Do the following until no further tokenization is possible:
   1. Evaluate templates (resulting in normal text, HTML or new wikitext)
   2. Cleanup (as above)
   3. Escape images (the filename may contain spaces)
   4. Tokenize in a specific order:
      1. Internal links
      2. Images (which are syntactically similar to internal links and therefor have to be parsed after internal links)
      3. External links (again, similar to internal links)
      4. Math
      5. Tables
      6. Lists (after tables because a table can contain lists and parsing lists afterwards makes things easier)

## Generator

There are currently two generators: One for HTML and one for EPUB.
The EPUB generator just uses `pandoc` and is therefore not that interesting.

So I just focus on the HTML generator here.

### Idea

The workflow is pretty straight forward: Go through all token of the tokenized wikitext and *expand* each to the according template.
Expanding here means that the content of the token (= the entry in the token map) is used and an HTML-template is filled with it.
The content of a token can itself contain one or more token, so this whole expanding-strategy contains recursion.

### Steps of generating HTML

The general steps in the `expand` function are the following:

1. Find - via a regex - all token in the given text
2. For each found token:
   1. Go through all known token types to find correct function to expand the token (= generate HTML for the token). This function itself may find some token and will therefore jump to step 1 to expand the found token. **Example:** The `expandListItem` function takes the content of its token and directly calls `expand` to expand all possible tokens within the list item. After that the template for a list item is filled and returned.
   2. Replace the original occurrence of the token by the generated HTML.

The `expand` function is initially called from the public `Generate` function of the HTML generator.
This wrapper function adds HTML-header and -footer and also writes the result to disk.

### Example
I think the best way to describe the process is to produce HTML for the image-token from the tokenization example above.
Remember: We tokenized the wikitext `[[Datei:img.jpg|100px|My caption]]`.

The code for this is the `expandImage` function in `html.go`.
It receives the token (so `$$TOKEN_IMAGE_3$$` in our example) and the token map and returns the HTML for this image.

1. Expand the image token: Find all sub-tokens (for the filename, size and caption) in the token map entry for out image token.
2. For each sub-token: Determine its type (e.g. the sub-token for the caption) and expand the token to get the actual content (e.g. the caption text).
   1. For the caption only: Call `expand` (s. above) to expand all tokens (for example a link) within the caption.
3. Determine which HTML-template to use (inline, with/without size).
4. Fill in the template and return the HTML.

For our example from above, the HTML would look like this:
```html
<div class="figure">
<img alt="image" src="img.jpg" style="vertical-align: middle; width: 100px; height: 100px;">
<div class="My caption">
```

