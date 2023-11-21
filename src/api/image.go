package api

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"wiki2book/util"
)

const imgSize = 600

// processImage will convert and rescale the image so that it's suitable for eBooks.
func processImage(imageFilepath string) error {
	sigolo.Trace("Process image '%s'", imageFilepath)

	err := util.Execute("convert", imageFilepath, "-colorspace", "gray", "-separate", "-average", "-resize", fmt.Sprintf("%dx%d>", imgSize, imgSize), "-quality", "75",
		"-define", "PNG:compression-level=9", "-define", "PNG:compression-filter=0", imageFilepath)

	if err != nil {
		sigolo.Error("Converting image %s failed", imageFilepath)
	}

	return err
}
