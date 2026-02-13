package cache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"wiki2book/config"
	"wiki2book/test"
	"wiki2book/util"

	"github.com/pkg/errors"
)

func setupCache() (*Cache, *config.ConfigService) {
	configService := config.NewConfigService()
	configService.Get().CacheDir = "cache-dir"
	return &Cache{configService: configService}, configService
}

func TestGetFilePathInCache(t *testing.T) {
	// Arrange
	fileCache, configService := setupCache()

	// Act & Assert
	filename := "foobar.png"
	path := fileCache.GetFilePathInCache(ImageCacheDirName, filename)
	test.AssertEqual(t, filepath.Join(configService.Get().CacheDir, ImageCacheDirName, filename), path)

	filename = "fööbär.png"
	path = fileCache.GetFilePathInCache(ImageCacheDirName, filename)
	test.AssertEqual(t, filepath.Join(configService.Get().CacheDir, ImageCacheDirName, filename), path)

	filename = "123_-!§()µ→.png"
	path = fileCache.GetFilePathInCache(ImageCacheDirName, filename)
	test.AssertEqual(t, filepath.Join(configService.Get().CacheDir, ImageCacheDirName, filename), path)

	filename = "a\"b|c/d\\e.p*n:g%"
	path = fileCache.GetFilePathInCache(ImageCacheDirName, filename)
	test.AssertEqual(t, filepath.Join(configService.Get().CacheDir, ImageCacheDirName, "a%22b%7Cc%2Fd%5Ce.p%2An%3Ag%25"), path)
}

func TestGetRelativeFilePathInCache(t *testing.T) {
	// Arrange
	fileCache, _ := setupCache()

	// Act & Assert
	filename := "foobar.png"
	path := fileCache.GetRelativeFilePathInCache(ImageCacheDirName, filename)
	test.AssertEqual(t, filepath.Join(".", ImageCacheDirName, filename), path)

	filename = "fööbär.png"
	path = fileCache.GetRelativeFilePathInCache(ImageCacheDirName, filename)
	test.AssertEqual(t, filepath.Join(".", ImageCacheDirName, filename), path)

	filename = "123_-!§()µ→.png"
	path = fileCache.GetRelativeFilePathInCache(ImageCacheDirName, filename)
	test.AssertEqual(t, filepath.Join(".", ImageCacheDirName, filename), path)

	filename = "a\"b|c/d\\e.p*n:g%"
	path = fileCache.GetRelativeFilePathInCache(ImageCacheDirName, filename)
	test.AssertEqual(t, filepath.Join(".", ImageCacheDirName, "a%22b%7Cc%2Fd%5Ce.p%2An%3Ag%25"), path)
}

func TestGetPathRelativeToCache(t *testing.T) {
	// Arrange
	fileCache, configService := setupCache()
	configService.Get().CacheDir = filepath.Join("foo", "bar", "cache")

	// Act
	filename := "foobar.png"
	path, err := fileCache.GetPathRelativeToCache(filepath.Join("foo", "other-dir", filename))

	// Assert
	test.AssertEqual(t, filepath.Join("..", "..", "other-dir", filename), path)
	test.AssertNil(t, err)
}

func TestDeleteLargestFileFromCache(t *testing.T) {
	// Arrange
	expectedPath := "some/path/to/file.txt"
	actualPath := ""
	fsMock := &util.MockFilesystem{
		FindLargestFileFunc: func(path string, exceptDir string) (error, int64, string) {
			return nil, 120_000, expectedPath
		},
		RemoveFunc: func(path string) error {
			actualPath = path
			return nil
		},
	}
	util.CurrentFilesystem = fsMock
	fileCache, _ := setupCache()

	// Act
	err, newCacheSize := fileCache.deleteLargestFileFromCache(1_230_000)

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, 1_110_000, newCacheSize)
	test.AssertEqual(t, expectedPath, actualPath)
}

func TestDeleteLargestFileFromCache_errorFindingFile(t *testing.T) {
	// Arrange
	expectedError := errors.New("test error")
	fsMock := &util.MockFilesystem{
		FindLargestFileFunc: func(path string, exceptDir string) (error, int64, string) {
			return expectedError, -1, ""
		},
	}
	util.CurrentFilesystem = fsMock
	fileCache, _ := setupCache()

	// Act
	err, newCacheSize := fileCache.deleteLargestFileFromCache(1_230_000)

	// Assert
	test.AssertNotNil(t, err)
	test.AssertTrue(t, strings.Contains(err.Error(), expectedError.Error()))
	test.AssertEqual(t, 1_230_000, newCacheSize)
}

