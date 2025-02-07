package api

import (
	"fmt"
	"github.com/hauke96/sigolo/v2"
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

	err := util.Execute("convert", args...)

	if err != nil {
		sigolo.Errorf("Converting image %s failed", imageFilepath)
	}

	return err
}
