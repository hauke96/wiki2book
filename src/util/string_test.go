package util

import (
	"testing"
	"wiki2book/test"
)

func TestElementHasPrefix(t *testing.T) {
	prefixe := []string{"f", "fo", "foo", "foo!"}

	element := "foo"
	hasPrefix := HasAnyPrefix(element, prefixe...)
	test.AssertTrue(t, hasPrefix)

	element = "oo"
	hasPrefix = HasAnyPrefix(element, prefixe...)
	test.AssertFalse(t, hasPrefix)
}

func TestGetTextAround(t *testing.T) {
	result := GetTextAround("foo bar 123", 5, 2)
	test.AssertEqual(t, " bar ", result)

	result = GetTextAround("foo bar 123", 0, 0)
	test.AssertEqual(t, "f", result)

	result = GetTextAround("foo bar 123", 0, 2)
	test.AssertEqual(t, "foo", result)

	result = GetTextAround("foo bar 123", 10, 0)
	test.AssertEqual(t, "3", result)

	result = GetTextAround("foo bar 123", 10, 2)
	test.AssertEqual(t, "123", result)

	result = GetTextAround("foo bar 123", 5, 20)
	test.AssertEqual(t, "foo bar 123", result)
}
