package parser

import (
	"fmt"
	"testing"
	"wiki2book/test"
	"wiki2book/wikipedia"
)

func TestNowiki(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})
	content := "Foo<nowiki>something</nowiki> bar"
	expectedContent := "Foo" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_NOWIKI, 0) + " bar"

	newContent := tokenizer.parseNowiki(content)

	test.AssertEqual(t, expectedContent, newContent)
	test.AssertMapEqual(t, map[string]Token{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_NOWIKI, 0): NowikiToken{Content: "something"},
	}, tokenizer.getTokenMap())
}

func TestNowiki_endOfText(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})
	content := "Foo<nowiki>something</nowiki>"
	expectedContent := "Foo" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_NOWIKI, 0)

	newContent := tokenizer.parseNowiki(content)

	test.AssertEqual(t, expectedContent, newContent)
	test.AssertMapEqual(t, map[string]Token{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_NOWIKI, 0): NowikiToken{Content: "something"},
	}, tokenizer.getTokenMap())
}
