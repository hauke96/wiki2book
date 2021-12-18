package epub

import (
	"github.com/hauke96/wiki2book/src/util"
	"github.com/pkg/errors"
)

func Generate(sourceFile string, outputFile string, styleFile string) error {
	// Example: pandoc -o Stern.epub --css ../../style.css --epub-embed-font="/usr/share/fonts/TTF/DejaVuSans*.ttf" Stern.html

	err := util.Execute("pandoc", "-o", outputFile, "--css", styleFile, "-t", "epub2", "--epub-embed-font=/usr/share/fonts/TTF/DejaVuSans*.ttf", sourceFile)
	if err != nil {
		return errors.Wrap(err, "Error generating EPUB file using pandoc")
	}

	return nil
}
