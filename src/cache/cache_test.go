package cache

import (
	"strings"
	"testing"
	"wiki2book/config"
	"wiki2book/test"
	"wiki2book/util"

	"github.com/pkg/errors"
)

func TestDeleteLargestFileFromCache(t *testing.T) {
	// Arrange
	expectedPath := "some/path/to/file.txt"
	actualPath := ""
	fsMock := &util.MockFilesystem{
		FindLargestFileFunc: func(path string) (error, int64, string) {
			return nil, 120_000, expectedPath
		},
		RemoveFunc: func(path string) error {
			actualPath = path
			return nil
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	err, newCacheSize := deleteLargestFileFromCache(1_230_000)

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, 1_110_000, newCacheSize)
	test.AssertEqual(t, expectedPath, actualPath)
}

func TestDeleteLargestFileFromCache_errorFindingLargestFile(t *testing.T) {
	// Arrange
	expectedError := errors.New("test error")
	fsMock := &util.MockFilesystem{
		FindLargestFileFunc: func(path string) (error, int64, string) {
			return expectedError, -1, ""
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	err, newCacheSize := deleteLargestFileFromCache(1_230_000)

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
		FindLargestFileFunc: func(path string) (error, int64, string) {
			return nil, 120_000, expectedPath
		},
		RemoveFunc: func(path string) error {
			return expectedError
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	err, newCacheSize := deleteLargestFileFromCache(1_230_000)

	// Assert
	test.AssertNotNil(t, err)
	test.AssertTrue(t, strings.Contains(err.Error(), expectedError.Error()))
	test.AssertEqual(t, 1_230_000, newCacheSize)
}

func TestHandleLargestFileEvictionStrategy(t *testing.T) {
	// Arrange
	config.Current.CacheMaxSize = 5_000_000
	config.Current.CacheEvictionStrategy = "largest"

	currentCacheSize := int64(10_000_000)
	fileSize := int64(3_000_000)

	fsMock := &util.MockFilesystem{
		GetSizeInBytesFunc: func(path string) (int64, error) {
			return 0, errors.New("test error: no existing file")
		},
		FindLargestFileFunc: func(path string) (error, int64, string) {
			return nil, fileSize, "path/to/file.txt"
		},
		RemoveFunc: func(path string) error {
			currentCacheSize -= fileSize
			return nil
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	err := handleLargestFileEvictionStrategy("cache/folder", "filename.txt", fileSize, currentCacheSize)

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, 1_000_000, currentCacheSize)
}

func TestHandleLargestFileEvictionStrategy_withExistingFile(t *testing.T) {
	// Arrange
	config.Current.CacheMaxSize = 5_000_000
	config.Current.CacheEvictionStrategy = "largest"

	currentCacheSize := int64(6_000_000)
	existingFileSize := int64(1500_000)
	fileSize := int64(2_000_000)

	fsMock := &util.MockFilesystem{
		GetSizeInBytesFunc: func(path string) (int64, error) {
			return existingFileSize, nil
		},
		FindLargestFileFunc: func(path string) (error, int64, string) {
			return nil, fileSize, "path/to/file.txt"
		},
		RemoveFunc: func(path string) error {
			currentCacheSize -= fileSize
			return nil
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	err := handleLargestFileEvictionStrategy("cache/folder", "filename.txt", fileSize, currentCacheSize)

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, 4_000_000, currentCacheSize)
}
