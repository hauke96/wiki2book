package cache

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
	"wiki2book/config"
	"wiki2book/util"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

const (
	TempDirName          = ".tmp"
	ArticleCacheDirName  = "articles"
	HtmlCacheDirName     = "html"
	StatsCacheDirName    = "stats"
	ImageCacheDirName    = "images"
	MathCacheDirName     = "math"
	TemplateCacheDirName = "templates"
)

var (
	cacheWriteMutex = &sync.Mutex{}
)

type Cache struct {
	configService *config.ConfigService
}

func NewCache(configService *config.ConfigService) *Cache {
	return &Cache{configService: configService}
}

func (c *Cache) GetFilePathInCache(cacheFolderName string, filename string) string {
	return filepath.Join(c.configService.Get().CacheDir, cacheFolderName, util.SanitizeFilename(filename))
}

func (c *Cache) GetRelativeFilePathInCache(cacheFolderName string, filename string) string {
	return filepath.Join(".", cacheFolderName, util.SanitizeFilename(filename))
}

// GetPathRelativeToCache returns the path relative to the given cache dir. If the given path is an absolute path to
// a file in the cache, the resulting path is a relative path to the same file within the cache.
func (c *Cache) GetPathRelativeToCache(path string) (string, error) {
	return util.ToRelativePathWithBasedir(c.configService.Get().CacheDir, path)
}

func (c *Cache) GetDirPathInCache(cacheFolderName string) string {
	return filepath.Join(c.configService.Get().CacheDir, cacheFolderName)
}

func (c *Cache) GetTempPath() string {
	return filepath.Join(c.configService.Get().CacheDir, TempDirName)
}

// CacheToFile writes the data from the reader into a file within the app cache. The cacheFolderName is the name of the
// folder within the cache, not a whole path. The filename is the name of the file in the cache. The full path is always
// returned. The error is only set when an error occurred.
func (c *Cache) CacheToFile(cacheFolderName string, filename string, reader io.Reader) (string, error) {
	cacheWriteMutex.Lock()
	defer cacheWriteMutex.Unlock()

	sanitizedFilename := util.SanitizeFilename(filename)

	outputFilepath := c.GetFilePathInCache(cacheFolderName, filename)
	sigolo.Debugf("Write data to cache file '%s'", outputFilepath)

	// Create the output folder
	outputFolderPath := c.GetDirPathInCache(cacheFolderName)
	sigolo.Tracef("Ensure cache folder '%s'", outputFolderPath)
	err := util.CurrentFilesystem.MkdirAll(outputFolderPath)
	if err != nil && !os.IsExist(err) {
		return outputFilepath, errors.Wrap(err, fmt.Sprintf("Unable to create output folder '%s'", outputFolderPath))
	}

	//
	// 1. Write to temporary file. This prevents broken files on disk in case the application exits during writing.
	//

	// Create the output file
	tempFile, err := util.CurrentFilesystem.CreateTemp(c.GetTempPath(), sanitizedFilename)
	if err != nil {
		return outputFilepath, errors.Wrap(err, fmt.Sprintf("Unable to create temporary file '%s'", filepath.Join(c.GetTempPath(), filename)))
	}
	defer tempFile.Close()
	tempFilepath := tempFile.Name()
	defer util.CurrentFilesystem.Remove(tempFilepath)
	sigolo.Tracef("Create temp file '%s'", tempFilepath)

	// Write the body to file
	sigolo.Trace("Copy data to temp file")
	_, err = io.Copy(tempFile, reader)
	if err != nil {
		return outputFilepath, errors.Wrap(err, fmt.Sprintf("Unable copy downloaded content to temp file '%s'", tempFilepath))
	}

	//
	// 2. Evict files from cache in case it's overflowing when the new file is added.
	//
	sigolo.Tracef("Caching strategy is '%s'", c.configService.Get().CacheEvictionStrategy)
	if c.configService.Get().CacheEvictionStrategy != config.CacheEvictionStrategyNone {
		var cacheSizeInBytes int64
		err, cacheSizeInBytes = util.CurrentFilesystem.DirSizeInBytes(c.configService.Get().CacheDir)
		if err != nil {
			return outputFilepath, errors.Wrap(err, fmt.Sprintf("Unable to determine size of cache folder '%s'", c.configService.Get().CacheDir))
		}

		var tempFileStat os.FileInfo
		tempFileStat, err = tempFile.Stat()
		if err != nil {
			return outputFilepath, errors.Wrap(err, fmt.Sprintf("Unable to determine size of file '%s' to cache (tmp file '%s')", sanitizedFilename, tempFilepath))
		}

		tempFileSizeInBytes := tempFileStat.Size()
		err = c.deleteFilesFromCacheIfNeeded(cacheFolderName, filename, tempFileSizeInBytes, cacheSizeInBytes)
		if err != nil {
			return outputFilepath, err
		}
	}

	// Close file as it's not needed anymore. Without closing it, Windows has problems moving the file.
	err = tempFile.Close()
	if err != nil {
		return outputFilepath, errors.Wrap(err, fmt.Sprintf("Unable to close temporary file '%s'", filepath.Join(c.GetTempPath(), filename)))
	}

	//
	// 3. Move file to actual location
	//

	sigolo.Tracef("Move temp file '%s' to '%s'", tempFilepath, outputFilepath)
	err = util.CurrentFilesystem.Rename(tempFilepath, outputFilepath)
	if err != nil {
		return outputFilepath, errors.Wrap(err, fmt.Sprintf("Error moving temp file '%s' to '%s'", tempFilepath, outputFilepath))
	}

	sigolo.Tracef("Cached file '%s' to '%s'", filename, outputFilepath)
	return outputFilepath, nil
}

