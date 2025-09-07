package parser

import (
	"testing"
	"wiki2book/test"
	"wiki2book/wikipedia"
)

func NewTokenizerWithMockWikipediaService() Tokenizer {
	return Tokenizer{
		tokenMap:         map[string]Token{},
		tokenCounter:     0,
		images:           []string{},
		wikipediaService: wikipedia.NewMockWikipediaService(),

		tokenizeContent: tokenizeContent,
	}
}

func TestNewTokenizer(t *testing.T) {
	tokenizer := NewTokenizerWithMockWikipediaService()
	test.AssertEqual(t, 0, tokenizer.tokenCounter)
	test.AssertMapEqual(t, map[string]Token{}, tokenizer.getTokenMap())
}

func TestSetToken(t *testing.T) {
	tokenizer := NewTokenizerWithMockWikipediaService()
	test.AssertMapEqual(t, map[string]Token{}, tokenizer.getTokenMap())

	tokenizeContentCallArgument := ""
	tokenizer.tokenizeContent = func(tokenizer *Tokenizer, content string) string {
		tokenizeContentCallArgument = content
		return "foo"
	}

	tokenizer.setToken("someKey", "tokenContent")
	test.AssertMapEqual(t, map[string]Token{"someKey": "foo"}, tokenizer.getTokenMap())
	test.AssertEqual(t, "tokenContent", tokenizeContentCallArgument)
}

func TestSetRawToken(t *testing.T) {
	tokenizer := NewTokenizerWithMockWikipediaService()
	test.AssertMapEqual(t, map[string]Token{}, tokenizer.getTokenMap())

	tokenizer.tokenizeContent = func(tokenizer *Tokenizer, content string) string {
		t.Error("This should not be called")
		t.Fail()
		return "foo"
	}

	tokenizer.setRawToken("someKey", "tokenContent")
	test.AssertMapEqual(t, map[string]Token{"someKey": "tokenContent"}, tokenizer.getTokenMap())
}
