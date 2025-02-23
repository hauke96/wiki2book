package html

import (
	"fmt"
	"testing"
	"wiki2book/parser"
	"wiki2book/test"
)

var generator = HtmlGenerator{}

func TestExpandMarker(t *testing.T) {

	test.AssertEqual(t, "<b>", generator.expandMarker(parser.MARKER_BOLD_OPEN))
	test.AssertEqual(t, "</b>", generator.expandMarker(parser.MARKER_BOLD_CLOSE))
	test.AssertEqual(t, "<i>", generator.expandMarker(parser.MARKER_ITALIC_OPEN))
	test.AssertEqual(t, "</i>", generator.expandMarker(parser.MARKER_ITALIC_CLOSE))
	test.AssertEqual(t, "<br><br>", generator.expandMarker(parser.MARKER_PARAGRAPH))

	test.AssertEqual(t, "<b><i><br><br></i></b>", generator.expandMarker(parser.MARKER_BOLD_OPEN+parser.MARKER_ITALIC_OPEN+parser.MARKER_PARAGRAPH+parser.MARKER_ITALIC_CLOSE+parser.MARKER_BOLD_CLOSE))
}

func TestExpandHeadings(t *testing.T) {
	tokenKey := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_HEADING, 3)
	token := parser.HeadingToken{
		Content: "foobar",
		Depth:   3,
	}
	tokenMap := map[string]parser.Token{
		tokenKey: token,
	}
	generator.TokenMap = tokenMap

	headings, err := generator.expand(token)
	test.AssertNil(t, err)
	test.AssertEqual(t, fmt.Sprintf("<h%d>foobar</h%d>", 3, 3), headings)
}

func TestExpandImage(t *testing.T) {
	result := `<div class="figure">
<img alt="image" src="./foo/image.jpg" style="vertical-align: middle; width: 10px; height: 20px;">
<div class="caption">
some <b>caption</b>
</div>
</div>`
	tokenImage := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE, 1)
	token := parser.ImageToken{
		Filename: "foo/image.jpg",
		Caption:  parser.CaptionToken{Content: "some " + parser.MARKER_BOLD_OPEN + "caption" + parser.MARKER_BOLD_CLOSE},
		SizeX:    10,
		SizeY:    20,
	}
	tokenMap := map[string]parser.Token{
		tokenImage: token,
	}
	generator.TokenMap = tokenMap

	actualResult, err := generator.expand(token)
	test.AssertNil(t, err)
	test.AssertEqual(t, result, actualResult)
}

func TestExpandImage_encodeSpecialCharacters(t *testing.T) {
	result := `<div class="figure">
<img alt="image" src="./foo/%22some%27special%25chars.jpg" style="vertical-align: middle; width: 10px; height: 20px;">
<div class="caption">
some <b>caption</b>
</div>
</div>`
	tokenImage := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE, 1)
	token := parser.ImageToken{
		Filename: "foo/\"some'special%chars.jpg",
		Caption:  parser.CaptionToken{Content: "some " + parser.MARKER_BOLD_OPEN + "caption" + parser.MARKER_BOLD_CLOSE},
		SizeX:    10,
		SizeY:    20,
	}
	tokenMap := map[string]parser.Token{
		tokenImage: token,
	}
	generator.TokenMap = tokenMap

	actualResult, err := generator.expand(token)
	test.AssertNil(t, err)
	test.AssertEqual(t, result, actualResult)
}

func TestExpandImage_noCaption(t *testing.T) {
	result := `<div class="figure">
<img alt="image" src="./foo/image.jpg" style="vertical-align: middle; width: 10px; height: 20px;">
<div class="caption">

</div>
</div>`
	tokenImage := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE, 0)
	token := parser.ImageToken{
		Filename: "foo/image.jpg",
		Caption:  parser.CaptionToken{},
		SizeX:    10,
		SizeY:    20,
	}
	tokenMap := map[string]parser.Token{
		tokenImage: token,
	}
	generator.TokenMap = tokenMap

	actualResult, err := generator.expand(token)
	test.AssertNil(t, err)
	test.AssertEqual(t, result, actualResult)
}

