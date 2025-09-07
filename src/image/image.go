package image

import (
	"strings"
	"wiki2book/config"
	"wiki2book/util"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

type ImageProcessingService interface {
	ResizeAndCompressImage(imageFilepath string, commandTemplate string) error
	ConvertToPng(webpFile string, pngFile string, commandTemplate string) error
}

type ImageProcessingServiceImpl struct{}

func NewImageProcessingService() ImageProcessingService {
	return &ImageProcessingServiceImpl{}
}

// ResizeAndCompressImage will convert and rescale the image so that it's suitable for eBooks.
func (s *ImageProcessingServiceImpl) ResizeAndCompressImage(imageFilepath string, commandTemplate string) error {
	sigolo.Tracef("Process image '%s'", imageFilepath)

	commandString := strings.ReplaceAll(commandTemplate, config.InputPlaceholder, imageFilepath)
	commandString = strings.ReplaceAll(commandString, config.OutputPlaceholder, imageFilepath)

	err := util.ExecuteCommandWithArgs(commandString, config.Current.CacheDir)
	return errors.Wrapf(err, "Converting image %s failed", imageFilepath)
}

func (s *ImageProcessingServiceImpl) ConvertToPng(inputFile string, pngFile string, commandTemplate string) error {
	sigolo.Tracef("Convert '%s' to PNG '%s'", inputFile, pngFile)

	commandString := strings.ReplaceAll(commandTemplate, config.InputPlaceholder, inputFile)
	commandString = strings.ReplaceAll(commandString, config.OutputPlaceholder, pngFile)

	err := util.ExecuteCommandWithArgs(commandString, config.Current.CacheDir)
	return errors.Wrapf(err, "Converting image '%s' to PNG failed", inputFile)
}
