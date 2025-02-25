package epub

import (
	"fmt"
	"github.com/go-shiori/go-epub"
	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
	"os"
	"regexp"
	"strconv"
	"wiki2book/config"
	"wiki2book/project"
	"wiki2book/util"
)

func Generate(sourceFiles []string, outputFile string, outputType string, styleFile string, coverFile string, pandocDataDir string, fontFiles []string, tocDepth int, metadata project.Metadata) error {
	// Example: pandoc -o Stern.epub --css ../../style.css --epub-embed-font="/usr/share/fonts/TTF/DejaVuSans*.ttf" Stern.html

	args := []string{
		"-f", "html",
		"-t", outputType,
		"-o", outputFile,
		"--metadata", "title=" + metadata.Title,
		"--metadata", "author=" + metadata.Author,
		"--metadata", "rights=" + metadata.License,
		"--metadata", "language=" + metadata.Language,
		"--metadata", "date=" + metadata.Date,
	}
	if tocDepth > 0 {
		args = append(args, "--toc", "--toc-depth", strconv.Itoa(tocDepth))
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

	err := util.Execute(config.Current.PandocExecutable, args...)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error generating EPUB file %s using pandoc", outputFile))
	}

	return nil
}

func GenerateWithGoLibrary(sourceFiles []string, outputFile string, coverFile string, styleFile string, fontFiles []string, metadata project.Metadata) error {

	epubObj, err := epub.NewEpub("My title")
	if err != nil {
		return errors.Wrap(err, "Error generating new EPUB object")
	}

	epubObj.SetAuthor(metadata.Author)
	epubObj.SetLang(metadata.Language)
	epubObj.SetTitle(metadata.Title)

	internalCoverImagePath, err := epubObj.AddImage(coverFile, "cover.png")
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error adding cover image %s to EPUB object", coverFile))
	}

	err = epubObj.SetCover(internalCoverImagePath, "")
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error setting internal image %s as cover image on EPUB object", internalCoverImagePath))
	}

	headingTitleRegex := regexp.MustCompile(`(?s)<h1>(.*)</h1>`)
	var fileBytes []byte
	for _, sourceFile := range sourceFiles {
		sigolo.Debugf("Add source file %s to EPUB object", sourceFile)
		fileBytes, err = os.ReadFile(sourceFile)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Error reading source file %s to add it to the EPUB object", sourceFile))
		}

		fileContent := string(fileBytes)

		sectionTitle := headingTitleRegex.FindStringSubmatch(fileContent)[1]

		// TODO Find a way to add subsections to TOC as well
		_, err = epubObj.AddSection(fileContent, sectionTitle, "", "")
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Error reading source file %s to add it to the EPUB object", sourceFile))
		}
		epubObj.EmbedImages()
	}

	_, err = epubObj.AddCSS(styleFile, "style.css")
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error adding CSS file %s to EPUB object", styleFile))
	}

	//for _, image := range images {
	//	image = strings.Split(image, ":")[1] // Remove whatever prefix (e.g. "File:" or "Datei:") this image has
	//	imageFilepath := filepath.Join(imageCache, image)
	//	sigolo.Debugf("Add image file %s to EPUB object", imageFilepath)
	//	_, err = epubObj.AddImage(imageFilepath, imageFilepath)
	//	if err != nil {
	//		return errors.Wrap(err, fmt.Sprintf("Error adding image file %s to EPUB object", image))
	//	}
	//}

	// TODO Test if this if working. The name in the CSS style and the name inside the EPUB probably need to match.
	for _, fontFile := range fontFiles {
		sigolo.Debugf("Add font file %s to EPUB object", fontFile)
		_, err = epubObj.AddFont(fontFile, fontFile)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Error adding font file %s to EPUB object", fontFile))
		}
	}

	err = epubObj.Write(outputFile)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error generating EPUB file %s", outputFile))
	}

	return nil
}
