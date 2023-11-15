package parser

import (
	"fmt"
	"testing"
	"wiki2book/test"
)

func TestParseTable_simple(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `{|
| foo || bar
|-
| blubb || moin
|}`
	tokenizedTable := tokenizer.parseTables(content)

	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 0), tokenizedTable)
	expectedTableToken := TableToken{
		Rows: []TableRowToken{
			TableRowToken{
				Columns: []TableColToken{
					{
						Attributes: TableColAttributeToken{},
						Content:    "foo",
					},
					{
						Attributes: TableColAttributeToken{},
						Content:    "bar",
					},
				},
			},
			TableRowToken{
				Columns: []TableColToken{
					{
						Attributes: TableColAttributeToken{},
						Content:    "blubb",
					},
					{
						Attributes: TableColAttributeToken{},
						Content:    "moin",
					},
				},
			},
		},
	}
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 0): expectedTableToken,
	}, tokenizer.getTokenMap())
}

func TestParseTable_withIndentation(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `  {|
  ! h1
   ! h2
  |-
  | foo 
   | bar
  |-
  | blubb || moin
  |-
  |+ 
  caption
  |}`
	tokenizedTable := tokenizer.parseTables(content)

	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 0), tokenizedTable)
	expectedTableToken := TableToken{
		Caption: TableCaptionToken{
			Content: "caption",
		},
		Rows: []TableRowToken{
			{
				Columns: []TableColToken{
					{
						Attributes: TableColAttributeToken{},
						Content:    "h1",
						IsHeading:  true,
					},
					{
						Attributes: TableColAttributeToken{},
						Content:    "h2",
						IsHeading:  true,
					},
				},
			},
			{
				Columns: []TableColToken{
					{
						Attributes: TableColAttributeToken{},
						Content:    "foo",
					},
					{
						Attributes: TableColAttributeToken{},
						Content:    "bar",
					},
				},
			},
			{
				Columns: []TableColToken{
					{
						Attributes: TableColAttributeToken{},
						Content:    "blubb",
					},
					{
						Attributes: TableColAttributeToken{},
						Content:    "moin",
					},
				},
			},
		},
	}
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 0): expectedTableToken,
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

	test.AssertEqual(t, fmt.Sprintf("before\n"+TOKEN_TEMPLATE+"\nafter", TOKEN_TABLE, 1), tokenizedTable)
	expectedTableToken := TableToken{
		Caption: TableCaptionToken{
			Content: "capti0n\nfoo",
		},
		Rows: []TableRowToken{
			{
				Columns: []TableColToken{
					{
						Attributes: TableColAttributeToken{},
						Content:    "head1",
						IsHeading:  true,
					},
					{
						Attributes: TableColAttributeToken{},
						Content:    MARKER_BOLD_OPEN + "head2" + MARKER_BOLD_CLOSE,
						IsHeading:  true,
					},
				},
			},
			{
				Columns: []TableColToken{
					{
						Attributes: TableColAttributeToken{},
						Content:    "foo " + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_INTERNAL_LINK, 0),
					},
					{
						Attributes: TableColAttributeToken{},
						Content:    "bar",
					},
				},
			},
			TableRowToken{
				Columns: []TableColToken{
					{
						Attributes: TableColAttributeToken{},
						Content:    "This row\nis",
					},
					{
						Attributes: TableColAttributeToken{},
						Content:    "multi-line wikitext",
					},
				},
			},
			TableRowToken{
				Columns: []TableColToken{
					{
						Attributes: TableColAttributeToken{
							Attributes: []string{
								`colspan="42"`,
								`style="text-align:right;"`,
							},
						},
						Content: "some",
					},
					{
						Attributes: TableColAttributeToken{
							Attributes: []string{
								`colspan="1"`,
							},
						},
						Content: "attributes",
					},
				},
			},
		},
	}
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 1): expectedTableToken,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_INTERNAL_LINK, 0): InternalLinkToken{
			ArticleName: "internal",
			LinkText:    "internal",
		},
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

	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 1), tokenizedTable)

	expectedInnerTableToken := TableToken{
		Rows: []TableRowToken{
			TableRowToken{
				Columns: []TableColToken{
					{
						Attributes: TableColAttributeToken{},
						Content:    "inner",
					},
					{
						Attributes: TableColAttributeToken{},
						Content:    "table",
					},
				},
			},
		},
	}
	expectedOuterTableToken := TableToken{
		Rows: []TableRowToken{
			TableRowToken{
				Columns: []TableColToken{
					{
						Attributes: TableColAttributeToken{},
						Content:    "foo",
					},
					{
						Attributes: TableColAttributeToken{},
						Content:    fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 0),
					},
				},
			},
		},
	}
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 0): expectedInnerTableToken,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 1): expectedOuterTableToken,
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

	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 0), tokenizedTable)
	expectedTableToken := TableToken{
		Rows: []TableRowToken{
			TableRowToken{
				Columns: []TableColToken{
					{
						Attributes: TableColAttributeToken{},
						Content:    "",
					},
					{
						Attributes: TableColAttributeToken{},
						Content:    "foo",
					},
				},
			},
			TableRowToken{
				Columns: []TableColToken{
					{
						Attributes: TableColAttributeToken{},
						Content:    "bar",
					},
					{
						Attributes: TableColAttributeToken{},
						Content:    "",
					},
				},
			},
		},
	}
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 0): expectedTableToken,
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

	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 0), tokenizedTable)
	expectedTableToken := TableToken{
		Rows: []TableRowToken{
			TableRowToken{
				Columns: []TableColToken{
					{
						Attributes: TableColAttributeToken{},
						Content:    "foo",
					},
				},
			},
			TableRowToken{},
			TableRowToken{
				Columns: []TableColToken{
					{
						Attributes: TableColAttributeToken{},
						Content:    "bar",
					},
				},
			},
			TableRowToken{},
		},
	}
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 0): expectedTableToken,
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

	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 0), tokenizedTable)
	expectedTableToken := TableToken{
		Rows: []TableRowToken{
			TableRowToken{
				Columns: []TableColToken{
					{
						Attributes: TableColAttributeToken{},
						Content:    "foo",
					},
					{
						Attributes: TableColAttributeToken{},
						Content:    "bar",
					},
				},
			},
			TableRowToken{
				Columns: []TableColToken{
					{
						Attributes: TableColAttributeToken{},
						Content:    "",
					},
					{
						Attributes: TableColAttributeToken{},
						Content:    "moin",
					},
				},
			},
		},
	}
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 0): expectedTableToken,
	}, tokenizer.getTokenMap())
}

