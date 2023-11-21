package epub

import (
	"fmt"
	"github.com/pkg/errors"
	"wiki2book/project"
	"wiki2book/util"
)

func Generate(sourceFiles []string, outputFile string, styleFile string, coverFile string, pandocDataDir string, fontFiles []string, metadata project.Metadata) error {
	// Example: pandoc -o Stern.epub --css ../../style.css --epub-embed-font="/usr/share/fonts/TTF/DejaVuSans*.ttf" Stern.html

	args := []string{
		"-f", "html",
		"-t", "epub3",
		"-o", outputFile,
		"--toc",
		"--metadata", "title=" + metadata.Title,
		"--metadata", "author=" + metadata.Author,
		"--metadata", "rights=" + metadata.License,
		"--metadata", "language=" + metadata.Language,
		"--metadata", "date=" + metadata.Date,
	}
	if pandocDataDir != "" {
		args = append(args, "--data-dir", pandocDataDir)
	}
	if styleFile != "" {
		args = append(args, "--css", styleFile)
	}
	if coverFile != "" {
		args = append(args, "--epub-cover-image="+coverFile)
	}
	if len(fontFiles) > 0 {
		for _, file := range fontFiles {
			args = append(args, "--epub-embed-font="+file)
		}
	}

	args = append(args, sourceFiles...)

	err := util.Execute("pandoc", args...)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error generating EPUB file %s using pandoc", outputFile))
	}

	return nil
}
