package cache

import (
	"strings"
	"testing"
	"wiki2book/test"
	"wiki2book/util"

	"github.com/pkg/errors"
)

func TestDeleteLargestFileFromCache(t *testing.T) {
	// Arrange
	expectedPath := "some/path/to/file.txt"
	actualPath := ""
	fsMock := &util.MockFilesystem{
		FindLargestFileFunc: func(path string) (error, float64, string) {
			return nil, 0.12, expectedPath
		},
		RemoveFunc: func(path string) error {
			actualPath = path
			return nil
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	err, newCacheSize := deleteLargestFileFromCache(1.23)

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, 1.11, newCacheSize)
	test.AssertEqual(t, expectedPath, actualPath)
}

func TestDeleteLargestFileFromCache_errorFindingLargestFile(t *testing.T) {
	// Arrange
	expectedError := errors.New("test error")
	fsMock := &util.MockFilesystem{
		FindLargestFileFunc: func(path string) (error, float64, string) {
			return expectedError, -1, ""
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	err, newCacheSize := deleteLargestFileFromCache(1.23)

	// Assert
	test.AssertNotNil(t, err)
	test.AssertTrue(t, strings.Contains(err.Error(), expectedError.Error()))
	test.AssertEqual(t, 1.23, newCacheSize)
}

func TestDeleteLargestFileFromCache_errorRemovingFile(t *testing.T) {
	// Arrange
	expectedPath := "some/path/to/file.txt"
	expectedError := errors.New("test error")
	fsMock := &util.MockFilesystem{
		FindLargestFileFunc: func(path string) (error, float64, string) {
			return nil, 0.12, expectedPath
		},
		RemoveFunc: func(path string) error {
			return expectedError
		},
	}
	util.CurrentFilesystem = fsMock

	// Act
	err, newCacheSize := deleteLargestFileFromCache(1.23)

	// Assert
	test.AssertNotNil(t, err)
	test.AssertTrue(t, strings.Contains(err.Error(), expectedError.Error()))
	test.AssertEqual(t, 1.23, newCacheSize)
}