func (c *Cache) CleanUpTempDir() {
	err := os.RemoveAll(c.GetTempPath())
	if err != nil {
		sigolo.Warnf("Error cleaning up '%s' directory", c.GetTempPath())
	}
}

// deleteFilesFromCacheIfNeeded deletes files from the cache based on the configured cache eviction strategy. When the
// cache is small enough for the new file, no (further) files will be deleted.
func (c *Cache) deleteFilesFromCacheIfNeeded(cacheFolderName string, newFileName string, newFileSizeInBytes int64, cacheSizeInBytes int64) error {
	// The new file might already exist in an older state (and thus with different size). The file will, therefore, not
	// just be added to the cache, but instead the old file will be replaced. The cache then grows much less in size or
	// might even shrink (in case the new file is smaller than the old one).
	var netCacheSizeChangeInBytes = newFileSizeInBytes
	existingFileSizeInBytes, err := util.CurrentFilesystem.GetSizeInBytes(c.GetFilePathInCache(cacheFolderName, newFileName))
	if err == nil {
		netCacheSizeChangeInBytes = newFileSizeInBytes - existingFileSizeInBytes
	}

	sigolo.Debugf("Max cache size: %f MB; current size: %f MB; new file size: %f MB; existing file size: %f MB (NaN means there's no existing file); net cache size change: %f MB", util.ToMB(c.configService.Get().CacheMaxSize), util.ToMB(cacheSizeInBytes), util.ToMB(newFileSizeInBytes), util.ToMB(existingFileSizeInBytes), util.ToMB(netCacheSizeChangeInBytes))
	for c.configService.Get().CacheMaxSize <= cacheSizeInBytes+netCacheSizeChangeInBytes {
		sigolo.Debugf("New file (%s ; %f MB) would exceed max cache size: Max cache size of %f MB < current size of %f MB + net size change of %f MB = new size of %f MB. Remove largest files until cache is small enough.", newFileName, util.ToMB(newFileSizeInBytes), util.ToMB(c.configService.Get().CacheMaxSize), util.ToMB(cacheSizeInBytes), util.ToMB(netCacheSizeChangeInBytes), util.ToMB(cacheSizeInBytes+netCacheSizeChangeInBytes))

		if c.configService.Get().CacheEvictionStrategy == config.CacheEvictionStrategyLargest {
			err, cacheSizeInBytes = c.deleteLargestFileFromCache(cacheSizeInBytes)
		} else if c.configService.Get().CacheEvictionStrategy == config.CacheEvictionStrategyLru {
			err, cacheSizeInBytes = c.deleteLruFileFromCache(cacheSizeInBytes)
		} else {
			sigolo.Fatalf("Unsupported cache eviction strategy '%s'. This is a Bug.", c.configService.Get().CacheEvictionStrategy)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Cache) deleteLargestFileFromCache(cacheSizeInBytes int64) (error, int64) {
	var err error
	var largestFilePath string
	var largestFileSizeInBytes int64
	err, largestFileSizeInBytes, largestFilePath = util.CurrentFilesystem.FindLargestFile(c.configService.Get().CacheDir, TempDirName)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to determine largest file in cache '%s'", c.configService.Get().CacheDir)), cacheSizeInBytes
	}

	sigolo.Debugf("Delete largest file from cache: '%s' (%f MB)", largestFilePath, util.ToMB(largestFileSizeInBytes))
	err = util.CurrentFilesystem.Remove(largestFilePath)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to remove largest file '%s'", largestFilePath)), cacheSizeInBytes
	}

	return err, cacheSizeInBytes - largestFileSizeInBytes
}

