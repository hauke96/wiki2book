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
