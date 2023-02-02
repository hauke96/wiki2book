package parser

import (
	"fmt"
	"github.com/hauke96/wiki2book/src/test"
	"testing"
)

func TestParseInternalLinks(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")

	content := tokenizer.parseInternalLinks("foo [[internal link]]")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_INTERNAL_LINK+"_2$$", content)
	test.AssertEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_INTERNAL_LINK + "_2$$":         fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_INTERNAL_LINK_ARTICLE, 0, TOKEN_INTERNAL_LINK_TEXT, 1),
		"$$TOKEN_" + TOKEN_INTERNAL_LINK_ARTICLE + "_0$$": "internal link",
		"$$TOKEN_" + TOKEN_INTERNAL_LINK_TEXT + "_1$$":    "internal link",
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseInternalLinks("foo [[internal link|bar]] abc")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_INTERNAL_LINK+"_2$$ abc", content)
	test.AssertEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_INTERNAL_LINK + "_2$$":         fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_INTERNAL_LINK_ARTICLE, 0, TOKEN_INTERNAL_LINK_TEXT, 1),
		"$$TOKEN_" + TOKEN_INTERNAL_LINK_ARTICLE + "_0$$": "internal link",
		"$$TOKEN_" + TOKEN_INTERNAL_LINK_TEXT + "_1$$":    "bar",
	}, tokenizer.getTokenMap())
}

func TestParseInternalLinks_withFile(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseInternalLinks("foo [[Datei:external-link.jpg|bar]]")
	test.AssertEqual(t, "foo [[Datei:external-link.jpg|bar]]", content)
	test.AssertEqual(t, map[string]string{}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseInternalLinks("foo [http://bar.com website]")
	test.AssertEqual(t, "foo [http://bar.com website]", content)
	test.AssertEqual(t, map[string]string{}, tokenizer.getTokenMap())
}

func TestParseMixedLinks(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseInternalLinks("foo [http://bar.com Has [[internal link]]]")
	test.AssertEqual(t, "foo [http://bar.com Has $$TOKEN_"+TOKEN_INTERNAL_LINK+"_2$$]", content)
	test.AssertEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_INTERNAL_LINK + "_2$$":         fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_INTERNAL_LINK_ARTICLE, 0, TOKEN_INTERNAL_LINK_TEXT, 1),
		"$$TOKEN_" + TOKEN_INTERNAL_LINK_ARTICLE + "_0$$": "internal link",
		"$$TOKEN_" + TOKEN_INTERNAL_LINK_TEXT + "_1$$":    "internal link",
	}, tokenizer.getTokenMap())
}

func TestParseExternalLinks(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseExternalLinks("foo [http://bar.com website]")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_EXTERNAL_LINK+"_2$$", content)
	test.AssertEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_EXTERNAL_LINK + "_2$$":      fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_EXTERNAL_LINK_URL, 0, TOKEN_EXTERNAL_LINK_TEXT, 1),
		"$$TOKEN_" + TOKEN_EXTERNAL_LINK_URL + "_0$$":  "http://bar.com",
		"$$TOKEN_" + TOKEN_EXTERNAL_LINK_TEXT + "_1$$": "website",
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseExternalLinks("foo [http://bar.com website] abc")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_EXTERNAL_LINK+"_2$$ abc", content)
	test.AssertEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_EXTERNAL_LINK + "_2$$":      fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_EXTERNAL_LINK_URL, 0, TOKEN_EXTERNAL_LINK_TEXT, 1),
		"$$TOKEN_" + TOKEN_EXTERNAL_LINK_URL + "_0$$":  "http://bar.com",
		"$$TOKEN_" + TOKEN_EXTERNAL_LINK_TEXT + "_1$$": "website",
	}, tokenizer.getTokenMap())
}
