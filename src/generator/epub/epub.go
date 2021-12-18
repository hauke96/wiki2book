package epub

import (
	"github.com/hauke96/wiki2book/src/util"
	"github.com/pkg/errors"
)

func Generate(sourceFile string, outputFile string, styleFile string, coverFile string, title string) error {
	// Example: pandoc -o Stern.epub --css ../../style.css --epub-embed-font="/usr/share/fonts/TTF/DejaVuSans*.ttf" Stern.html

	err := util.Execute("pandoc",
		"-f", "html",
		"-t", "epub2",
		"-o", outputFile,
		"--css", styleFile,
		"--toc",
		"--epub-embed-font=/usr/share/fonts/TTF/DejaVuSans.ttf",
		"--epub-embed-font=/usr/share/fonts/TTF/DejaVuSansMono*.ttf",
		"--epub-embed-font=/usr/share/fonts/TTF/DejaVuSans-B*.ttf",
		"--epub-embed-font=/usr/share/fonts/TTF/DejaVuSans-O*.ttf",
		"--epub-cover-image="+coverFile,
		"--epub-metadata=metadata.xml",
		"--metadata", "title=Wikipedia: "+title,
		"--metadata", "author=Wikipedia contributors",
		sourceFile)
	if err != nil {
		return errors.Wrap(err, "Error generating EPUB file using pandoc")
	}

	return nil
}