func TestExpandImage_onlyOneSizeSpecified(t *testing.T) {
	// Only width
	result := `<div class="figure">
<img alt="image" src="./foo/image.jpg" style="vertical-align: middle; width: 10px; height: auto;">
<div class="caption">
some <b>caption</b>
</div>
</div>`
	tokenImage := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE, 1)
	token := parser.ImageToken{
		Filename: "foo/image.jpg",
		Caption:  parser.CaptionToken{Content: "some " + parser.MARKER_BOLD_OPEN + "caption" + parser.MARKER_BOLD_CLOSE},
		SizeX:    10,
		SizeY:    -1,
	}
	tokenMap := map[string]parser.Token{
		tokenImage: token,
	}
	generator.TokenMap = tokenMap

	actualResult, err := generator.expand(token)
	test.AssertNil(t, err)
	test.AssertEqual(t, result, actualResult)

	// Only height
	result = `<div class="figure">
<img alt="image" src="./foo/image.jpg" style="vertical-align: middle; width: auto; height: 10px;">
<div class="caption">
some <b>caption</b>
</div>
</div>`
	tokenImage = fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE, 1)
	token = parser.ImageToken{
		Filename: "foo/image.jpg",
		Caption:  parser.CaptionToken{Content: "some " + parser.MARKER_BOLD_OPEN + "caption" + parser.MARKER_BOLD_CLOSE},
		SizeX:    -1,
		SizeY:    10,
	}
	tokenMap = map[string]parser.Token{
		tokenImage: token,
	}
	generator.TokenMap = tokenMap

	actualResult, err = generator.expand(token)
	test.AssertNil(t, err)
	test.AssertEqual(t, result, actualResult)
}

func TestExpandImageInline(t *testing.T) {
	result := `<img alt="image" class="inline" src="./foo/image.jpg" style="vertical-align: middle; width: 10px; height: 20px;">`
	tokenImage := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE_INLINE, 1)
	token := parser.InlineImageToken{
		Filename: "foo/image.jpg",
		SizeX:    10,
		SizeY:    20,
	}
	tokenMap := map[string]parser.Token{
		tokenImage: token,
	}
	generator.TokenMap = tokenMap

	actualResult, err := generator.expand(token)
	test.AssertNil(t, err)
	test.AssertEqual(t, result, actualResult)
}

func TestExpandImageInline_encodeSpecialCharacters(t *testing.T) {
	result := `<img alt="image" class="inline" src="./foo/%22some%27special%25chars.jpg" style="vertical-align: middle; width: 10px; height: 20px;">`
	tokenImage := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE_INLINE, 1)
	token := parser.InlineImageToken{
		Filename: "foo/\"some'special%chars.jpg",
		SizeX:    10,
		SizeY:    20,
	}
	tokenMap := map[string]parser.Token{
		tokenImage: token,
	}
	generator.TokenMap = tokenMap

	actualResult, err := generator.expand(token)
	test.AssertNil(t, err)
	test.AssertEqual(t, result, actualResult)
}

func TestExpandInternalLink(t *testing.T) {
	tokenLink := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_INTERNAL_LINK, 0)
	tokenMap := map[string]parser.Token{
		tokenLink: parser.InternalLinkToken{
			ArticleName: "foo",
			LinkText:    "b" + parser.MARKER_BOLD_OPEN + "a" + parser.MARKER_BOLD_CLOSE + "r",
		},
	}
	generator.TokenMap = tokenMap

	link, err := generator.expand(tokenLink)
	test.AssertNil(t, err)
	test.AssertEqual(t, "b<b>a</b>r", link)
}

func TestExpandExternalLink(t *testing.T) {
	tokenLink := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_EXTERNAL_LINK, 0)
	url := "https://foo.com"
	tokenMap := map[string]parser.Token{
		tokenLink: parser.ExternalLinkToken{
			URL:      url,
			LinkText: "b" + parser.MARKER_BOLD_OPEN + "a" + parser.MARKER_BOLD_CLOSE + "r",
		},
	}
	generator.TokenMap = tokenMap

	link, err := generator.expand(tokenLink)
	test.AssertNil(t, err)
	test.AssertEqual(t, "<a href=\""+url+"\">b<b>a</b>r</a>", link)
}

