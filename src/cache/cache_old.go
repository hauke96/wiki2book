package cache

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"wiki2book/util"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

func CacheToFile_old(cacheFolder string, filename string, reader io.ReadCloser) error {
	outputFilepath := filepath.Join(cacheFolder, filename)
	sigolo.Debugf("Write data to cache file '%s'", outputFilepath)

	// Create the output folder
	sigolo.Tracef("Ensure cache folder '%s'", cacheFolder)
	err := os.MkdirAll(cacheFolder, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, fmt.Sprintf("Unable to create output folder '%s'", cacheFolder))
	}

	//
	// 1. Write to temporary file. This prevents broken files on disk in case the application exits during writing.
	//

	// Create the output file
	tempFile, err := os.CreateTemp(util.TempDirName, filename)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to create temporary file '%s'", filepath.Join(util.TempDirName, filename)))
	}
	tempFilepath := tempFile.Name()
	defer os.Remove(tempFilepath)
	sigolo.Tracef("Create temp file '%s'", tempFilepath)

	// Write the body to file
	sigolo.Trace("Copy data to temp file")
	_, err = io.Copy(tempFile, reader)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable copy downloaded content to temp file '%s'", tempFilepath))
	}

	//
	// 2. Move file to actual location
	//

	sigolo.Tracef("Move temp file '%s' to '%s'", tempFilepath, outputFilepath)
	err = os.Rename(tempFilepath, outputFilepath)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error moving temp file '%s' to '%s'", tempFilepath, outputFilepath))
	}

	sigolo.Tracef("Cached file '%s' to '%s'", filename, outputFilepath)
	return nil
}
