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

func TestNewTokenizer(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	test.AssertEqual(t, "foo", tokenizer.imageFolder)
	test.AssertEqual(t, "bar", tokenizer.templateFolder)
	test.AssertEqual(t, 0, tokenizer.tokenCounter)
	test.AssertEqual(t, map[string]string{}, tokenizer.getTokenMap())
}

func TestSetToken(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	test.AssertEqual(t, map[string]string{}, tokenizer.getTokenMap())

	tokenizeContentCallArgument := ""
	tokenizer.tokenizeContent = func(tokenizer *Tokenizer, content string) string {
		tokenizeContentCallArgument = content
		return "foo"
	}

	tokenizer.setToken("someKey", "tokenContent")
	test.AssertEqual(t, map[string]string{"someKey": "foo"}, tokenizer.getTokenMap())
	test.AssertEqual(t, "tokenContent", tokenizeContentCallArgument)
}

func TestSetRawToken(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	test.AssertEqual(t, map[string]string{}, tokenizer.getTokenMap())

	tokenizer.tokenizeContent = func(tokenizer *Tokenizer, content string) string {
		t.Error("This should not be called")
		t.Fail()
		return "foo"
	}

	tokenizer.setRawToken("someKey", "tokenContent")
	test.AssertEqual(t, map[string]string{"someKey": "tokenContent"}, tokenizer.getTokenMap())
}