func TestDeleteLargestFileFromCache_errorRemovingFile(t *testing.T) {
	// Arrange
	expectedPath := "some/path/to/file.txt"
	expectedError := errors.New("test error")
	fsMock := &util.MockFilesystem{
		FindLargestFileFunc: func(path string, exceptDir string) (error, int64, string) {
			return nil, 120_000, expectedPath
		},
		RemoveFunc: func(path string) error {
			return expectedError
		},
	}
	util.CurrentFilesystem = fsMock
	fileCache, _ := setupCache()

	// Act
	err, newCacheSize := fileCache.deleteLargestFileFromCache(1_230_000)

	// Assert
	test.AssertNotNil(t, err)
	test.AssertTrue(t, strings.Contains(err.Error(), expectedError.Error()))
	test.AssertEqual(t, 1_230_000, newCacheSize)
}

func TestDeleteLruFromCache(t *testing.T) {
	// Arrange
	expectedPath := "some/path/to/file.txt"
	actualPath := ""
	fsMock := &util.MockFilesystem{
		FindLruFileFunc: func(path string, exceptDir string) (error, int64, string) {
			return nil, 120_000, expectedPath
		},
		RemoveFunc: func(path string) error {
			actualPath = path
			return nil
		},
	}
	util.CurrentFilesystem = fsMock
	fileCache, _ := setupCache()

	// Act
	err, newCacheSize := fileCache.deleteLruFileFromCache(1_230_000)

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, 1_110_000, newCacheSize)
	test.AssertEqual(t, expectedPath, actualPath)
}

func TestDeleteLruFromCache_errorFindingFile(t *testing.T) {
	// Arrange
	expectedError := errors.New("test error")
	fsMock := &util.MockFilesystem{
		FindLruFileFunc: func(path string, exceptDir string) (error, int64, string) {
			return expectedError, -1, ""
		},
	}
	util.CurrentFilesystem = fsMock
	fileCache, _ := setupCache()

	// Act
	err, newCacheSize := fileCache.deleteLruFileFromCache(1_230_000)

	// Assert
	test.AssertNotNil(t, err)
	test.AssertTrue(t, strings.Contains(err.Error(), expectedError.Error()))
	test.AssertEqual(t, 1_230_000, newCacheSize)
}

func TestDeleteLruFromCache_errorRemovingFile(t *testing.T) {
	// Arrange
	expectedPath := "some/path/to/file.txt"
	expectedError := errors.New("test error")
	fsMock := &util.MockFilesystem{
		FindLruFileFunc: func(path string, exceptDir string) (error, int64, string) {
			return nil, 120_000, expectedPath
		},
		RemoveFunc: func(path string) error {
			return expectedError
		},
	}
	util.CurrentFilesystem = fsMock
	fileCache, _ := setupCache()

	// Act
	err, newCacheSize := fileCache.deleteLruFileFromCache(1_230_000)

	// Assert
	test.AssertNotNil(t, err)
	test.AssertTrue(t, strings.Contains(err.Error(), expectedError.Error()))
	test.AssertEqual(t, 1_230_000, newCacheSize)
}

