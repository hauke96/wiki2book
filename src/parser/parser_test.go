package parser

import (
	"github.com/hauke96/wiki2book/src/test"
	"testing"
)

type MockTokenizer struct {
}

func (m *MockTokenizer) tokenize(content string) string {
	images = []string{"Datei:foo.png"}
	return "foobar"
}

func (m *MockTokenizer) getTokenMap() map[string]string {
	return map[string]string{
		"foo": "bar",
	}
}

func TestParse(t *testing.T) {
	tokenizer := MockTokenizer{}
	title := "tiiitle"

	article := Parse("Test content", title, &tokenizer)

	test.AssertEqual(t,
		Article{
			Title:    title,
			Content:  "foobar",
			TokenMap: map[string]string{"foo": "bar"},
			Images:   []string{"Datei:foo.png"},
		},
		article)
}

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
		test.AssertMatch(t, IMAGE_REGEX, i)
	}

	invalid := []string{
		"",
		"Datei.foo.png",
		"[Datei:foo.png]",
		"[[Fiel:foo.png]]",
		"[[foo.png]]",
	}

	for _, i := range invalid {
		test.AssertNoMatch(t, IMAGE_REGEX, i)
	}
}
