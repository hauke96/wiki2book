package parser

import (
	"fmt"
	"testing"
	"wiki2book/test"
)

func TestParseBoldAndItalic_wrongFormats(t *testing.T) {
	var tokenizer Tokenizer
	var content string

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseBoldAndItalic("'''a'''b''c")
	test.AssertEqual(t, MARKER_BOLD_OPEN+"a"+MARKER_BOLD_CLOSE+"b''c", content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseBoldAndItalic("'''a''''b")
	test.AssertEqual(t, "'''a''''b", content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseBoldAndItalic("'''a")
	test.AssertEqual(t, "'''a", content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseBoldAndItalic("It's a beautiful day")
	test.AssertEqual(t, "It's a beautiful day", content)
}

func TestParseBoldAndItalic(t *testing.T) {
	var tokenizer Tokenizer
	var content string

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseBoldAndItalic("''foo'' some text '''bar'''")
	test.AssertEqual(t, MARKER_ITALIC_OPEN+"foo"+MARKER_ITALIC_CLOSE+" some text "+MARKER_BOLD_OPEN+"bar"+MARKER_BOLD_CLOSE, content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseBoldAndItalic("'''''a'''b''")
	test.AssertEqual(t, MARKER_ITALIC_OPEN+MARKER_BOLD_OPEN+"a"+MARKER_BOLD_CLOSE+"b"+MARKER_ITALIC_CLOSE, content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseBoldAndItalic("'''''a''b'''")
	test.AssertEqual(t, MARKER_BOLD_OPEN+MARKER_ITALIC_OPEN+"a"+MARKER_ITALIC_CLOSE+"b"+MARKER_BOLD_CLOSE, content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseBoldAndItalic("''x'''a''b'''")
	test.AssertEqual(t, MARKER_ITALIC_OPEN+"x"+MARKER_BOLD_OPEN+"a"+MARKER_BOLD_CLOSE+MARKER_ITALIC_CLOSE+MARKER_BOLD_OPEN+"b"+MARKER_BOLD_CLOSE, content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseBoldAndItalic("''a'''a b'''''x")
	test.AssertEqual(t, MARKER_ITALIC_OPEN+"a"+MARKER_BOLD_OPEN+"a b"+MARKER_BOLD_CLOSE+MARKER_ITALIC_CLOSE+"x", content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseBoldAndItalic("'''foo [[bar]] abc'''")
	test.AssertEqual(t, MARKER_BOLD_OPEN+"foo [[bar]] abc"+MARKER_BOLD_CLOSE, content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseBoldAndItalic("a'''b''c'''d''e")
	test.AssertEqual(t, "a"+MARKER_BOLD_OPEN+"b"+MARKER_ITALIC_OPEN+"c"+MARKER_ITALIC_CLOSE+MARKER_BOLD_CLOSE+MARKER_ITALIC_OPEN+"d"+MARKER_ITALIC_CLOSE+"e", content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseBoldAndItalic("a''b'''c''d'''e")
	test.AssertEqual(t, "a"+MARKER_ITALIC_OPEN+"b"+MARKER_BOLD_OPEN+"c"+MARKER_BOLD_CLOSE+MARKER_ITALIC_CLOSE+MARKER_BOLD_OPEN+"d"+MARKER_BOLD_CLOSE+"e", content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseBoldAndItalic("'''''plane'''tary '''m'''ass '''o'''bject''")
	test.AssertEqual(t, MARKER_ITALIC_OPEN+MARKER_BOLD_OPEN+"plane"+MARKER_BOLD_CLOSE+"tary "+MARKER_BOLD_OPEN+"m"+MARKER_BOLD_CLOSE+"ass "+MARKER_BOLD_OPEN+"o"+MARKER_BOLD_CLOSE+"bject"+MARKER_ITALIC_CLOSE, content)
}

func TestParseParagraph(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `foo

bar
 
blubb`

	tokenizedContent := tokenizer.tokenizeContent(&tokenizer, content)

	test.AssertEqual(t, fmt.Sprintf(`foo
%s
bar
 
blubb`, MARKER_PARAGRAPH), tokenizedContent)
}

func TestParseParagraph_afterLists(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `* foo

bar



blubb`

	tokenizedContent := tokenizer.tokenizeContent(&tokenizer, content)

	test.AssertEqual(t, fmt.Sprintf(`%s
bar
%s
blubb`, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 1), MARKER_PARAGRAPH), tokenizedContent)
}

func TestParseParagraph_afterHeading(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `== foo ==

bar
== hi ==


blubb

par`

	tokenizedContent := tokenizer.tokenizeContent(&tokenizer, content)

	test.AssertEqual(t, fmt.Sprintf(`%s
bar
%s
blubb
%s
par`, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_HEADING, 0), fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_HEADING, 1), MARKER_PARAGRAPH), tokenizedContent)
}

func TestParseParagraph_beforeToken(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `bar



blubb

== hi ==
cool`

	tokenizedContent := tokenizer.tokenizeContent(&tokenizer, content)

	test.AssertEqual(t, fmt.Sprintf(`bar
%s
blubb
%s
cool`, MARKER_PARAGRAPH, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_HEADING, 0)), tokenizedContent)
}
