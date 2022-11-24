package parser

import (
	"fmt"
	"github.com/hauke96/wiki2book/src/test"
	"strings"
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

func TestParseHeading(t *testing.T) {
	for i := 1; i < 7; i++ {
		tokenizer := NewTokenizer("foo", "bar")
		headingPrefixSuffix := strings.Repeat("=", i)
		content := tokenizer.parseHeadings(fmt.Sprintf("%s h%d %s", headingPrefixSuffix, i, headingPrefixSuffix))
		token := fmt.Sprintf(TOKEN_TEMPLATE, fmt.Sprintf(TOKEN_HEADING_TEMPLATE, i), 0)
		test.AssertEqual(t, token, content)
		test.AssertEqual(t, map[string]string{
			token: fmt.Sprintf("h%d", i),
		}, tokenizer.getTokenMap())
	}
}

func TestParseHeadingWithFormatting(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseHeadings("== H2 ''with formatting'' ==")

	test.AssertEqual(t, "$$TOKEN_HEADING_2_0$$", content)
	test.AssertEqual(t, map[string]string{
		"$$TOKEN_HEADING_2_0$$": "H2 $$MARKER_ITALIC_OPEN$$with formatting$$MARKER_ITALIC_CLOSE$$",
	}, tokenizer.getTokenMap())
}

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

func TestParseGalleries(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseGalleries(`foo
<gallery>file0.jpg
file1.jpg|captiion
</gallery>
bar
<gallery some="parameter">
file2.jpg|test123
file 3.jpg
</gallery>
blubb`)

	test.AssertEqual(t, `foo
[[File:File0.jpg]]
[[File:File1.jpg|captiion]]
bar
[[File:File2.jpg|test123]]
[[File:File_3.jpg]]
blubb`, content)

	test.AssertEqual(t, map[string]string{}, tokenizer.getTokenMap())
}

func TestParseImagemaps(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseImageMaps(`foo
<imagemap>File:picture.jpg
some
stuff
</imagemap>
bar
<imagemap some="parameter">
Image:picture.jpg
some stuff
</imagemap>
blubb`)

	test.AssertEqual(t, `foo
[[File:Picture.jpg]]
bar
[[Image:Picture.jpg]]
blubb`, content)

	test.AssertEqual(t, map[string]string{}, tokenizer.getTokenMap())
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
	content = tokenizer.parseImages("foo [[Datei:image.jpg|100x50px]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE_INLINE+"_2$$ bar", content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("foo [[Datei:image.jpg|101x51px]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_2$$ bar", content)

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
|+  style="text-align:left;"| capti0n
foo
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

	test.AssertEqual(t, fmt.Sprintf("before\n"+TOKEN_TEMPLATE+"\nafter", TOKEN_TABLE, 19), tokenizedTable)
	test.AssertMapEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 19): fmt.Sprintf(
			TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE,
			TOKEN_TABLE_CAPTION, 1, TOKEN_TABLE_ROW, 4, TOKEN_TABLE_ROW, 10, TOKEN_TABLE_ROW, 13, TOKEN_TABLE_ROW, 18,
		),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL_ATTRIBUTES, 0): "style=\"text-align:left;\"",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_CAPTION, 1):        fmt.Sprintf(TOKEN_TEMPLATE+" capti0n\nfoo", TOKEN_TABLE_COL_ATTRIBUTES, 0),
		// row 0: heading
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 4):  fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_HEAD, 2, TOKEN_TABLE_HEAD, 3),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_HEAD, 2): " head1 ",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_HEAD, 3): " " + MARKER_BOLD_OPEN + "head2" + MARKER_BOLD_CLOSE,
		// row 1: internal link
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 10):            fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_COL, 8, TOKEN_TABLE_COL, 9),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 8):             fmt.Sprintf(" foo "+TOKEN_TEMPLATE+" ", TOKEN_INTERNAL_LINK, 7),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_INTERNAL_LINK, 7):         fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_INTERNAL_LINK_ARTICLE, 5, TOKEN_INTERNAL_LINK_TEXT, 6),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_INTERNAL_LINK_ARTICLE, 5): "internal",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_INTERNAL_LINK_TEXT, 6):    "internal",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 9):             " bar",
		// row 2: multi-line
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 13): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_COL, 11, TOKEN_TABLE_COL, 12),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 11): " This row\nis",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 12): " multi-line wikitext",
		// row 3: attributes
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 18):            fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_COL, 15, TOKEN_TABLE_COL, 17),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 15):            fmt.Sprintf(TOKEN_TEMPLATE+" some ", TOKEN_TABLE_COL_ATTRIBUTES, 14),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL_ATTRIBUTES, 14): "colspan=\"42\" style=\"text-align:right;\"",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 17):            fmt.Sprintf(TOKEN_TEMPLATE+" attributes ", TOKEN_TABLE_COL_ATTRIBUTES, 16),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL_ATTRIBUTES, 16): "colspan=\"1\"",
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
	test.AssertMapEqual(t, map[string]string{
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
	test.AssertMapEqual(t, map[string]string{
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
	test.AssertMapEqual(t, map[string]string{
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
	tokenizedColumn, attributeToken := tokenizer.tokenizeTableEntry(content)
	test.AssertEqual(t, fmt.Sprintf(" %sfoo%s bar", MARKER_ITALIC_OPEN, MARKER_ITALIC_CLOSE), tokenizedColumn)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL_ATTRIBUTES, 0), attributeToken)
	test.AssertMapEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL_ATTRIBUTES, 0): `colspan="2" style="text-align:center;"`,
	}, tokenizer.getTokenMap())
}

