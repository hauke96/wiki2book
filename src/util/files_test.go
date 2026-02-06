package util

import (
	"path/filepath"
	"testing"
	"wiki2book/test"
)

func TestGetPngPathForFile(t *testing.T) {
	// Act
	actualPath := GetPngPathForFile("some/path/to/foo.pdf")

	// Assert
	test.AssertEqual(t, "some/path/to/foo.pdf.png", actualPath)
}

func TestIsSanitized(t *testing.T) {
	// Act & Assert
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "bar", "blubb.png")))
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "bar", "blübb.png")))
	test.AssertTrue(t, IsSanitized(filepath.Join("f.o", "b,r", "bl&bb.png")))

	// Act & Assert - Validity of file names
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "bar", "blubb.png")))
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "bar", "blu%25bb.png")))
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "bar", "blu%22bb.png")))
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "bar", "blu%7Cbb.png")))
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "bar", "blu.bb.png")))
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "bar", "blöäübb.png")))
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "bar", "bl$bb.png")))
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "bar", "bl&bb.png")))
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "bar", "blµbb.png")))
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "bar", "bl→bb.png")))
	test.AssertFalse(t, IsSanitized(filepath.Join("foo", "bar", "blu%bb.png")))
	test.AssertFalse(t, IsSanitized(filepath.Join("foo", "bar", "blu:bb.png")))
	test.AssertFalse(t, IsSanitized(filepath.Join("foo", "bar", "blu|bb.png")))

	// Act & Assert - Validity of file dirs
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "blubb", "bar.png")))
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "blu%25bb", "bar.png")))
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "blu%22bb", "bar.png")))
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "blu%7Cbb", "bar.png")))
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "blu.bb", "bar.png")))
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "blöäübb", "bar.png")))
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "bl$bb", "bar.png")))
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "bl&bb", "bar.png")))
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "blµbb", "bar.png")))
	test.AssertTrue(t, IsSanitized(filepath.Join("foo", "bl→bb", "bar.png")))
	test.AssertFalse(t, IsSanitized(filepath.Join("foo", "blu%bb", "bar.png")))
	test.AssertFalse(t, IsSanitized(filepath.Join("foo", "blu:bb", "bar.png")))
	test.AssertFalse(t, IsSanitized(filepath.Join("foo", "blu|bb", "bar.png")))
}
