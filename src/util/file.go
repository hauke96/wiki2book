package util

import (
	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

func ToRelativePaths(paths ...string) ([]string, error) {
	var result = make([]string, len(paths))

	for i, path := range paths {
		relativePath, err := ToRelativePath(path)
		if err != nil {
			return nil, err
		}
		result[i] = relativePath
	}

	return result, nil
}

func ToRelativePath(path string) (string, error) {
	currentDir, err := os.Getwd()
	sigolo.FatalCheck(err)

	if path != "" {
		path, err = filepath.Rel(currentDir, path)
		sigolo.FatalCheck(errors.Wrap(err, "Unable to make style file path absolute"))
	}

	return path, err
}

func ToAbsolutePaths(paths ...string) ([]string, error) {
	var result = make([]string, len(paths))

	for i, path := range paths {
		absolutePath, err := ToAbsolutePath(path)
		if err != nil {
			return nil, err
		}
		result[i] = absolutePath
	}

	return result, nil
}

func ToAbsolutePath(path string) (string, error) {
	var err error
	if path != "" {
		path, err = filepath.Abs(path)
		sigolo.FatalCheck(errors.Wrap(err, "Unable to make style file path relative"))
	}
	return path, err
}

func AssertFileExists(path string) {
	if _, err := os.Stat(path); strings.TrimSpace(path) != "" && err != nil {
		sigolo.Fatalf("File path '%s' does not exist", path)
	}
}
