package epub

import (
	"fmt"
	"github.com/hauke96/wiki2book/src/project"
	"github.com/hauke96/wiki2book/src/util"
	"github.com/pkg/errors"
)

func Generate(sourceFiles []string, outputFile string, styleFile string, coverFile string, pandocDataDir string, metadata project.Metadata) error {
	// Example: pandoc -o Stern.epub --css ../../style.css --epub-embed-font="/usr/share/fonts/TTF/DejaVuSans*.ttf" Stern.html

	args := []string{
		"-f", "html",
		"-t", "epub2",
		"-o", outputFile,
		"--toc",
		"--epub-embed-font=/usr/share/fonts/TTF/DejaVuSans.ttf",
		"--epub-embed-font=/usr/share/fonts/TTF/DejaVuSansMono*.ttf",
		"--epub-embed-font=/usr/share/fonts/TTF/DejaVuSans-B*.ttf",
		"--epub-embed-font=/usr/share/fonts/TTF/DejaVuSans-O*.ttf",
		"--metadata", "title=" + metadata.Title,
		"--metadata", "author=" + metadata.Author,
		"--metadata", "rights=" + metadata.License,
		"--metadata", "language=" + metadata.Language,
		"--metadata", "date=" + metadata.Date,
		"--data-dir", pandocDataDir,
	}
	if styleFile != "" {
		args = append(args, "--css", styleFile)
	}
	if coverFile != "" {
		args = append(args, "--epub-cover-image="+coverFile)
	}

	args = append(args, sourceFiles...)

	err := util.Execute("pandoc", args...)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error generating EPUB file %s using pandoc", outputFile))
	}

	return nil
}
