This is a description of the code structure and architecture.

# File and folder Structure

| Folder      | Contains ...                                                                                                             |
|-------------|--------------------------------------------------------------------------------------------------------------------------|
| `cache`     | Cache storage and eviction logic. Is and should be used by other code reading/writing files.                             |
| `config`    | Code for the config-file- and project-based wiki2book configuration.                                                     |
| `generator` | Generators to produce HTML and EPUB files.                                                                               |
| `http`      | HTTP logic and abstraction layer.                                                                                        |
| `image`     | Image processing and conversion logic.                                                                                   |
| `parser`    | Heart of wiki2book: Parser and tokenizer to replace wikitext features (e.g. a table) into tokens for the HTML generator. |
| `test`      | Helper and dummy files for the unit tests.                                                                               |
| `util`      | All sort of helper functions and auxiliary abstraction layers.                                                           |
| `wikipedia` | Abstraction layer for the Wikipedia API.                                                                                 |

The `main.go` contains the CLI setup and orchestrates the CLI commands.

# Architecture

To generate an eBook based on a project file, the following high-level steps are executed:

1. Read the given project file and merge it with the configurations from the default-config, config-file and CLI.
2. This might be executed in parallel, depending on the config: For each Wikipedia article in the project, do the following:
   1. Download the wikitext of the article.
   2. The wikitext is tokenized, resulting in the tokenized text and a token map.
      During this step, templates are evaluated and math is rendered to an SVG.
   3. After tokenization all images that have been found are downloaded and cached to disk.
   4. Finally, an HTML file for the article is generated.
3. All HTML files and the metadata provided in the project file are used to generate an EPUB file.

Only the tokenizer and the HTML generator are actually interesting, all other party are rather small and simple.

## Tokenizer

### Idea

The general idea of the tokenization is to abstract from the wikitext.
Turning all interesting (= wikitext specific) features of the content into abstract tokens enables a generator to produce output (e.g. HTML) specific content.
Each token is a struct and contains the parsed wikitext and often references specific child tokens (e.g. the table token references its rows and caption).

Tokens are structures hierarchically, which means a token potentially has a parent (not stored explicitly) and children.
There are some high-level token types that do not require to be inside other tokens and, therefore, do not necessarily have a parent token.

All token (high-level and normal ones) are stored in a map, which required each token to have a unique key.

If a high-level token has no parent token, then the wikitext content of that structure (e.g. a table) is replaced by a unique string as shown below.
The HTML generator then extracts these unique strings and turns the underlying token with all its children into HTML. 

**Example:**
A table-row is *not* a high-level token because it *must* be inside a table.
The surrounding table, however, *is* a high-level token, because it does *not* need to be inside any other wikitext structure.
But the table can still be a child of some other token, for example of another table cell for nested tables.

### Token format

A high-level token (token without parent) is referenced in the text by a unique string (token-key) that looks like this: `Some text $$TOKEN_<type>_<counter>$$ some further text`.
The `<type>` (e.g. `IMAGE`) is one of the many `TOKEN_...` constants in `tokenizer.go` and describes what type of thing this token represents (e.g. an image).
The `<counter>` value is used to create unique values and is increased for each token string regardless of its type.

Because just to know "okay, here's an image" is not that helpful, each token is stored in a map structure (token map) mapping from token-key to content of the token.
The content of a token is a struct that may contain child-tokens.

**Example:**<br>
A simple image like `[[Datei:img.jpg|100px|My caption]]` would be replaced with `$$TOKEN_IMAGE_0$$`.
However, this token consists of the following three parts (as go-struct but here as JSON representation):
```json
{
   "Token": null,
   "Filename": "images/Img.jpg",
   "Caption": {
      "Token": null,
      "Content": ""
   },
   "SizeX": 100,
   "SizeY": -1
}
```

The `"Token"`-field results from the fact that all go struct token inherit from the interface `Token`.

### Token vs. marker

There are also things called *marker*, which represent parts of the text without having any content themselves.
A classic example would be the start of a paragraph or a newline.
However, newlines are pretty much ignored.

