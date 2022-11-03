# Cover generation

The `generate-cover.sh` script generates a `cover.png` file with the given logo on top.

# CLI usage

`./generate-cover.sh [-h|--help] [-f|--font <font>] [-l|--logo <logo-file>] <title> <pre> <post>`

Optional parameter:

* Use `-h` or `--help` to show usage information
* `-f` or `--font`: Specify the LaTeX font, for example `--font lmodern`. Default is `libertinus`.
* `-l` or `--logo`: Specify the logo file on top of the title, for example `--logo ../../my-logo.jpg`. Default is `./logo`.
* `-t` or `--template`: Specify the LaTeX file to be used, for example `--template ../../my-template.tex`. Default is `./cover.tex`.

Arguments:
* `title`: The title to show.
* `pre`: Text before the title.
* `post`: Text after the title.

# Example

`./generate-cover.sh -f lmodern -l ../logo.png "Astronomy" "This is what Wikipedia writes about ..."