func TestParseTable_captionInsideRow(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `{|
|-
|+ cap
| foo
|}`
	tokenizedTable := tokenizer.parseTables(content)

	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 0), tokenizedTable)
	expectedTableToken := TableToken{
		Caption: TableCaptionToken{
			Content: "cap",
		},
		Rows: []TableRowToken{
			{},
			{
				Columns: []TableColToken{
					{
						Attributes: TableColAttributeToken{},
						Content:    "foo",
					},
				},
			},
		},
	}
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_TABLE, 0): expectedTableToken,
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

	expectedRowToken := TableRowToken{
		Columns: []TableColToken{
			{
				Attributes: TableColAttributeToken{},
				Content:    "foo",
				IsHeading:  true,
			},
			{
				Attributes: TableColAttributeToken{},
				Content:    "bar",
				IsHeading:  true,
			},
		},
	}
	test.AssertEqual(t, expectedRowToken, tokenizedColumn)
	test.AssertEqual(t, 1, i)
	test.AssertMapEqual(t, map[string]interface{}{}, tokenizer.getTokenMap())
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

	expectedRowToken := TableRowToken{
		Columns: []TableColToken{
			{
				Attributes: TableColAttributeToken{},
				Content:    "foo",
			},
			{
				Attributes: TableColAttributeToken{},
				Content:    "bar",
			},
			{
				Attributes: TableColAttributeToken{
					Attributes: []string{`colspan="2"`},
				},
				Content: "abc\ndef",
			},
		},
	}

	test.AssertEqual(t, expectedRowToken, tokenizedColumn)
	test.AssertEqual(t, 3, i)
	test.AssertMapEqual(t, map[string]interface{}{}, tokenizer.getTokenMap())
}

func TestTokenizeTableColumn(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `colspan="2" style="text-align:center; background:Lightgray;" | ''foo'' bar`

	tokenizedColumn, attributeToken := tokenizer.tokenizeTableEntry(content)

	expectedAttributeToken := TableColAttributeToken{
		Attributes: []string{
			`colspan="2"`,
			`style="text-align:center;"`,
		},
	}
	test.AssertEqual(t, expectedAttributeToken, attributeToken)
	test.AssertEqual(t, fmt.Sprintf(" %sfoo%s bar", MARKER_ITALIC_OPEN, MARKER_ITALIC_CLOSE), tokenizedColumn)
	test.AssertMapEqual(t, map[string]interface{}{}, tokenizer.getTokenMap())
}
