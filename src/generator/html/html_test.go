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
	token := &parser.HeadingToken{
		Content: "foobar",
		Depth:   3,
	}
	tokenMap := map[string]interface{}{
		tokenKey: token,
	}

	headings, err := generator.expand("some\n"+tokenKey+"\nheading", tokenMap)
	test.AssertNil(t, err)
	test.AssertEqual(t, fmt.Sprintf("some\n<h%d>foobar</h%d>\nheading", 3, 3), headings)
}

func TestExpandImage(t *testing.T) {
	result := `<div class="figure">
<img alt="image" src="./foo/image.jpg" style="vertical-align: middle; width: 10px; height: 20px;">
<div class="caption">
some <b>caption</b>
</div>
</div>`
	tokenImage := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE, 1)
	tokenCaption := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE_CAPTION, 0)
	token := &parser.ImageToken{
		Filename:        "foo/image.jpg",
		CaptionTokenKey: tokenCaption,
		SizeX:           10,
		SizeY:           20,
	}
	tokenMap := map[string]interface{}{
		tokenImage:   token,
		tokenCaption: "some " + parser.MARKER_BOLD_OPEN + "caption" + parser.MARKER_BOLD_CLOSE,
	}

	actualResult, err := generator.expand(token, tokenMap)
	test.AssertNil(t, err)
	test.AssertEqual(t, result, actualResult)

	actualResult, err = generator.expand("$$TOKEN_"+parser.TOKEN_IMAGE+"_23852376$$", tokenMap)
	test.AssertNotNil(t, err)
	test.AssertEqual(t, "", actualResult)
}

func TestExpandImage_noCaption(t *testing.T) {
	result := `<div class="figure">
<img alt="image" src="./foo/image.jpg" style="vertical-align: middle; width: 10px; height: 20px;">
<div class="caption">

</div>
</div>`
	tokenImage := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE, 0)
	token := &parser.ImageToken{
		Filename:        "foo/image.jpg",
		CaptionTokenKey: "",
		SizeX:           10,
		SizeY:           20,
	}
	tokenMap := map[string]interface{}{
		tokenImage: token,
	}

	actualResult, err := generator.expand(token, tokenMap)
	test.AssertNil(t, err)
	test.AssertEqual(t, result, actualResult)
}

func TestExpandImage_onlyOneSizeSpecified(t *testing.T) {
	// Only width
	result := `<div class="figure">
<img alt="image" src="./foo/image.jpg" style="vertical-align: middle; width: 10px;">
<div class="caption">
some <b>caption</b>
</div>
</div>`
	tokenImage := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE, 1)
	tokenCaption := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE_CAPTION, 0)
	token := &parser.ImageToken{
		Filename:        "foo/image.jpg",
		CaptionTokenKey: tokenCaption,
		SizeX:           10,
		SizeY:           -1,
	}
	tokenMap := map[string]interface{}{
		tokenImage:   token,
		tokenCaption: "some " + parser.MARKER_BOLD_OPEN + "caption" + parser.MARKER_BOLD_CLOSE,
	}

	actualResult, err := generator.expand(token, tokenMap)
	test.AssertNil(t, err)
	test.AssertEqual(t, result, actualResult)

	// Only height
	result = `<div class="figure">
<img alt="image" src="./foo/image.jpg" style="vertical-align: middle; height: 10px;">
<div class="caption">
some <b>caption</b>
</div>
</div>`
	tokenImage = fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE, 1)
	tokenCaption = fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE_CAPTION, 0)
	token = &parser.ImageToken{
		Filename:        "foo/image.jpg",
		CaptionTokenKey: tokenCaption,
		SizeX:           -1,
		SizeY:           10,
	}
	tokenMap = map[string]interface{}{
		tokenImage:   token,
		tokenCaption: "some " + parser.MARKER_BOLD_OPEN + "caption" + parser.MARKER_BOLD_CLOSE,
	}

	actualResult, err = generator.expand(token, tokenMap)
	test.AssertNil(t, err)
	test.AssertEqual(t, result, actualResult)
}

func TestExpandImageInline(t *testing.T) {
	result := `<img alt="image" class="inline" src="./foo/image.jpg" style="vertical-align: middle; width: 10px; height: 20px;">`
	tokenImage := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE_INLINE, 1)
	tokenCaption := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE_CAPTION, 0)
	token := &parser.InlineImageToken{
		Filename: "foo/image.jpg",
		SizeX:    10,
		SizeY:    20,
	}
	tokenMap := map[string]interface{}{
		tokenImage:   token,
		tokenCaption: "some " + parser.MARKER_BOLD_OPEN + "caption" + parser.MARKER_BOLD_CLOSE,
	}

	actualResult, err := generator.expand(token, tokenMap)
	test.AssertNil(t, err)
	test.AssertEqual(t, result, actualResult)
}

func TestExpandInternalLink(t *testing.T) {
	tokenLink := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_INTERNAL_LINK, 0)
	tokenMap := map[string]interface{}{
		tokenLink: &parser.InternalLinkToken{
			ArticleName: "foo",
			LinkText:    "b" + parser.MARKER_BOLD_OPEN + "a" + parser.MARKER_BOLD_CLOSE + "r",
		},
	}

	link, err := generator.expand(tokenLink, tokenMap)
	test.AssertNil(t, err)
	test.AssertEqual(t, "b<b>a</b>r", link)
}

