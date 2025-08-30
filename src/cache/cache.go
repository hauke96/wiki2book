package cache

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"wiki2book/config"
	"wiki2book/util"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

// CacheToFile writes the data from the reader into a file within the app cache. The cacheFolderName is the name of the
// folder within the cache, not a whole path. The filename is the name of the file in the cache.
func CacheToFile(cacheFolderName string, filename string, reader io.ReadCloser) error {
	outputFilepath := filepath.Join(config.Current.CacheDir, cacheFolderName, filename)
	sigolo.Debugf("Write data to cache file '%s'", outputFilepath)

	// Create the output folder
	sigolo.Tracef("Ensure cache folder '%s'", cacheFolderName)
	err := os.MkdirAll(cacheFolderName, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, fmt.Sprintf("Unable to create output folder '%s'", cacheFolderName))
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
	// 2. Evict files from cache in case it's overflowing when the new file is added.
	//
	sigolo.Tracef("Caching strategy is '%s'", config.Current.CacheEvictionStrategy)
	if config.Current.CacheEvictionStrategy != "none" {
		var cacheSizeInMB float64
		err, cacheSizeInMB = dirSizeMB(config.Current.CacheDir)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Unable to determine size of cache folder '%s'", config.Current.CacheDir))
		}

		var tempFileStat os.FileInfo
		tempFileStat, err = tempFile.Stat()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Unable to determine size of file '%s' to cache (tmp file '%s')", filename, tempFilepath))
		}

		tempFileSizeInMB := float64(tempFileStat.Size()) / 1024.0 / 1024.0
		if config.Current.CacheEvictionStrategy == "largest" {

			var netCacheSizeChangeInMB = tempFileSizeInMB
			var existingFileStat os.FileInfo
			var existingFileSizeInMB = -1.0
			existingFileStat, err = os.Stat(filepath.Join(config.Current.CacheDir, cacheFolderName, filename))
			if existingFileStat != nil {
				existingFileSizeInMB = float64(existingFileStat.Size()) / 1024.0 / 1024.0
				netCacheSizeChangeInMB = existingFileSizeInMB - tempFileSizeInMB
			}

			sigolo.Debugf("Max cache size: %f MB; current size: %f MB; new file size: %f MB; existing file size: %f MB (-1 means there's no existing file); net cache size change: %f MB", config.Current.CacheMaxSize, cacheSizeInMB, tempFileSizeInMB, existingFileSizeInMB, netCacheSizeChangeInMB)
			for config.Current.CacheMaxSize < cacheSizeInMB+netCacheSizeChangeInMB {
				sigolo.Debugf("New file (%f MB) would exceed max cache size: Max cache size of %f MB < current size of %f MB + net size change of %f MB = new size of %f MB. Remove largest files until cache is small enough.", tempFileSizeInMB, config.Current.CacheMaxSize, cacheSizeInMB, netCacheSizeChangeInMB, cacheSizeInMB+netCacheSizeChangeInMB)
				var largestFileSizeInMB float64
				var largestFilePath string
				err, largestFileSizeInMB, largestFilePath = findLargestFile(config.Current.CacheDir)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("Unable to determine largest file in cache '%s'", config.Current.CacheDir))
				}

				sigolo.Debugf("Delete largest file from cache: '%s' (%f MB)", largestFilePath, largestFileSizeInMB)
				err = os.Remove(largestFilePath)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("Unable to remove largest file '%s'", largestFilePath))
				}

				cacheSizeInMB -= largestFileSizeInMB
			}
		}
		// TODO Add LRU support
	}

	//
	// 3. Move file to actual location
	//

	sigolo.Tracef("Move temp file '%s' to '%s'", tempFilepath, outputFilepath)
	err = os.Rename(tempFilepath, outputFilepath)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error moving temp file '%s' to '%s'", tempFilepath, outputFilepath))
	}

	sigolo.Tracef("Cached file '%s' to '%s'", filename, outputFilepath)
	return nil
}

func dirSizeMB(path string) (error, float64) {
	var dirSizeBytes int64 = 0

	readSize := func(path string, file os.FileInfo, err error) error {
		if !file.IsDir() {
			dirSizeBytes += file.Size()
		}
		return nil
	}

	err := filepath.Walk(path, readSize)
	if err != nil {
		return err, -1
	}

	return nil, float64(dirSizeBytes) / 1024.0 / 1024.0
}

func findLargestFile(path string) (error, float64, string) {
	var currentLargestFile os.FileInfo
	var currentLargestFilePath string

	readSize := func(path string, file os.FileInfo, err error) error {
		if !file.IsDir() {
			if currentLargestFile == nil || file.Size() > currentLargestFile.Size() {
				currentLargestFile = file
				currentLargestFilePath = path
			}
		} else if file.Name() == util.TempDirName {
			// The directory for temporary files might be inside the cache folder. This is fine, but we don't want to
			// count in temporary files then.
			return filepath.SkipDir
		}
		return nil
	}

	err := filepath.Walk(path, readSize)
	if err != nil {
		return err, -1, ""
	}

	return nil, float64(currentLargestFile.Size()) / 1024.0 / 1024.0, currentLargestFilePath
}
