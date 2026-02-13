package generator

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"wiki2book/config"
	"wiki2book/util"

	"github.com/go-shiori/go-epub"
	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

type EpubGenerator struct {
	configService *config.ConfigService
}

func NewEpubGenerator(configService *config.ConfigService) *EpubGenerator {
	return &EpubGenerator{configService: configService}
}

func (g *EpubGenerator) GenerateEpub(articleFiles []string, outputFile string, metadata config.Metadata) error {
	var err error

	sigolo.Debugf("Generate EPUB to '%s' for articles %v", outputFile, articleFiles)

	if g.configService.Get().OutputType != config.OutputTypeEpub2 && g.configService.Get().OutputType != config.OutputTypeEpub3 {
		return errors.Errorf("Output type '%s' does not support EPUB generation. This is a Bug.", g.configService.Get().OutputType)
	}

	if g.configService.Get().OutputDriver == config.OutputDriverPandoc {
		err = g.GenerateEpubWithPandoc(articleFiles, outputFile, metadata)
	} else if g.configService.Get().OutputDriver == config.OutputDriverInternal {
		err = g.GenerateEpubWithGoLibrary(articleFiles, outputFile, metadata)
	} else {
		return errors.Errorf("Output type '%s' does not support EPUB generation. This is a Bug.", g.configService.Get().OutputType)
	}

	return err
}

func (g *EpubGenerator) GenerateEpubWithPandoc(sourceFiles []string, outputFile string, metadata config.Metadata) error {
	// Example: pandoc -o Stern.epub --css ../../style.css --epub-embed-font="/usr/share/fonts/TTF/DejaVuSans*.ttf" Stern.html

	args := []string{
		"-f", "html",
		"-t", g.configService.Get().OutputType,
		"-o", outputFile,
		"--metadata", "title=" + metadata.Title,
		"--metadata", "author=" + metadata.Author,
		"--metadata", "rights=" + metadata.License,
		"--metadata", "language=" + metadata.Language,
		"--metadata", "date=" + metadata.Date,
	}
	if g.configService.Get().TocDepth > 0 {
		args = append(args, "--toc", "--toc-depth", strconv.Itoa(g.configService.Get().TocDepth))
	}
	if g.configService.Get().PandocDataDir != "" {
		args = append(args, "--data-dir", g.configService.Get().PandocDataDir)
	}
	if g.configService.Get().StyleFile != "" {
		args = append(args, "--css", g.configService.Get().StyleFile)
	}
	if g.configService.Get().CoverImage != "" {
		args = append(args, "--epub-cover-image="+g.configService.Get().CoverImage)
	}
	if len(g.configService.Get().FontFiles) > 0 {
		for _, file := range g.configService.Get().FontFiles {
			args = append(args, "--epub-embed-font="+file)
		}
	}

	args = append(args, sourceFiles...)

	err := util.Execute(g.configService.Get().PandocExecutable, g.configService.Get().CacheDir, args...)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error generating EPUB file '%s' using pandoc", outputFile))
	}

	return nil
}

func (g *EpubGenerator) GenerateEpubWithGoLibrary(sourceFiles []string, outputFile string, metadata config.Metadata) error {

	epubObj, err := epub.NewEpub("My title")
	if err != nil {
		return errors.Wrap(err, "Error generating new EPUB object")
	}

	epubObj.SetAuthor(metadata.Author)
	epubObj.SetLang(metadata.Language)
	epubObj.SetTitle(metadata.Title)

	internalCoverImagePath, err := epubObj.AddImage(g.configService.Get().CoverImage, "cover.png")
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error adding cover image '%s' to EPUB object", g.configService.Get().CoverImage))
	}

	err = epubObj.SetCover(internalCoverImagePath, "")
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error setting internal image '%s' as cover image on EPUB object", internalCoverImagePath))
	}

	headingTitleRegex := regexp.MustCompile(`(?s)<h1>(.*)</h1>`)
	var fileBytes []byte
	for _, sourceFile := range sourceFiles {
		sigolo.Debugf("Add source file %s to EPUB object", sourceFile)
		fileBytes, err = os.ReadFile(sourceFile)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Error reading source file '%s' to add it to the EPUB object", sourceFile))
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

	_, err = epubObj.AddCSS(g.configService.Get().StyleFile, "style.css")
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error adding CSS file %s to EPUB object", g.configService.Get().StyleFile))
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
	for _, fontFile := range g.configService.Get().FontFiles {
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
