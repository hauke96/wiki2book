package image

import (
	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
	"strings"
	"wiki2book/config"
	"wiki2book/util"
)

type ImageProcessingService interface {
	ResizeAndCompressImage(imageFilepath string, commandTemplate string) error
	ConvertPdfToPng(inputPdfFilepath string, outputPngFilepath string, commandTemplate string) error
	ConvertSvgToPng(svgFile string, pngFile string, commandTemplate string) error
}

type ImageProcessingServiceImpl struct{}

func NewImageProcessingService() ImageProcessingService {
	return &ImageProcessingServiceImpl{}
}

// resizeAndCompressImage will convert and rescale the image so that it's suitable for eBooks.
func (s *ImageProcessingServiceImpl) ResizeAndCompressImage(imageFilepath string, commandTemplate string) error {
	sigolo.Tracef("Process image '%s'", imageFilepath)

	commandString := strings.ReplaceAll(commandTemplate, config.InputPlaceholder, imageFilepath)
	commandString = strings.ReplaceAll(commandString, config.OutputPlaceholder, imageFilepath)

	err := util.ExecuteCommandWithArgs(commandString)
	return errors.Wrapf(err, "Converting image %s failed", imageFilepath)
}

// convertPdfToPng will convert the given PDF file into a PNG image at the given location. This conversion does neither
// rescale nor process the image in any other way, use resizeAndCompressImage accordingly.
func (s *ImageProcessingServiceImpl) ConvertPdfToPng(inputPdfFilepath string, outputPngFilepath string, commandTemplate string) error {
	sigolo.Tracef("Convert PDF '%s' to PNG '%s'", inputPdfFilepath, outputPngFilepath)

	commandString := strings.ReplaceAll(commandTemplate, config.InputPlaceholder, inputPdfFilepath)
	commandString = strings.ReplaceAll(commandString, config.OutputPlaceholder, outputPngFilepath)

	err := util.ExecuteCommandWithArgs(commandString)
	return errors.Wrapf(err, "Converting PNG %s into an PNG image failed", inputPdfFilepath)
}

func (s *ImageProcessingServiceImpl) ConvertSvgToPng(svgFile string, pngFile string, commandTemplate string) error {
	sigolo.Tracef("Convert SVG %s to PNG %s", svgFile, pngFile)

	commandString := strings.ReplaceAll(commandTemplate, config.InputPlaceholder, svgFile)
	commandString = strings.ReplaceAll(commandString, config.OutputPlaceholder, pngFile)

	err := util.ExecuteCommandWithArgs(commandString)
	return errors.Wrapf(err, "Converting image %s to PNG failed", svgFile)
}
