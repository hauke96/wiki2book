package parser

import (
	"testing"
	"wiki2book/test"
)

func TestParseInternalLinks(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")

	content := tokenizer.parseInternalLinks("foo [[internal link]]")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_INTERNAL_LINK+"_0$$", content)
	test.AssertMapEqual(t, map[string]interface{}{
		"$$TOKEN_" + TOKEN_INTERNAL_LINK + "_0$$": InternalLinkToken{
			ArticleName: "internal link",
			LinkText:    "internal link",
		},
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseInternalLinks("foo [[internal link|bar]] abc")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_INTERNAL_LINK+"_0$$ abc", content)
	test.AssertMapEqual(t, map[string]interface{}{
		"$$TOKEN_" + TOKEN_INTERNAL_LINK + "_0$$": InternalLinkToken{
			ArticleName: "internal link",
			LinkText:    "bar",
		},
	}, tokenizer.getTokenMap())
}

func TestParseInternalLinks_withFile(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseInternalLinks("foo [[file:external-link.jpg|bar]]")
	test.AssertEqual(t, "foo [[file:external-link.jpg|bar]]", content)
	test.AssertMapEqual(t, map[string]interface{}{}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseInternalLinks("foo [[file:external-link.jpg|foo [[bar]]]]")
	test.AssertEqual(t, "foo [[file:external-link.jpg|foo $$TOKEN_"+TOKEN_INTERNAL_LINK+"_0$$]]", content)
	test.AssertMapEqual(t, map[string]interface{}{
		"$$TOKEN_" + TOKEN_INTERNAL_LINK + "_0$$": InternalLinkToken{
			ArticleName: "bar",
			LinkText:    "bar",
		},
	}, tokenizer.getTokenMap())
}

func TestParseInternalLinks_withSectionReference(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")

	content := tokenizer.parseInternalLinks("foo [[article#section]]")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_INTERNAL_LINK+"_0$$", content)
	test.AssertMapEqual(t, map[string]interface{}{
		"$$TOKEN_" + TOKEN_INTERNAL_LINK + "_0$$": InternalLinkToken{
			ArticleName: "article",
			LinkText:    "article",
		},
	}, tokenizer.getTokenMap())
}

func TestParseInternalLinks_externalLinkWillNotBeTouched(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseInternalLinks("foo [http://bar.com website]")
	test.AssertEqual(t, "foo [http://bar.com website]", content)
	test.AssertMapEqual(t, map[string]interface{}{}, tokenizer.getTokenMap())
}

func TestParseMixedLinks(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseInternalLinks("foo [http://bar.com Has [[internal link]]]")
	test.AssertEqual(t, "foo [http://bar.com Has $$TOKEN_"+TOKEN_INTERNAL_LINK+"_0$$]", content)
	test.AssertMapEqual(t, map[string]interface{}{
		"$$TOKEN_" + TOKEN_INTERNAL_LINK + "_0$$": InternalLinkToken{
			ArticleName: "internal link",
			LinkText:    "internal link",
		},
	}, tokenizer.getTokenMap())
}

func TestParseExternalLinks(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseExternalLinks("foo [http://bar.com website]")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_EXTERNAL_LINK+"_0$$", content)
	test.AssertMapEqual(t, map[string]interface{}{
		"$$TOKEN_" + TOKEN_EXTERNAL_LINK + "_0$$": ExternalLinkToken{
			URL:      "http://bar.com",
			LinkText: "website",
		},
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseExternalLinks("foo [http://bar.com website] abc")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_EXTERNAL_LINK+"_0$$ abc", content)
	test.AssertMapEqual(t, map[string]interface{}{
		"$$TOKEN_" + TOKEN_EXTERNAL_LINK + "_0$$": ExternalLinkToken{
			URL:      "http://bar.com",
			LinkText: "website",
		},
	}, tokenizer.getTokenMap())
}

func TestParseExternalLinks_simpleBracketsNotRegisteredAsLinks(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseExternalLinks("Simple [brackets] will stay as is.")
	test.AssertEqual(t, "Simple [brackets] will stay as is.", content)
	test.AssertMapEqual(t, map[string]interface{}{}, tokenizer.getTokenMap())
}
