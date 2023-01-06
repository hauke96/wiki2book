package util

import (
	"github.com/hauke96/wiki2book/src/test"
	"testing"
)

func TestElementHasPrefix(t *testing.T) {
	prefixe := []string{"f", "fo", "foo", "foo!"}

	element := "foo"
	hasPrefix := ElementHasPrefix(element, prefixe)
	test.AssertTrue(t, hasPrefix)

	element = "oo"
	hasPrefix = ElementHasPrefix(element, prefixe)
	test.AssertFalse(t, hasPrefix)
}
