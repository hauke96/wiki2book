package parser

import (
	"fmt"
	"github.com/hauke96/wiki2book/src/test"
	"testing"
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

func TestParseBoldAndItalic(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseBoldAndItalic("''foo'' some text '''bar'''")
	test.AssertEqual(t, MARKER_ITALIC_OPEN+"foo"+MARKER_ITALIC_CLOSE+" some text "+MARKER_BOLD_OPEN+"bar"+MARKER_BOLD_CLOSE, content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseBoldAndItalic("'''''a'''b''")
	test.AssertEqual(t, MARKER_BOLD_OPEN+MARKER_ITALIC_OPEN+"a"+MARKER_BOLD_CLOSE+"b"+MARKER_ITALIC_CLOSE, content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseBoldAndItalic("'''''a''b'''")
	test.AssertEqual(t, MARKER_BOLD_OPEN+MARKER_ITALIC_OPEN+"a"+MARKER_ITALIC_CLOSE+"b"+MARKER_BOLD_CLOSE, content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseBoldAndItalic("'''a''b'''c''")
	test.AssertEqual(t, MARKER_BOLD_OPEN+"a"+MARKER_ITALIC_OPEN+"b"+MARKER_BOLD_CLOSE+"c"+MARKER_ITALIC_CLOSE, content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseBoldAndItalic("'''foo [[bar]] abc'''")
	test.AssertEqual(t, MARKER_BOLD_OPEN+"foo [[bar]] abc"+MARKER_BOLD_CLOSE, content)
}

func TestParseImages(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseImages("foo [[Datei:image.jpg]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE_INLINE+"_1$$ bar", content)

	for _, param := range imageNonInlineParameters {
		tokenizer = NewTokenizer("foo", "bar")
		content = tokenizer.parseImages(fmt.Sprintf("foo [[Datei:image.jpg|%s]] bar", param))
		test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_1$$ bar", content)
	}

	for _, param := range imageIgnoreParameters {
		tokenizer := NewTokenizer("foo", "bar")
		content = tokenizer.parseImages(fmt.Sprintf("foo [[Datei:image.jpg|%s]] bar", param))
		test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE_INLINE+"_1$$ bar", content)
	}

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("foo [[Datei:image.jpg|10x20px|mini|some caption]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_3$$ bar", content)
	test.AssertEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_IMAGE + "_3$$":          fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_IMAGE_FILENAME, 0, TOKEN_IMAGE_CAPTION, 2, TOKEN_IMAGE_SIZE, 1),
		"$$TOKEN_" + TOKEN_IMAGE_FILENAME + "_0$$": "foo/image.jpg",
		"$$TOKEN_" + TOKEN_IMAGE_CAPTION + "_2$$":  "some caption",
		"$$TOKEN_" + TOKEN_IMAGE_SIZE + "_1$$":     "10x20",
	}, tokenizer.getTokenMap())
}

func TestElementHasPrefix(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	prefixe := []string{"f", "fo", "foo", "foo!"}

	element := "foo"
	hasPrefix := tokenizer.elementHasPrefix(element, prefixe)
	test.AssertTrue(t, hasPrefix)

	element = "oo"
	hasPrefix = tokenizer.elementHasPrefix(element, prefixe)
	test.AssertFalse(t, hasPrefix)
}

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

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseInternalLinks("foo [[Datei:external-link.jpg|bar]]")
	test.AssertEqual(t, "foo [[Datei:external-link.jpg|bar]]", content)
	test.AssertEqual(t, map[string]string{}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseInternalLinks("foo [http://bar.com website]")
	test.AssertEqual(t, "foo [http://bar.com website]", content)
	test.AssertEqual(t, map[string]string{}, tokenizer.getTokenMap())
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

func TestParseTable(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `before
{| class="wikitable"
|+ capti0n
|-
! head1 !! '''head2'''
|-
| foo [[internal]] || bar
|-
| This row
is
| multi-line wikitext
|-
| colspan="42" style="text-align:right; background: white;" | some || colspan="1" | attributes 
|}
after`
	tokenizedTable := tokenizer.parseTables(content)

	test.AssertEqual(t, fmt.Sprintf("before\n"+TOKEN_TEMPLATE+"\nafter", TOKEN_TABLE, 18), tokenizedTable)
	test.AssertEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 18): fmt.Sprintf(
			TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE,
			TOKEN_TABLE_CAPTION, 0, TOKEN_TABLE_ROW, 3, TOKEN_TABLE_ROW, 9, TOKEN_TABLE_ROW, 12, TOKEN_TABLE_ROW, 17,
		),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_CAPTION, 0): " capti0n",
		// row 0: heading
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 3):  fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_HEAD, 1, TOKEN_TABLE_HEAD, 2),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_HEAD, 1): " head1 ",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_HEAD, 2): " " + MARKER_BOLD_OPEN + "head2" + MARKER_BOLD_CLOSE,
		// row 1: internal link
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 9):             fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_COL, 7, TOKEN_TABLE_COL, 8),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 7):             fmt.Sprintf(" foo "+TOKEN_TEMPLATE+" ", TOKEN_INTERNAL_LINK, 6),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_INTERNAL_LINK, 6):         fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_INTERNAL_LINK_ARTICLE, 4, TOKEN_INTERNAL_LINK_TEXT, 5),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_INTERNAL_LINK_ARTICLE, 4): "internal",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_INTERNAL_LINK_TEXT, 5):    "internal",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 8):             " bar",
		// row 2: multi-line
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 12): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_COL, 10, TOKEN_TABLE_COL, 11),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 10): " This row\nis",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 11): " multi-line wikitext",
		// row 3: attributes
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 17):            fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_COL, 14, TOKEN_TABLE_COL, 16),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 14):            fmt.Sprintf(TOKEN_TEMPLATE+" some ", TOKEN_TABLE_COL_ATTRIBUTES, 13),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL_ATTRIBUTES, 13): "colspan=\"42\" style=\"text-align:right;\"",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 16):            fmt.Sprintf(TOKEN_TEMPLATE+" attributes ", TOKEN_TABLE_COL_ATTRIBUTES, 15),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL_ATTRIBUTES, 15): "colspan=\"1\"",
	}, tokenizer.getTokenMap())
}