func TestParseList(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `foo
* a
bar
# b
end
`

	newContent := tokenizer.parseLists(content)

	expectedTokenizedContent := fmt.Sprintf(`foo
%s
bar
%s
end
`, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 1), fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_ORDERED_LIST, 3))
	test.AssertEqual(t, expectedTokenizedContent, newContent)
	test.AssertMapEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 1): fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_ORDERED_LIST, 3):   fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 2),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0):      " a",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 2):      " b",
	}, tokenizer.getTokenMap())
}

func TestBelongsToListPrefix(t *testing.T) {
	test.AssertTrue(t, belongsToListPrefix("* foo", "*"))
	test.AssertTrue(t, belongsToListPrefix("# foo", "#"))
	test.AssertTrue(t, belongsToListPrefix(": foo", ":"))
	test.AssertTrue(t, belongsToListPrefix("; foo", ";"))
	test.AssertTrue(t, belongsToListPrefix(": foo", ";"))

	test.AssertFalse(t, belongsToListPrefix("; foo", ":"))
	test.AssertFalse(t, belongsToListPrefix("# foo", "*"))
}

func TestRemoveListPrefix(t *testing.T) {
	test.AssertEqual(t, " foo", removeListPrefix("* foo", "*"))
	test.AssertEqual(t, " foo", removeListPrefix("# foo", "#"))
	test.AssertEqual(t, " foo", removeListPrefix(": foo", ":"))
	test.AssertEqual(t, " foo", removeListPrefix("; foo", ";"))
	test.AssertEqual(t, " foo", removeListPrefix(": foo", ";"))

	test.AssertEqual(t, "* foo", removeListPrefix("* foo", "#"))
	test.AssertEqual(t, "; foo", removeListPrefix("; foo", ":"))
}

func TestParseList_withoutListInContent(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `some
text
without list
`

	newContent := tokenizer.parseLists(content)

	test.AssertEqual(t, content, newContent)
}

func TestTokenizeList(t *testing.T) {
	content := `stuff before list
* foo
* bar
end of list
`
	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, "*", TOKEN_UNORDERED_LIST)
	test.AssertEqual(t, 3, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 2), token)
	test.AssertMapEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 2): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0, TOKEN_LIST_ITEM, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0):      " foo",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 1):      " bar",
	}, tokenizer.getTokenMap())
}

func TestTokenizeList_twoTypesBelowEachOther(t *testing.T) {
	content := `stuff before list
* a
* A
# b
# B
end of list
`
	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, "*", TOKEN_UNORDERED_LIST)
	test.AssertEqual(t, 3, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 2), token)
	test.AssertMapEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 2): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0, TOKEN_LIST_ITEM, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0):      " a",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 1):      " A",
	}, tokenizer.getTokenMap())
}

func TestTokenizeList_withSubList(t *testing.T) {
	content := `stuff before list
* foo
** b''a''r
* bar
end of list
`
	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, "*", TOKEN_UNORDERED_LIST)
	test.AssertEqual(t, 4, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 4), token)
	test.AssertMapEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 4): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 2, TOKEN_LIST_ITEM, 3),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 1): fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 2):      fmt.Sprintf(" foo "+TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0):      fmt.Sprintf(" b%sa%sr", MARKER_ITALIC_OPEN, MARKER_ITALIC_CLOSE),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 3):      " bar",
	}, tokenizer.getTokenMap())
}

