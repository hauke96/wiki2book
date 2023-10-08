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
		content := tokenizer.parseHeadings(fmt.Sprintf("%s h%d %s", headingPrefixSuffix, i, headingPrefixSuffix))
		token := fmt.Sprintf(TOKEN_TEMPLATE, fmt.Sprintf(TOKEN_HEADING_TEMPLATE, i), 0)
		test.AssertEqual(t, token, content)
		test.AssertMapEqual(t, map[string]interface{}{
			token: fmt.Sprintf("h%d", i),
		}, tokenizer.getTokenMap())
	}
}

func TestParseHeading_withFormatting(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseHeadings("== H2 ''with formatting'' ==")

	test.AssertEqual(t, "$$TOKEN_HEADING_2_0$$", content)
	test.AssertMapEqual(t, map[string]interface{}{
		"$$TOKEN_HEADING_2_0$$": "H2 $$MARKER_ITALIC_OPEN$$with formatting$$MARKER_ITALIC_CLOSE$$",
	}, tokenizer.getTokenMap())
}

func TestParseHeading_withSpacesAroundEqualCharacters(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseHeadings("  == foo == ")

	test.AssertEqual(t, "$$TOKEN_HEADING_2_0$$", content)
	test.AssertMapEqual(t, map[string]interface{}{
		"$$TOKEN_HEADING_2_0$$": "foo",
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
$$TOKEN_HEADING_2_2$$
2
$$TOKEN_HEADING_3_0$$
3
$$TOKEN_HEADING_1_3$$
1
$$TOKEN_HEADING_3_1$$
3-2`, content)
	test.AssertEqual(t, map[string]interface{}{
		"$$TOKEN_HEADING_3_0$$": "heading3",
		"$$TOKEN_HEADING_3_1$$": "heading3-2",
		"$$TOKEN_HEADING_2_2$$": "heading2",
		"$$TOKEN_HEADING_1_3$$": "heading1",
	}, tokenizer.getTokenMap())
}