func TestExpandTable(t *testing.T) {
	tokenTable := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_TABLE, 0)
	tokenMap := map[string]parser.Token{
		tokenTable: parser.TableToken{
			Caption: parser.TableCaptionToken{
				Content: "caption",
			},
			Rows: []parser.TableRowToken{
				{
					Columns: []parser.TableColToken{
						{
							Attributes: parser.TableColAttributeToken{},
							Content:    "b" + parser.MARKER_BOLD_OPEN + "a" + parser.MARKER_BOLD_CLOSE + "r",
							IsHeading:  false,
						},
					},
				},
			},
		},
	}
	generator.TokenMap = tokenMap

	row, err := generator.expand(tokenTable)
	test.AssertNil(t, err)
	test.AssertEqual(t, `<div class="figure">
<table>
<tr>
<td>
b<b>a</b>r
</td>
</tr>
</table>
<div class="caption">
caption
</div>
</div>`, row)
}

func TestExpandTableRow(t *testing.T) {
	tokenRow := parser.TableRowToken{
		Columns: []parser.TableColToken{
			{
				Attributes: parser.TableColAttributeToken{},
				Content:    "b" + parser.MARKER_BOLD_OPEN + "a" + parser.MARKER_BOLD_CLOSE + "r",
				IsHeading:  false,
			},
		},
	}

	row, err := generator.expandTableRow(tokenRow)
	test.AssertNil(t, err)
	test.AssertEqual(t, "<tr>\n<td>\nb<b>a</b>r\n</td>\n</tr>", row)
}

func TestExpandTableColumn(t *testing.T) {
	tokenCol := parser.TableColToken{
		Attributes: parser.TableColAttributeToken{},
		Content:    "b" + parser.MARKER_BOLD_OPEN + "a" + parser.MARKER_BOLD_CLOSE + "r",
		IsHeading:  false,
	}

	row, err := generator.expandTableColumn(tokenCol)
	test.AssertNil(t, err)
	test.AssertEqual(t, "<td>\nb<b>a</b>r\n</td>", row)
}

func TestExpandTableColumnWithAttributes(t *testing.T) {
	tokenCol := parser.TableColToken{
		Attributes: parser.TableColAttributeToken{
			Attributes: []string{"style=\"width: infinity lol\""},
		},
		Content:   "b" + parser.MARKER_BOLD_OPEN + "a" + parser.MARKER_BOLD_CLOSE + "r",
		IsHeading: false,
	}

	row, err := generator.expandTableColumn(tokenCol)
	test.AssertNil(t, err)
	test.AssertEqual(t, "<td style=\"width: infinity lol\">\nb<b>a</b>r\n</td>", row)
}

func TestExpandUnorderedList(t *testing.T) {
	item1 := parser.ListItemToken{Type: parser.NORMAL_ITEM, Content: "foo"}
	item2 := parser.ListItemToken{Type: parser.NORMAL_ITEM, Content: fmt.Sprintf("b%sa%sr", parser.MARKER_BOLD_OPEN, parser.MARKER_BOLD_CLOSE)}
	list3 := parser.UnorderedListToken{Items: []parser.ListItemToken{item1, item2}}

	row, err := generator.expand(list3)
	test.AssertNil(t, err)
	test.AssertEqual(t, `<ul>
<li>
foo
</li>
<li>
b<b>a</b>r
</li>
</ul>`, row)
}

func TestExpandOrderedList(t *testing.T) {
	item1 := parser.ListItemToken{Type: parser.NORMAL_ITEM, Content: "foo"}
	item2 := parser.ListItemToken{Type: parser.NORMAL_ITEM, Content: fmt.Sprintf("b%sa%sr", parser.MARKER_BOLD_OPEN, parser.MARKER_BOLD_CLOSE)}
	list3 := parser.OrderedListToken{Items: []parser.ListItemToken{item1, item2}}

	row, err := generator.expand(list3)
	test.AssertNil(t, err)
	test.AssertEqual(t, `<ol>
<li>
foo
</li>
<li>
b<b>a</b>r
</li>
</ol>`, row)
}