func TestExpandExternalLink(t *testing.T) {
	tokenLink := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_EXTERNAL_LINK, 0)
	url := "https://foo.com"
	tokenMap := map[string]interface{}{
		tokenLink: &parser.ExternalLinkToken{
			URL:      url,
			LinkText: "b" + parser.MARKER_BOLD_OPEN + "a" + parser.MARKER_BOLD_CLOSE + "r",
		},
	}

	link, err := generator.expand(tokenLink, tokenMap)
	test.AssertNil(t, err)
	test.AssertEqual(t, "<a href=\""+url+"\">b<b>a</b>r</a>", link)
}

func TestExpandTable(t *testing.T) {
	tokenTable := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_TABLE, 0)
	tokenRow := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_TABLE_ROW, 1)
	tokenCol := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_TABLE_COL, 2)
	tokenCaption := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_TABLE_CAPTION, 3)
	tokenMap := map[string]interface{}{
		tokenTable:   tokenRow + "" + tokenCaption,
		tokenRow:     tokenCol,
		tokenCol:     "b" + parser.MARKER_BOLD_OPEN + "a" + parser.MARKER_BOLD_CLOSE + "r",
		tokenCaption: "caption",
	}

	row, err := generator.expandTable(tokenTable, tokenMap)
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
	tokenRow := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_TABLE_ROW, 0)
	tokenMap := map[string]interface{}{
		tokenRow: "b" + parser.MARKER_BOLD_OPEN + "a" + parser.MARKER_BOLD_CLOSE + "r",
	}

	row, err := generator.expandTableRow(tokenRow, tokenMap)
	test.AssertNil(t, err)
	test.AssertEqual(t, "<tr>\nb<b>a</b>r\n</tr>\n", row)
}

func TestExpandTableColumn(t *testing.T) {
	tokenCol := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_TABLE_COL, 0)
	tokenMap := map[string]interface{}{
		tokenCol: "b" + parser.MARKER_BOLD_OPEN + "a" + parser.MARKER_BOLD_CLOSE + "r",
	}

	row, err := generator.expandTableColumn(tokenCol, tokenMap, TABLE_TEMPLATE_COL)
	test.AssertNil(t, err)
	test.AssertEqual(t, "<td>\nb<b>a</b>r\n</td>\n", row)
}

func TestExpandTableColumnWithAttributes(t *testing.T) {
	tokenCol := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_TABLE_COL, 0)
	tokenAttrib := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_TABLE_COL_ATTRIBUTES, 1)

	tokenMap := map[string]interface{}{
		tokenCol:    tokenAttrib + "b" + parser.MARKER_BOLD_OPEN + "a" + parser.MARKER_BOLD_CLOSE + "r",
		tokenAttrib: "style=\"width: infinity lol\"",
	}

	row, err := generator.expandTableColumn(tokenCol, tokenMap, TABLE_TEMPLATE_COL)
	test.AssertNil(t, err)
	test.AssertEqual(t, "<td style=\"width: infinity lol\">\nb<b>a</b>r\n</td>\n", row)
}

func TestExpandUnorderedList(t *testing.T) {
	tokenLi1 := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_LIST_ITEM, 0)
	tokenLi2 := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_LIST_ITEM, 1)
	tokenList := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_UNORDERED_LIST, 2)

	tokenMap := map[string]interface{}{
		tokenList: tokenLi1 + tokenLi2,
		tokenLi1:  "foo",
		tokenLi2:  fmt.Sprintf("b%sa%sr", parser.MARKER_BOLD_OPEN, parser.MARKER_BOLD_CLOSE),
	}

	row, err := generator.expandUnorderedList(tokenList, tokenMap)
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
	tokenLi1 := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_LIST_ITEM, 0)
	tokenLi2 := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_LIST_ITEM, 1)
	tokenList := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_ORDERED_LIST, 2)

	tokenMap := map[string]interface{}{
		tokenList: tokenLi1 + tokenLi2,
		tokenLi1:  "foo",
		tokenLi2:  fmt.Sprintf("b%sa%sr", parser.MARKER_BOLD_OPEN, parser.MARKER_BOLD_CLOSE),
	}

	row, err := generator.expandOrderedList(tokenList, tokenMap)
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

func TestExpandDescriptionList(t *testing.T) {
	tokenLi1 := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_DESCRIPTION_LIST_HEAD, 0)
	tokenLi2 := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_DESCRIPTION_LIST_ITEM, 1)
	tokenList := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_DESCRIPTION_LIST, 2)

	tokenMap := map[string]interface{}{
		tokenList: tokenLi1 + tokenLi2,
		tokenLi1:  "foo",
		tokenLi2:  fmt.Sprintf("b%sa%sr", parser.MARKER_BOLD_OPEN, parser.MARKER_BOLD_CLOSE),
	}

	row, err := generator.expandDescriptionList(tokenList, tokenMap)
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

func TestExpandRefDefinition(t *testing.T) {
	tokenRef := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_REF_DEF, 2)

	tokenMap := map[string]interface{}{
		tokenRef: "42 f" + parser.MARKER_BOLD_OPEN + "o" + parser.MARKER_BOLD_CLOSE + "o",
	}

	row, err := generator.expandRefDefinition(tokenRef, tokenMap)
	test.AssertNil(t, err)
	test.AssertEqual(t, `[42] f<b>o</b>o<br>`, row)
}

func TestExpandRefUsage(t *testing.T) {
	tokenRef := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_REF_DEF, 2)

	tokenMap := map[string]interface{}{
		tokenRef: "42 f" + parser.MARKER_BOLD_OPEN + "o" + parser.MARKER_BOLD_CLOSE + "o",
	}

	row, err := generator.expandRefUsage(tokenRef, tokenMap)
	test.AssertNil(t, err)
	test.AssertEqual(t, `[42]`, row)
}
