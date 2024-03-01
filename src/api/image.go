package api

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"wiki2book/util"
)

const imgSize = 600

// processImage will convert and rescale the image so that it's suitable for eBooks.
func processImage(imageFilepath string, toGrayscale bool) error {
	sigolo.Trace("Process image '%s'", imageFilepath)

	args := []string{
		imageFilepath,
		"-resize", fmt.Sprintf("%dx%d>", imgSize, imgSize),
		"-quality", "75",
		"-define", "PNG:compression-level=9",
		"-define", "PNG:compression-filter=0",
	}

	if toGrayscale {
		sigolo.Trace("Add args to convert '%s' to grayscale", imageFilepath)
		args = append(args, "-colorspace", "gray")
	}

	args = append(args, imageFilepath)

	err := util.Execute("convert", args...)

	if err != nil {
		sigolo.Error("Converting image %s failed", imageFilepath)
	}

	return err
}
