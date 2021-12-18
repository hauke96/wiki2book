package epub

import (
	"github.com/hauke96/wiki2book/src/project"
	"github.com/hauke96/wiki2book/src/util"
	"github.com/pkg/errors"
)

func Generate(sourceFiles []string, outputFile string, styleFile string, coverFile string, metadata project.Metadata) error {
	// Example: pandoc -o Stern.epub --css ../../style.css --epub-embed-font="/usr/share/fonts/TTF/DejaVuSans*.ttf" Stern.html

	args := []string{
		"-f", "html",
		"-t", "epub2",
		"-o", outputFile,
		"--css", styleFile,
		"--toc",
		"--epub-embed-font=/usr/share/fonts/TTF/DejaVuSans.ttf",
		"--epub-embed-font=/usr/share/fonts/TTF/DejaVuSansMono*.ttf",
		"--epub-embed-font=/usr/share/fonts/TTF/DejaVuSans-B*.ttf",
		"--epub-embed-font=/usr/share/fonts/TTF/DejaVuSans-O*.ttf",
		"--epub-cover-image=" + coverFile,
		"--metadata", "title=" + metadata.Title,
		"--metadata", "author=" + metadata.Author,
		"--metadata", "rights=" + metadata.License,
		"--metadata", "language=" + metadata.Language,
	}
	args = append(args, sourceFiles...)

	err := util.Execute("pandoc",
		args...)
	if err != nil {
		return errors.Wrap(err, "Error generating EPUB file using pandoc")
	}

	return nil
}