func (c *Cache) deleteLruFileFromCache(cacheSizeInBytes int64) (error, int64) {
	var err error
	var lruFilePath string
	var lruFileSizeInBytes int64
	err, lruFileSizeInBytes, lruFilePath = util.CurrentFilesystem.FindLruFile(c.configService.Get().CacheDir, TempDirName)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to determine least recently used file in cache '%s'", c.configService.Get().CacheDir)), cacheSizeInBytes
	}

	sigolo.Debugf("Delete least recently used file from cache: '%s' (%f MB)", lruFilePath, util.ToMB(lruFileSizeInBytes))
	err = util.CurrentFilesystem.Remove(lruFilePath)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to remove least recently used file '%s'", lruFilePath)), cacheSizeInBytes
	}

	return err, cacheSizeInBytes - lruFileSizeInBytes
}

// GetFile determines whether the file is caches or not. It already returns the full file path, a boolean and an error.
// The boolean only has a meaning when the error is nil. In such cases "true" means the file exists and can be used,
// "false" means the file doesn't exist. In case of an error, the boolean is always "false".
func (c *Cache) GetFile(cacheFolderName string, filename string) (string, bool, error) {
	cacheWriteMutex.Lock()
	defer cacheWriteMutex.Unlock()

	filePath := c.GetFilePathInCache(cacheFolderName, filename)

	fileIsOutdated, fileExists, err := c.isOutdated(cacheFolderName, filename)
	if err != nil {
		return filePath, false, errors.Wrapf(err, "Unable to determine if file '%s' is outdated", filename)
	}
	if !fileExists {
		// A "file not found" situation is not unusual and not considered an error. Simply return that the file doesn't exist.
		return filePath, false, nil
	}

	if fileIsOutdated {
		sigolo.Debugf("File '%s' is outdated, I'll try to remove it", filename)
		err = util.CurrentFilesystem.Remove(filePath)
		if err != nil {
			return filePath, false, errors.Wrap(err, fmt.Sprintf("Unable to remove oudated file '%s'", filePath))
		}
	}

	if c.configService.Get().CacheEvictionStrategy == config.CacheEvictionStrategyLru {
		// When using the LRU cache, update access and modification time (both, since linux usually only knows the
		// latter) to correctly determine the least recently used file.
		now := time.Now()
		err = util.CurrentFilesystem.Chtimes(filePath, now, now)
		if err != nil {
			sigolo.Warnf("Unable to update access-time of file '%s': %s. This has no direct negative effect on the further execution.", filename, err.Error())
		}
	}

	return filePath, !fileIsOutdated, nil
}

// isOutdated returns whether the file is outdated and if the file even exists. When an error is returned, both boolean
// values have no defined meaning. When the second boolean (whether the file exists) is "false", the other values
// have no defined meaning.
func (c *Cache) isOutdated(cacheFolderName string, filename string) (bool, bool, error) {
	filePath := c.GetFilePathInCache(cacheFolderName, filename)

	fileStat, err := util.CurrentFilesystem.Stat(filePath)
	if os.IsNotExist(err) {
		return false, false, nil
	}
	if err != nil {
		return false, false, errors.Wrapf(err, "Unable to determine file stats of tile '%s'", filePath)
	}

	fileAgeDuration := time.Now().Sub(fileStat.ModTime())
	fileAgeInMinutes := int64(fileAgeDuration.Minutes())
	fileIsOutdated := fileAgeInMinutes > c.configService.Get().CacheMaxAge
	sigolo.Tracef("File '%s' is outdated (age: %s, max age for files: %s)", filePath, fileAgeDuration, time.Duration(c.configService.Get().CacheMaxAge)*time.Minute)

	return fileIsOutdated, true, nil
}