Start and end of bold and italic parts are actually turned into markers.
This is because they can overlap so that normal tokens don't work here:
```
normal ''' bold '' bold-italic ''' italic '' normal
          |-------bold--------|
                  |--------italic--------|
```

### Steps of tokenization

The `Tikenizer.tokenizeContent()` function gives a good overview of what's happening but the actual process starts in the `Tokenizer.Tokenize()` function.
The received string can be any wikitext (even with existing tokens in it) and the resulting string is then a tokenized version of the input.
The token map is stored in the struct `Tokenizer` which is used the whole time.

Compiler do often parse the content char by char.
Wiki2book uses a different approach because wikitext is quite complex.
Each structure of wikitext is parsed after another, meaning when all references have been parsed, then the internal links are parsed and so on.

The process is recursive, because the content of a structure, e.g. the caption of an image, is given to the tokenization function to obtain a tokenized version of that caption.
This makes parsing deeply nested structures quite easy.

The steps during tokenization are the following:

1. Cleanup: Remove unwanted stuff like categories, specific templates, empty sections, ...
2. Evaluate templates. Each evaluated template consists of HTML, wikitext or a mixture of both but doesn't contain new templates.
3. A new cleanup call ensures that the templates haven't added new unwanted stuff to the overall content.
4. Actual tokenization starts by calling numerous parsing-functions for each aspect of wikitext.
The order of each parsing function is important because e.g. embedded images and external links are quite similar and parsing images first makes things a bit easier.

## Generator

There are currently two generators: One for HTML and one for EPUB.
The EPUB generator just uses `pandoc` and, therefore, is not that interesting.

So I just focus on the HTML generator here.

### Idea

The workflow is pretty straight forward:
Find each high-level token in the tokenized wikitext and expand it to the according template.
The term "expanding" here means, that the content of the token (e.g. the filepath and caption of an image) is used to fill an HTML-template with it.
The content text of a token can itself contain one or more high-level token, so this whole expanding-strategy is recursive.

### Steps of generating HTML

The general steps in the `expand()` function are the following:

1. When expanding text (with potential high-level tokens in it) → `expandString()`
   1. Find all high-level token-keys in the given text
   2. For each found token object (obtained via the token map):
      Call the `expand()` function again, which will lead to step 2 below.
         1. Go through all known token types to find correct function to expand the token (= generate HTML for the token).
            This function itself may find some token and will therefore jump to step 1 to expand the found token.
            **Example:** The `expandListItem` function takes the content of its token and directly calls `expand` to expand all possible tokens within the list item.
            After that the template for a list item is filled and returned.
         2. Replace the original occurrence of the token by the generated HTML.
2. When expanding a token-struct → `expandToken()`
   1. Calls the matching expansion function for the given type (e.g. for `ExternalLinkToken` it's the `expandExternalLink()` function)
   2. Each such function themselves calls the `expand()` function, which then leads to step 1 or 2 above.
   3. The result of the expanded child-elements of the token (if there are any) is used in a simple HTML template to create the result value.

The `expand` function is initially called from the public `Generate` function of the HTML generator.
This wrapper function adds HTML-header and -footer and also writes the result to disk.

### Example

I think the best way to describe the process is to produce HTML for the image-token from the tokenization example above.
So we consider the already tokenized wikitext `[[Datei:img.jpg|100px|My caption]]`.

The code for this is the `expandImage` function in `html.go`.
It receives the token (so object for which we had the JSON-representation in the above example) and returns the HTML for this image.

1. Expand the child elements, which is only the caption in this case.
The caption is just a string and may contain high-level token (which is not the case in this example).
It therefore is passed to the `expand()` function to get these high-level tokens turned into HTML.
2. The size of the image if turned into HTML.
This is a bit complex due to the different ways of specifying the size of an image.
3. Finally, the HTML-template for an image is used and filled with the expanded string from step 1 and 2.

For our example from above, the HTML would look like this:

```html
<div class="figure">
<img alt="image" src="./img.jpg" style="vertical-align: middle; width: 100px; height: 100px;">
<div class="caption">
My Caption
</div>
</div>
```

