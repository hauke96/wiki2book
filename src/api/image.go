package api

import (
	"fmt"
	"github.com/hauke96/sigolo/v2"
	"strings"
	"wiki2book/config"
	"wiki2book/util"
)

const imgSize = 600

// resizeAndCompressImage will convert and rescale the image so that it's suitable for eBooks.
func resizeAndCompressImage(imageFilepath string, toGrayscale bool) error {
	sigolo.Tracef("Process image '%s'", imageFilepath)

	args := []string{
		imageFilepath,
		"-resize", fmt.Sprintf("%dx%d>", imgSize, imgSize),
		"-quality", "75",
		"-define", "PNG:compression-level=9",
		"-define", "PNG:compression-filter=0",
	}

	if toGrayscale {
		sigolo.Tracef("Add args to convert '%s' to grayscale", imageFilepath)
		args = append(args, "-colorspace", "gray")
	}

	args = append(args, imageFilepath)

	err := util.Execute(config.Current.ImageMagickExecutable, args...)

	if err != nil {
		sigolo.Errorf("Converting image %s failed", imageFilepath)
	}

	return err
}

// convertPdfToPng will convert the given PDF file into a PNG image at the given location. This conversion does neither
// rescale nor process the image in any other way, use resizeAndCompressImage accordingly.
func convertPdfToPng(inputPdfFilepath string, outputPngFilepath string) error {
	sigolo.Tracef("Convert PDF '%s' to PNG '%s'", inputPdfFilepath, outputPngFilepath)

	args := []string{
		"-density", "300",
		inputPdfFilepath,
		outputPngFilepath,
	}

	err := util.Execute(config.Current.ImageMagickExecutable, args...)
	if err != nil {
		sigolo.Errorf("Converting PNG %s into an PNG image failed", inputPdfFilepath)
	}

	return err
}

func convertSvgToPng(svgFile string, pngFile string, svgToPngCommandTemplate string) error {
	sigolo.Tracef("Convert SVG %s to PNG %s", svgFile, pngFile)

	svgToPngCommandString := strings.ReplaceAll(svgToPngCommandTemplate, config.InputPlaceholder, svgFile)
	svgToPngCommandString = strings.ReplaceAll(svgToPngCommandString, config.OutputPlaceholder, pngFile)

	splitCommand := strings.Split(svgToPngCommandString, " ")
	svgToPngCommand := splitCommand[0]
	svgToPngCommandArgs := splitCommand[1:]

	err := util.Execute(svgToPngCommand, svgToPngCommandArgs...)

	if err != nil {
		sigolo.Errorf("Converting image %s to PNG failed", svgFile)
	}

	return err
}
