package parser

import (
	"fmt"
	"strings"
	"testing"
	"wiki2book/test"
)

func TestParseHeading(t *testing.T) {
	for i := 1; i < 7; i++ {
		tokenizer := NewTokenizer("foo", "bar")
		headingPrefixSuffix := strings.Repeat("=", i)

		content := tokenizer.tokenizeContent(&tokenizer, fmt.Sprintf("%s heading of depth %d %s", headingPrefixSuffix, i, headingPrefixSuffix))

		token := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_HEADING, 0)
		test.AssertEqual(t, token, content)
		test.AssertMapEqual(t, map[string]Token{
			token: HeadingToken{
				Content: fmt.Sprintf("heading of depth %d", i),
				Depth:   i,
			},
		}, tokenizer.getTokenMap())
	}
}

func TestParseHeading_withFormatting(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseHeadings("== H2 ''with formatting'' ==")

	test.AssertEqual(t, "$$TOKEN_HEADING_0$$", content)
	test.AssertMapEqual(t, map[string]Token{
		"$$TOKEN_HEADING_0$$": HeadingToken{
			Content: "H2 $$MARKER_ITALIC_OPEN$$with formatting$$MARKER_ITALIC_CLOSE$$",
			Depth:   2,
		},
	}, tokenizer.getTokenMap())
}

func TestParseHeading_withSpacesAroundEqualCharacters(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseHeadings("  == foo == ")

	test.AssertEqual(t, "$$TOKEN_HEADING_0$$", content)
	test.AssertMapEqual(t, map[string]Token{
		"$$TOKEN_HEADING_0$$": HeadingToken{
			Content: "foo",
			Depth:   2,
		},
	}, tokenizer.getTokenMap())
}

func TestParseMultipleHeadings(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseHeadings(`foo
== heading2 ==
2
=== heading3 ===
3
= heading1 =
1
=== heading3-2 ===
3-2`)

	test.AssertEqual(t, `foo
$$TOKEN_HEADING_2$$
2
$$TOKEN_HEADING_0$$
3
$$TOKEN_HEADING_3$$
1
$$TOKEN_HEADING_1$$
3-2`, content)
	test.AssertMapEqual(t, map[string]Token{
		"$$TOKEN_HEADING_0$$": HeadingToken{
			Content: "heading3",
			Depth:   3,
		},
		"$$TOKEN_HEADING_1$$": HeadingToken{
			Content: "heading3-2",
			Depth:   3,
		},
		"$$TOKEN_HEADING_2$$": HeadingToken{
			Content: "heading2",
			Depth:   2,
		},
		"$$TOKEN_HEADING_3$$": HeadingToken{
			Content: "heading1",
			Depth:   1,
		},
	}, tokenizer.getTokenMap())
}
