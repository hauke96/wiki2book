package api

import (
	"github.com/hauke96/sigolo/v2"
	"strings"
	"wiki2book/config"
	"wiki2book/util"
)

const imgSize = "600"

// resizeAndCompressImage will convert and rescale the image so that it's suitable for eBooks.
func resizeAndCompressImage(imageFilepath string, commandTemplate string) error {
	sigolo.Tracef("Process image '%s'", imageFilepath)

	commandString := strings.ReplaceAll(commandTemplate, config.InputPlaceholder, imageFilepath)
	commandString = strings.ReplaceAll(commandString, config.OutputPlaceholder, imageFilepath)
	commandString = strings.ReplaceAll(commandString, config.SizePlaceholder, imgSize)

	commandParts := strings.Split(commandString, " ")
	commandExecutable := commandParts[0]
	commandArgs := commandParts[1:]

	err := util.Execute(commandExecutable, commandArgs...)

	if err != nil {
		sigolo.Errorf("Converting image %s failed", imageFilepath)
	}

	return err
}

// convertPdfToPng will convert the given PDF file into a PNG image at the given location. This conversion does neither
// rescale nor process the image in any other way, use resizeAndCompressImage accordingly.
func convertPdfToPng(inputPdfFilepath string, outputPngFilepath string, commandTemplate string) error {
	sigolo.Tracef("Convert PDF '%s' to PNG '%s'", inputPdfFilepath, outputPngFilepath)

	commandString := strings.ReplaceAll(commandTemplate, config.InputPlaceholder, inputPdfFilepath)
	commandString = strings.ReplaceAll(commandString, config.OutputPlaceholder, outputPngFilepath)

	commandParts := strings.Split(commandString, " ")
	commandExecutable := commandParts[0]
	commandArgs := commandParts[1:]

	err := util.Execute(commandExecutable, commandArgs...)

	if err != nil {
		sigolo.Errorf("Converting PNG %s into an PNG image failed", inputPdfFilepath)
	}

	return err
}

func convertSvgToPng(svgFile string, pngFile string, commandTemplate string) error {
	sigolo.Tracef("Convert SVG %s to PNG %s", svgFile, pngFile)

	commandString := strings.ReplaceAll(commandTemplate, config.InputPlaceholder, svgFile)
	commandString = strings.ReplaceAll(commandString, config.OutputPlaceholder, pngFile)

	commandParts := strings.Split(commandString, " ")
	commandExecutable := commandParts[0]
	commandArgs := commandParts[1:]

	err := util.Execute(commandExecutable, commandArgs...)

	if err != nil {
		sigolo.Errorf("Converting image %s to PNG failed", svgFile)
	}

	return err
}