func TestParseTable_tableInTable(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `{| class="wikitable"
|-
| foo ||
{| class="wikitable"
|-
| inner || table
|}
|}`
	tokenizedTable := tokenizer.parseTables(content)

	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 7), tokenizedTable)
	test.AssertEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 7): fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 6),
		// outer table
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 6): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_COL, 4, TOKEN_TABLE_COL, 5),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 4): " foo ",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 5): fmt.Sprintf("\n"+TOKEN_TEMPLATE, TOKEN_TABLE, 3),
		// inner table
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 3):     fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 2),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 2): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_COL, 0, TOKEN_TABLE_COL, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 0): " inner ",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 1): " table",
	}, tokenizer.getTokenMap())
}

func TestTokenizeTableRow_withHead(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	lines := []string{
		"! foo",
		"!bar",
		"|-",
	}
	tokenizedColumn, i := tokenizer.tokenizeTableRow(lines, 0, "!")
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 2), tokenizedColumn)
	test.AssertEqual(t, 1, i)
	test.AssertEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_HEAD, 0): " foo",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_HEAD, 1): "bar",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 2):  fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_HEAD, 0, TOKEN_TABLE_HEAD, 1),
	}, tokenizer.getTokenMap())
}

func TestTokenizeTableRow_withColumn(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	lines := []string{
		"| foo",
		"|bar",
		"| colspan=\"2\"| abc",
		"def",
		"|-",
		"| this row should be ignored",
	}
	tokenizedColumn, i := tokenizer.tokenizeTableRow(lines, 0, "|")
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 4), tokenizedColumn)
	test.AssertEqual(t, 3, i)
	test.AssertEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 0):            " foo",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 1):            "bar",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL_ATTRIBUTES, 2): `colspan="2"`,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 3):            fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL_ATTRIBUTES, 2) + " abc\ndef",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 4):            fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_COL, 0, TOKEN_TABLE_COL, 1, TOKEN_TABLE_COL, 3),
	}, tokenizer.getTokenMap())
}

func TestTokenizeTableColumn(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `colspan="2" style="text-align:center; background:Lightgray;" | ''foo'' bar`
	tokenizedColumn, css := tokenizer.tokenizeTableColumn(content)
	test.AssertEqual(t, fmt.Sprintf(" %sfoo%s bar", MARKER_ITALIC_OPEN, MARKER_ITALIC_CLOSE), tokenizedColumn)
	test.AssertEqual(t, `colspan="2" style="text-align:center;"`, css)
}