func TestDeleteFilesFromCacheIfNeeded_largest(t *testing.T) {
	// Arrange
	fileCache, configService := setupCache()
	configService.Get().CacheMaxSize = 5_000_000
	configService.Get().CacheEvictionStrategy = config.CacheEvictionStrategyLargest

	currentCacheSize := int64(10_000_000)
	fileSize := int64(3_000_000)

	fsMock := &util.MockFilesystem{
		GetSizeInBytesFunc: func(path string) (int64, error) {
			return 0, errors.New("test error: no existing file")
		},
		FindLargestFileFunc: func(path string, exceptDir string) (error, int64, string) {
			return nil, fileSize, "path/to/file.txt"
		},
		RemoveFunc: func(path string) error {
			currentCacheSize -= fileSize
			return nil
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	err := fileCache.deleteFilesFromCacheIfNeeded("cache/folder", "filename.txt", fileSize, currentCacheSize)

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, 1_000_000, currentCacheSize)
}

func TestDeleteFilesFromCacheIfNeeded_withExistingFileIncreasingCacheSize(t *testing.T) {
	// Arrange
	fileCache, configService := setupCache()
	configService.Get().CacheMaxSize = 5_000_000
	configService.Get().CacheEvictionStrategy = config.CacheEvictionStrategyLargest

	currentCacheSize := int64(6_600_000)
	existingFileSize := int64(1_500_000)
	fileSize := int64(2_000_000)
	removeCalls := 0

	fsMock := &util.MockFilesystem{
		GetSizeInBytesFunc: func(path string) (int64, error) {
			return existingFileSize, nil
		},
		FindLargestFileFunc: func(path string, exceptDir string) (error, int64, string) {
			return nil, fileSize, "path/to/file.txt"
		},
		RemoveFunc: func(path string) error {
			removeCalls++
			currentCacheSize -= fileSize
			return nil
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	err := fileCache.deleteFilesFromCacheIfNeeded("cache/folder", "filename.txt", fileSize, currentCacheSize)

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, 2_600_000, currentCacheSize)
	test.AssertEqual(t, 2, removeCalls)
}

func TestDeleteFilesFromCacheIfNeeded_withExistingFileReducingCacheSize(t *testing.T) {
	// Arrange
	fileCache, configService := setupCache()
	configService.Get().CacheMaxSize = 5_000_000
	configService.Get().CacheEvictionStrategy = config.CacheEvictionStrategyLargest

	currentCacheSize := int64(6_900_000)
	existingFileSize := int64(3_000_000)
	fileSize := int64(2_000_000)
	removeCalls := 0

	fsMock := &util.MockFilesystem{
		GetSizeInBytesFunc: func(path string) (int64, error) {
			return existingFileSize, nil
		},
		FindLargestFileFunc: func(path string, exceptDir string) (error, int64, string) {
			return nil, fileSize, "path/to/file.txt"
		},
		RemoveFunc: func(path string) error {
			removeCalls++
			currentCacheSize -= fileSize
			return nil
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	err := fileCache.deleteFilesFromCacheIfNeeded("cache/folder", "filename.txt", fileSize, currentCacheSize)

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, 4_900_000, currentCacheSize)
	test.AssertEqual(t, 1, removeCalls)
}

func TestDeleteFilesFromCacheIfNeeded_errorDeletingFile(t *testing.T) {
	// Arrange
	fileCache, configService := setupCache()
	configService.Get().CacheMaxSize = 5_000_000
	configService.Get().CacheEvictionStrategy = config.CacheEvictionStrategyLargest

	currentCacheSize := int64(6_900_000)
	existingFileSize := int64(3_000_000)
	fileSize := int64(2_000_000)
	expectedError := errors.New("test error")

	fsMock := &util.MockFilesystem{
		GetSizeInBytesFunc: func(path string) (int64, error) {
			return existingFileSize, nil
		},
		FindLargestFileFunc: func(path string, exceptDir string) (error, int64, string) {
			return nil, fileSize, "path/to/file.txt"
		},
		RemoveFunc: func(path string) error {
			return expectedError
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	err := fileCache.deleteFilesFromCacheIfNeeded("cache/folder", "filename.txt", fileSize, currentCacheSize)

	// Assert
	test.AssertTrue(t, strings.Contains(err.Error(), expectedError.Error()))
	test.AssertEqual(t, 6_900_000, currentCacheSize)
}

func TestDeleteFilesFromCacheIfNeeded_lru(t *testing.T) {
	// Arrange
	fileCache, configService := setupCache()
	configService.Get().CacheMaxSize = 5_000_000
	configService.Get().CacheEvictionStrategy = config.CacheEvictionStrategyLru

	currentCacheSize := int64(10_000_000)
	fileSize := int64(3_000_000)

	fsMock := &util.MockFilesystem{
		GetSizeInBytesFunc: func(path string) (int64, error) {
			return 0, errors.New("test error: no existing file")
		},
		FindLruFileFunc: func(path string, exceptDir string) (error, int64, string) {
			return nil, fileSize, "path/to/file.txt"
		},
		RemoveFunc: func(path string) error {
			currentCacheSize -= fileSize
			return nil
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	err := fileCache.deleteFilesFromCacheIfNeeded("cache/folder", "filename.txt", fileSize, currentCacheSize)

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, 1_000_000, currentCacheSize)
}

func TestGetFile(t *testing.T) {
	// Arrange
	fileCache, configService := setupCache()
	configService.Get().CacheDir = "cache-dir"
	configService.Get().CacheMaxAge = 100

	fsMock := util.NewDefaultMockFilesystem()
	fsMock.StatFunc = func(path string) (os.FileInfo, error) {
		fileInfoTime := time.Now().Add(-20 * time.Minute)
		fileInfo := util.NewMockFileInfoWithTime("file", fileInfoTime)
		return fileInfo, nil
	}
	util.CurrentFilesystem = fsMock

	// Act
	filePath, exists, err := fileCache.GetFile("cache", "file")

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, "cache-dir/cache/file", filePath)
	test.AssertTrue(t, exists)
}

func TestGetFile_outdated(t *testing.T) {
	// Arrange
	fileCache, configService := setupCache()
	configService.Get().CacheDir = "cache-dir"
	configService.Get().CacheMaxAge = 10

	fsMock := util.NewDefaultMockFilesystem()
	fsMock.StatFunc = func(path string) (os.FileInfo, error) {
		fileInfoTime := time.Now().Add(-20 * time.Minute)
		fileInfo := util.NewMockFileInfoWithTime("file", fileInfoTime)
		return fileInfo, nil
	}
	fsMock.RemoveFunc = func(name string) error {
		return nil
	}
	util.CurrentFilesystem = fsMock

	// Act
	filePath, exists, err := fileCache.GetFile("cache", "file")

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, "cache-dir/cache/file", filePath)
	test.AssertFalse(t, exists)
}

func TestGetFile_outdatedAndRmovalFailed(t *testing.T) {
	// Arrange
	fileCache, configService := setupCache()
	configService.Get().CacheDir = "cache-dir"
	configService.Get().CacheMaxAge = 10

	fsMock := &util.MockFilesystem{
		StatFunc: func(path string) (os.FileInfo, error) {
			fileInfoTime := time.Now().Add(-20 * time.Minute)
			fileInfo := util.NewMockFileInfoWithTime("file", fileInfoTime)
			return fileInfo, nil
		},
		RemoveFunc: func(name string) error {
			return errors.New("test error")
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	filePath, exists, err := fileCache.GetFile("cache", "file")

	// Assert
	test.AssertNotNil(t, err)
	test.AssertEqual(t, "cache-dir/cache/file", filePath)
	test.AssertFalse(t, exists)
}

func TestGetFile_fileNotExists(t *testing.T) {
	// Arrange
	fileCache, configService := setupCache()
	configService.Get().CacheDir = "cache-dir"
	configService.Get().CacheMaxAge = 10

	fsMock := &util.MockFilesystem{
		StatFunc: func(path string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	filePath, exists, err := fileCache.GetFile("cache", "file")

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, "cache-dir/cache/file", filePath)
	test.AssertFalse(t, exists)
}

func TestGetFile_errorGettingStats(t *testing.T) {
	// Arrange
	fileCache, configService := setupCache()
	configService.Get().CacheDir = "cache-dir"
	configService.Get().CacheMaxAge = 10

	fsMock := &util.MockFilesystem{
		StatFunc: func(path string) (os.FileInfo, error) {
			return nil, errors.New("test error")
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	filePath, _, err := fileCache.GetFile("cache", "file")

	// Assert
	test.AssertNotNil(t, err)
	test.AssertEqual(t, "cache-dir/cache/file", filePath)
}

func TestGetFile_updatingModTimeWhenUsingLruCache(t *testing.T) {
	// Arrange
	fileCache, configService := setupCache()
	configService.Get().CacheDir = "cache-dir"
	configService.Get().CacheMaxAge = 99999
	configService.Get().CacheEvictionStrategy = config.CacheEvictionStrategyLru

	chtimesCalls := 0
	chTimesCallNameParam := ""
	fsMock := &util.MockFilesystem{
		StatFunc: func(path string) (os.FileInfo, error) {
			fileInfoTime := time.Now().Add(-20 * time.Minute)
			fileInfo := util.NewMockFileInfoWithTime("file", fileInfoTime)
			return fileInfo, nil
		},
		ChtimesFunc: func(name string, atime time.Time, mtime time.Time) error {
			chtimesCalls++
			chTimesCallNameParam = name
			return nil
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	filePath, exists, err := fileCache.GetFile("cache", "file")

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, "cache-dir/cache/file", filePath)
	test.AssertTrue(t, exists)
	test.AssertEqual(t, 1, chtimesCalls)
	test.AssertEqual(t, "cache-dir/cache/file", chTimesCallNameParam)
}

func TestGetFile_updatingModTimeWhenUsingLruCache_errorUpdatingTime(t *testing.T) {
	// Arrange
	fileCache, configService := setupCache()
	configService.Get().CacheDir = "cache-dir"
	configService.Get().CacheMaxAge = 99999
	configService.Get().CacheEvictionStrategy = config.CacheEvictionStrategyLru

	fsMock := &util.MockFilesystem{
		StatFunc: func(path string) (os.FileInfo, error) {
			fileInfoTime := time.Now().Add(-20 * time.Minute)
			fileInfo := util.NewMockFileInfoWithTime("file", fileInfoTime)
			return fileInfo, nil
		},
		ChtimesFunc: func(name string, atime time.Time, mtime time.Time) error {
			return errors.New("test error")
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	filePath, exists, err := fileCache.GetFile("cache", "file")

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, "cache-dir/cache/file", filePath)
	test.AssertTrue(t, exists)
}

func TestIsOutdated_outdated(t *testing.T) {
	// Arrange
	fileCache, configService := setupCache()
	configService.Get().CacheMaxAge = 10

	fsMock := &util.MockFilesystem{
		StatFunc: func(path string) (os.FileInfo, error) {
			fileInfoTime := time.Now().Add(-20 * time.Minute)
			fileInfo := util.NewMockFileInfoWithTime("file", fileInfoTime)
			return fileInfo, nil
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	outdated, exists, err := fileCache.isOutdated("cache", "file")

	// Assert
	test.AssertNil(t, err)
	test.AssertTrue(t, outdated)
	test.AssertTrue(t, exists)
}

func TestIsOutdated_notOutdated(t *testing.T) {
	// Arrange
	fileCache, configService := setupCache()
	configService.Get().CacheMaxAge = 100

	fsMock := &util.MockFilesystem{
		StatFunc: func(path string) (os.FileInfo, error) {
			fileInfoTime := time.Now().Add(-20 * time.Minute)
			fileInfo := util.NewMockFileInfoWithTime("file", fileInfoTime)
			return fileInfo, nil
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	outdated, exists, err := fileCache.isOutdated("cache", "file")

	// Assert
	test.AssertNil(t, err)
	test.AssertFalse(t, outdated)
	test.AssertTrue(t, exists)
}

func TestIsOutdated_notExistingFile(t *testing.T) {
	// Arrange
	fileCache, configService := setupCache()
	configService.Get().CacheMaxAge = 100

	fsMock := &util.MockFilesystem{
		StatFunc: func(path string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	outdated, exists, err := fileCache.isOutdated("cache", "file")

	// Assert
	test.AssertNil(t, err)
	test.AssertFalse(t, outdated)
	test.AssertFalse(t, exists)
}

func TestCacheToFile(t *testing.T) {
	// Arrange
	fileCache, configService := setupCache()
	configService.Get().CacheDir = "cache-dir"
	configService.Get().CacheMaxAge = 100

	outputFilename := "File:ima*ge.png"
	fileContent := "foo bar"
	stringReader := strings.NewReader(fileContent)

	var mockTempFile util.FileLike

	fsMock := util.NewDefaultMockFilesystem()
	fsMock.CreateTempFunc = func(dir, pattern string) (util.FileLike, error) {
		mockTempFile = util.NewMockFile(pattern)
		return mockTempFile, nil
	}
	util.CurrentFilesystem = fsMock

	// Act
	filePath, err := fileCache.CacheToFile(ImageCacheDirName, outputFilename, stringReader)

	// Assert
	expectedOutputFilename := "File%3Aima%2Age.png"
	test.AssertNil(t, err)
	test.AssertEqual(t, "cache-dir/"+ImageCacheDirName+"/"+expectedOutputFilename, filePath)
	test.AssertEqual(t, expectedOutputFilename, mockTempFile.Name())
}