func TestExpandOrderedList_specifyNumberOfItems(t *testing.T) {
	item1 := parser.ListItemToken{Type: parser.NORMAL_ITEM, Content: "foo"}
	item2 := parser.ListItemToken{Type: parser.NORMAL_ITEM, Content: "<li value=42> bar"}
	item3 := parser.ListItemToken{Type: parser.NORMAL_ITEM, Content: "<li value=10> another item with custom value   </li>"}
	item4 := parser.ListItemToken{Type: parser.NORMAL_ITEM, Content: fmt.Sprintf("b%sa%sr", parser.MARKER_BOLD_OPEN, parser.MARKER_BOLD_CLOSE)}
	list3 := parser.OrderedListToken{Items: []parser.ListItemToken{item1, item2, item3, item4}}

	row, err := generator.expand(list3)
	test.AssertNil(t, err)
	test.AssertEqual(t, `<ol>
<li>
foo
</li>
<li value=42> bar</li>
<li value=10> another item with custom value   </li>
<li>
b<b>a</b>r
</li>
</ol>`, row)
}

func TestExpandDescriptionList(t *testing.T) {
	item1 := parser.ListItemToken{Type: parser.DESCRIPTION_HEAD, Content: "foo"}
	item2 := parser.ListItemToken{Type: parser.DESCRIPTION_ITEM, Content: fmt.Sprintf("b%sa%sr", parser.MARKER_BOLD_OPEN, parser.MARKER_BOLD_CLOSE)}
	list3 := parser.DescriptionListToken{Items: []parser.ListItemToken{item1, item2}}

	row, err := generator.expand(list3)
	test.AssertNil(t, err)
	test.AssertEqual(t, `<div class="description-list">
<div class="dt">
foo
</div>
<div class="dd">
b<b>a</b>r
</div>
</div>`, row)
}

func TestExpandNestedLists(t *testing.T) {
	tokenList := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_UNORDERED_LIST, 2)

	itemInner := parser.ListItemToken{Type: parser.NORMAL_ITEM, Content: "bar"}
	listInner := parser.OrderedListToken{Items: []parser.ListItemToken{itemInner}}
	itemOuter := parser.ListItemToken{Type: parser.NORMAL_ITEM, Content: "foo", SubLists: []parser.ListToken{listInner}}
	listOuter := parser.UnorderedListToken{Items: []parser.ListItemToken{itemOuter}}
	tokenMap := map[string]parser.Token{
		tokenList: tokenList,
	}
	generator.TokenMap = tokenMap

	row, err := generator.expand(listOuter)
	test.AssertNil(t, err)
	test.AssertEqual(t, `<ul>
<li>
foo
<ol>
<li>
bar
</li>
</ol>
</li>
</ul>`, row)
}

func TestExpandRefDefinition(t *testing.T) {
	tokenKey := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_REF_DEF, 2)
	tokenMap := map[string]parser.Token{
		tokenKey: parser.RefDefinitionToken{
			Index:   42,
			Content: "f" + parser.MARKER_BOLD_OPEN + "o" + parser.MARKER_BOLD_CLOSE + "o",
		},
	}
	generator.TokenMap = tokenMap

	row, err := generator.expand(tokenKey)

	test.AssertNil(t, err)
	test.AssertEqual(t, `[43] f<b>o</b>o<br>`, row)
}

func TestExpandRefUsage(t *testing.T) {
	tokenKey := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_REF_DEF, 2)
	tokenMap := map[string]parser.Token{
		tokenKey: parser.RefUsageToken{
			Index: 42,
		},
	}
	generator.TokenMap = tokenMap

	row, err := generator.expand(tokenKey)

	test.AssertNil(t, err)
	test.AssertEqual(t, `[43]`, row)
}

func TestExpandNowiki(t *testing.T) {
	tokenKey := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_NOWIKI, 0)
	tokenMap := map[string]parser.Token{
		tokenKey: parser.NowikiToken{
			Content: "something",
		},
	}
	generator.TokenMap = tokenMap

	row, err := generator.expand(tokenKey)

	test.AssertNil(t, err)
	test.AssertEqual(t, "something", row)
}
