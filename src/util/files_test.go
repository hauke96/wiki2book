package util

import (
	"testing"
	"wiki2book/test"
)

func TestGetPngPathForFile(t *testing.T) {
	// Act
	actualPath := GetPngPathForFile("some/path/to/foo.pdf")

	// Assert
	test.AssertEqual(t, "some/path/to/foo.pdf.png", actualPath)
}