func TestTokenizeList_higherLevelStart(t *testing.T) {
	content := `stuff before list
*** foo
*** bar
end of list
`
	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, "*", TOKEN_UNORDERED_LIST)
	test.AssertEqual(t, 3, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 4), token)
	test.AssertMapEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 4): fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 3),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 3): fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 2),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 2): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0, TOKEN_LIST_ITEM, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0):      " foo",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 1):      " bar",
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList(t *testing.T) {
	content := `bla
; foo
: bar
blubb`

	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ";", TOKEN_DESCRIPTION_LIST)

	test.AssertEqual(t, 3, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2), token)
	test.AssertMapEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2):      fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 0, TOKEN_DESCRIPTION_LIST_ITEM, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 0): " foo",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 1): " bar",
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList_headingIsSecondItem(t *testing.T) {
	content := `bla
: foo
; bar
blubb`

	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ";", TOKEN_DESCRIPTION_LIST)

	test.AssertEqual(t, 3, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2), token)
	test.AssertMapEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2):      fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 0, TOKEN_DESCRIPTION_LIST_HEAD, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 0): " foo",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 1): " bar",
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList_withoutHeading(t *testing.T) {
	content := `bla
: foo
: bar
blubb`

	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ";", TOKEN_DESCRIPTION_LIST)

	test.AssertEqual(t, 3, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2), token)
	test.AssertMapEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2):      fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 0, TOKEN_DESCRIPTION_LIST_ITEM, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 0): " foo",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 1): " bar",
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList_deepBeginning(t *testing.T) {
	content := `bla
::: foo
::: bar
blubb`

	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ":", TOKEN_DESCRIPTION_LIST)

	test.AssertEqual(t, 3, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 4), token)
	test.AssertMapEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 4):      fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 3),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 3):      fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2):      fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 0, TOKEN_DESCRIPTION_LIST_ITEM, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 0): " foo",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 1): " bar",
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList_withEmptyItem(t *testing.T) {
	content := `foo
; list
:
: bar
blubb`

	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ";", TOKEN_DESCRIPTION_LIST)

	test.AssertEqual(t, 4, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 3), token)
	test.AssertMapEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 3):      fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 0, TOKEN_DESCRIPTION_LIST_ITEM, 1, TOKEN_DESCRIPTION_LIST_ITEM, 2),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 0): " list",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 1): "",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 2): " bar",
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList_withOtherSubList(t *testing.T) {
	content := `foo
; list
: foo1
:* bar1
:* bar2
: foo2
blubb`

	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ";", TOKEN_DESCRIPTION_LIST)

	test.AssertEqual(t, 6, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 9), token)
	test.AssertMapEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 9):      fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 0, TOKEN_DESCRIPTION_LIST_ITEM, 1, TOKEN_DESCRIPTION_LIST_ITEM, 4, TOKEN_DESCRIPTION_LIST_ITEM, 7, TOKEN_DESCRIPTION_LIST_ITEM, 8),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 0): " list",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 1): " foo1",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 4): fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 3),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 3):        fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 2),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 2):             " bar1",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 7): fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 6),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 6):        fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 5),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 5):             " bar2",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 8): " foo2",
	}, tokenizer.getTokenMap())
}

func TestTokenizeList_withDescriptionSubList(t *testing.T) {
	content := `stuff before list
* foo
*; descr
*: descr-item
* bar
end of list
`
	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, "*", TOKEN_UNORDERED_LIST)
	test.AssertEqual(t, 5, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 5), token)
	test.AssertMapEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 5):        fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 3, TOKEN_LIST_ITEM, 4),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 3):             fmt.Sprintf(" foo "+TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 4):             " bar",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2):      fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 0, TOKEN_DESCRIPTION_LIST_ITEM, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 0): " descr",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 1): " descr-item",
	}, tokenizer.getTokenMap())
}

func TestGetListTokenString(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	test.AssertEqual(t, TOKEN_UNORDERED_LIST, tokenizer.getListTokenString("*"))
	test.AssertEqual(t, TOKEN_ORDERED_LIST, tokenizer.getListTokenString("#"))
	test.AssertEqual(t, TOKEN_DESCRIPTION_LIST, tokenizer.getListTokenString(";"))
	test.AssertEqual(t, "UNKNOWN_LIST_TYPE_-", tokenizer.getListTokenString("-"))
	test.AssertEqual(t, "UNKNOWN_LIST_TYPE_~", tokenizer.getListTokenString("~"))
	test.AssertEqual(t, "UNKNOWN_LIST_TYPE_ ", tokenizer.getListTokenString(" "))
	test.AssertEqual(t, "UNKNOWN_LIST_TYPE_", tokenizer.getListTokenString(""))
}

