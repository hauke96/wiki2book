package parser

import (
	"github.com/hauke96/wiki2book/src/test"
	"testing"
)

func TestImageRegex(t *testing.T) {
	valid := []string{
		"[[Datei:foo]]",
		"[[Datei:foo.png]]",
		"[[Datei:foo.png|mini]]",
		"[[Datei:foo|mini]]",
		"[[Datei:foo.jpg|mini|16px]]",
		"[[File:foo.png]]",
	}

	for _, i := range valid {
		test.AssertMatch(t, IMAGE_REGEX_PATTERN, i)
	}

	invalid := []string{
		"",
		"Datei.foo.png",
		"[Datei:foo.png]",
		"[[Fiel:foo.png]]",
		"[[foo.png]]",
	}

	for _, i := range invalid {
		test.AssertNoMatch(t, IMAGE_REGEX_PATTERN, i)
	}
}
