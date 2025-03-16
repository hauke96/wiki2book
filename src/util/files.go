package util

import (
	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	TempDirName = ".tmp/"

	FileEndingSvg = ".svg"
	FileEndingPng = ".png"
	FileEndingPdf = ".pdf"
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
		err = errors.Wrapf(err, "Unable to make file path %s relative", path)
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
	if path != "" && !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		err = errors.Wrapf(err, "Unable to make file path %s absolute", path)
	}
	return path, err
}

func ToAbsolutePathWithBasedir(basedir string, path string) (string, error) {
	var err error
	if path != "" && !filepath.IsAbs(path) {
		path, err = filepath.Abs(filepath.Join(basedir, path))
		err = errors.Wrapf(err, "Unable to make file path %s absolute", filepath.Join(basedir, path))
	}
	return path, err
}

func AssertPathExists(path string) {
	if !PathExists(path) {
		sigolo.FatalCheck(errors.Errorf("Path '%s' does not exist", path))
	}
}

func PathExists(path string) bool {
	if _, err := os.Stat(path); strings.TrimSpace(path) != "" && err != nil {
		return false
	}
	return true
}

func EnsureDirectory(path string) {
	if !PathExists(path) {
		sigolo.Debugf("Create directory '%s'", path)
		err := os.MkdirAll(path, os.ModePerm)
		sigolo.FatalCheck(errors.Wrapf(err, "Error creating '%s' directory", path))
	} else {
		sigolo.Debugf("Directory '%s' already exists", path)
	}
}

// GetPngPathForPdf converts the given path of a PDF file into a PNG file.
func GetPngPathForPdf(path string) string {
	Requiref(filepath.Ext(strings.ToLower(path)) == FileEndingPdf, "Filepath must lead to a PDF file but was '%s'", path)
	return path + FileEndingPng
}

// GetPngPathForSvg converts the given path of a PDF file into a PNG file.
func GetPngPathForSvg(path string) string {
	Requiref(filepath.Ext(strings.ToLower(path)) == FileEndingSvg, "Filepath must lead to a SVG file but was '%s'", path)
	return path + FileEndingPng
}
