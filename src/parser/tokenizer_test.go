package parser

import (
	"testing"
	"wiki2book/test"
)

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
