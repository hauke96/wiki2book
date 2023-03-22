package parser

import (
	"fmt"
	"github.com/hauke96/wiki2book/src/test"
	"testing"
)

func TestParseTable_simple(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `{|
| foo || bar
|-
| blubb || moin
|}`
	tokenizedTable := tokenizer.parseTables(content)

	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 6), tokenizedTable)
	test.AssertMapEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 6): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 2, TOKEN_TABLE_ROW, 5),
		// outer table
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 2): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_COL, 0, TOKEN_TABLE_COL, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 0): " foo ",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 1): " bar",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 5): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_COL, 3, TOKEN_TABLE_COL, 4),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 3): " blubb ",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 4): " moin",
	}, tokenizer.getTokenMap())
}

func TestParseTable_complex(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `before
{| class="wikitable"
|+ rowspan="2" style="text-align:left;"| capti0n
foo
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
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL_ATTRIBUTES, 0): "rowspan=\"2\" style=\"text-align:left;\"",
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
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 5): fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 3),
		// inner table
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 3):     fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 2),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 2): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_COL, 0, TOKEN_TABLE_COL, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 0): " inner ",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 1): " table",
	}, tokenizer.getTokenMap())
}

func TestParseTable_withoutExplicitRowStart(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `{| class="wikitable"
|
| foo
|-
| bar
|
|}`
	tokenizedTable := tokenizer.parseTables(content)

	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 6), tokenizedTable)
	test.AssertMapEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 6): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 2, TOKEN_TABLE_ROW, 5),
		// outer table
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 2): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_COL, 0, TOKEN_TABLE_COL, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 0): "",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 1): " foo",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 5): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_COL, 3, TOKEN_TABLE_COL, 4),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 3): " bar",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 4): "",
	}, tokenizer.getTokenMap())
}

func TestParseTable_withEmptyRows(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `{| class="wikitable"
|-
| foo
|-
|-
| bar
|-
|}`
	tokenizedTable := tokenizer.parseTables(content)

	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 4), tokenizedTable)
	test.AssertMapEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 4): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 1, TOKEN_TABLE_ROW, 3),
		// outer table
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 1): fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 0),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 0): " foo",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 3): fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 2),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 2): " bar",
	}, tokenizer.getTokenMap())
}

func TestParseTable_withEmptyColumn(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `{|
| foo || bar
|-
|  || moin
|}`
	tokenizedTable := tokenizer.parseTables(content)

	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 6), tokenizedTable)
	test.AssertMapEqual(t, map[string]string{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 6): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 2, TOKEN_TABLE_ROW, 5),
		// outer table
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 2): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_COL, 0, TOKEN_TABLE_COL, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 0): " foo ",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 1): " bar",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_ROW, 5): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_TABLE_COL, 3, TOKEN_TABLE_COL, 4),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 3): "  ",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE_COL, 4): " moin",
	}, tokenizer.getTokenMap())
}

func TestTokenizeTableRow_withHead(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	lines := []string{
		"! foo",
		"!bar",
		"|-",
	}
	tokenizedColumn, i := tokenizer.tokenizeTableRow(lines, 0)
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
	tokenizedColumn, i := tokenizer.tokenizeTableRow(lines, 0)
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