func TestGetListItemToken(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	test.AssertEqual(t, TOKEN_LIST_ITEM, tokenizer.getListItemTokenString("*"))
	test.AssertEqual(t, TOKEN_LIST_ITEM, tokenizer.getListItemTokenString("#"))
	test.AssertEqual(t, TOKEN_DESCRIPTION_LIST_HEAD, tokenizer.getListItemTokenString(";"))
	test.AssertEqual(t, TOKEN_DESCRIPTION_LIST_ITEM, tokenizer.getListItemTokenString(":"))
	test.AssertEqual(t, "UNKNOWN_LIST_ITEM_TYPE_-", tokenizer.getListItemTokenString("-"))
	test.AssertEqual(t, "UNKNOWN_LIST_ITEM_TYPE_~", tokenizer.getListItemTokenString("~"))
	test.AssertEqual(t, "UNKNOWN_LIST_ITEM_TYPE_ ", tokenizer.getListItemTokenString(" "))
	test.AssertEqual(t, "UNKNOWN_LIST_ITEM_TYPE_", tokenizer.getListItemTokenString(""))
}

func TestGetReferenceHeadAndFoot(t *testing.T) {
	head := "some text<ref>foo</ref> with refs<ref name=\"barbar\">bar</ref>.\n"
	foot := "foooooooter"
	content := head + "<references />\n" + foot

	tokenizer := NewTokenizer("foo", "bar")
	newHead, newFoot, newContent, noRefListFound := tokenizer.getReferenceHeadAndFoot(content)

	test.AssertEqual(t, head, newHead)
	test.AssertEqual(t, foot, newFoot)
	test.AssertEqual(t, content, newContent)
	test.AssertFalse(t, noRefListFound)
}

func TestGetSortedReferenceNames(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	refIndexToName := map[int]string{
		3: "bar",
		1: "foo",
		2: "blubb",
	}

	sortedName, refNameToIndex := tokenizer.getSortedReferenceNames(refIndexToName)

	test.AssertEqual(t, []string{"foo", "blubb", "bar"}, sortedName)
	test.AssertEqual(t, map[string]int{
		"foo":   1,
		"blubb": 2,
		"bar":   3,
	}, refNameToIndex)
}

func TestReplaceNamedReferences(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	head := `some<ref>foo</ref> text with refs<ref name="barbar">bar</ref> some other <ref>foo</ref>`
	content := head + ` and even more<ref>blubb</ref>text`
	referenceDefinitions := map[string]string{}
	newHead := tokenizer.replaceNamedReferences(content, referenceDefinitions, head)

	test.AssertMapEqual(t, map[string]string{
		"barbar": `<ref name="barbar">bar</ref>`,
	}, referenceDefinitions)
	test.AssertEqual(t, `some<ref>foo</ref> text with refs<ref name="barbar" /> some other <ref>foo</ref>`, newHead)
}

func TestReplaceUnnamedReferences(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	head := `some<ref some="param">foo</ref> text with refs<ref name="barbar">bar</ref>`
	content := head + ` and even more<ref>blubb</ref>text`
	referenceDefinitions := map[string]string{}
	newHead := tokenizer.replaceUnnamedReferences(content, referenceDefinitions, head)

	test.AssertMapEqual(t, map[string]string{
		"2ae457b665ef5955b2fc685cdaaa879c96c14801": `<ref some="param">foo</ref>`,
		"2839c654c615b6833625f76f38b609c71b74ada4": "<ref>blubb</ref>",
	}, referenceDefinitions)
	test.AssertEqual(t, `some<ref name="2ae457b665ef5955b2fc685cdaaa879c96c14801" /> text with refs<ref name="barbar">bar</ref>`, newHead)
}

func TestGetReferenceUsages(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `some<ref name="bar" /> text with refs<ref name="foo" />`
	usages, _ := tokenizer.getReferenceUsages(content)

	test.AssertEqual(t, map[string]string{
		"bar": `<ref name="bar" />`,
		"foo": `<ref name="foo" />`,
	}, usages)
}

func TestParseReferences(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `some text<ref>bar</ref>
some<ref name="blubb">blubbeldy</ref> other<ref name="fooref" /> text
<references responsive>
<ref name="fooref">foo</ref>
</references>
some footer`
	expectedContent := "some text" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 0) + "\n" +
		"some" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 1) + " other" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 2) + " text\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 3) + "\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 4) + "\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 5) + "\n" +
		"some footer"

	newContent := tokenizer.parseReferences(content)

	test.AssertEqual(t, expectedContent, newContent)
}

func TestParseMath(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `abc<math>x \cdot y</math>def
some
<math>
\multiline{math}
</math>
end`
	tokenizedContent := tokenizer.tokenizeContent(&tokenizer, content)

	test.AssertEqual(t, "abc"+fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_MATH, 0)+"def\nsome\n"+fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_MATH, 1)+"\nend", tokenizedContent)
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
