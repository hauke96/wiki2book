package util

import (
	"github.com/hauke96/sigolo"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

func MakePathRelative(styleFile string) (string, error) {
	currentDir, err := os.Getwd()
	sigolo.FatalCheck(err)

	if styleFile != "" {
		styleFile, err = filepath.Rel(currentDir, styleFile)
		sigolo.FatalCheck(errors.Wrap(err, "Unable to make style file path absolute"))
	}

	return styleFile, err
}

func MakePathAbsolute(styleFile string) (string, error) {
	var err error
	if styleFile != "" {
		styleFile, err = filepath.Abs(styleFile)
		sigolo.FatalCheck(errors.Wrap(err, "Unable to make style file path relative"))
	}
	return styleFile, err
}

func AssertFileExists(path string) {
	if _, err := os.Stat(path); strings.TrimSpace(path) != "" && err != nil {
		sigolo.Fatal("File path '%s' does not exist", path)
	}
}
