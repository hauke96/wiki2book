package html

import (
	"fmt"
	"github.com/hauke96/wiki2book/src/parser"
	"github.com/hauke96/wiki2book/src/test"
	"testing"
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
	tokenMap := map[string]string{
		"foo": "bar",
	}

	for i := 1; i <= 7; i++ {
		headings, err := generator.expandHeadings("foo", tokenMap, i)
		test.AssertNil(t, err)
		test.AssertEqual(t, fmt.Sprintf("<h%d>bar</h%d>", i, i), headings)
	}
}

func TestExpandImage(t *testing.T) {
	result := `<div class="figure">
<img alt="image" src="./foo/image.jpg" style="vertical-align: middle; width: 10px; height: 20px;">
<div class="caption">
some <b>caption</b>
</div>
</div>`
	tokenImage := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE, 3)
	tokenString := tokenImage
	tokenFilename := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE_FILENAME, 0)
	tokenCaption := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE_CAPTION, 2)
	tokenImageSize := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE_SIZE, 1)
	tokenMap := map[string]string{
		tokenImage:     tokenFilename + " " + tokenCaption + " " + tokenImageSize,
		tokenFilename:  "foo/image.jpg",
		tokenCaption:   "some " + parser.MARKER_BOLD_OPEN + "caption" + parser.MARKER_BOLD_CLOSE,
		tokenImageSize: "10x20",
	}

	actualResult, err := generator.expandImage(tokenString, tokenMap)
	test.AssertNil(t, err)
	test.AssertEqual(t, result, actualResult)

	actualResult, err = generator.expandImage("$$TOKEN_"+parser.TOKEN_IMAGE+"_23852376$$", tokenMap)
	test.AssertNotNil(t, err)
	test.AssertEqual(t, "", actualResult)
}

func TestExpandImageInline(t *testing.T) {
	result := `<img alt="image" class="inline" src="./foo/image.jpg" style="vertical-align: middle; width: 10px; height: 20px;">`
	tokenImage := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE_INLINE, 3)
	tokenString := tokenImage
	tokenFilename := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE_FILENAME, 0)
	tokenCaption := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE_CAPTION, 2)
	tokenImageSize := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE_SIZE, 1)
	tokenMap := map[string]string{
		tokenImage:     tokenFilename + " " + tokenCaption + " " + tokenImageSize,
		tokenFilename:  "foo/image.jpg",
		tokenCaption:   "some " + parser.MARKER_BOLD_OPEN + "caption" + parser.MARKER_BOLD_CLOSE,
		tokenImageSize: "10x20",
	}

	actualResult, err := generator.expandImage(tokenString, tokenMap)
	test.AssertNil(t, err)
	test.AssertEqual(t, result, actualResult)
}

func TestExpandInternalLink(t *testing.T) {
	tokenLink := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_INTERNAL_LINK, 0)
	tokenArticle := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_INTERNAL_LINK_ARTICLE, 1)
	tokenText := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_INTERNAL_LINK_TEXT, 2)
	tokenMap := map[string]string{
		tokenLink:    tokenArticle + " " + tokenText,
		tokenArticle: "foo",
		tokenText:    "b" + parser.MARKER_BOLD_OPEN + "a" + parser.MARKER_BOLD_CLOSE + "r",
	}
	tokenString := tokenLink

	link, err := generator.expandInternalLink(tokenString, tokenMap)
	test.AssertNil(t, err)
	test.AssertEqual(t, "b<b>a</b>r", link)
}

func TestExpandExternalLink(t *testing.T) {
	tokenLink := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_EXTERNAL_LINK, 0)
	tokenUrl := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_EXTERNAL_LINK_URL, 1)
	tokenText := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_EXTERNAL_LINK_TEXT, 2)
	url := "https://foo.com"
	tokenMap := map[string]string{
		tokenLink: tokenUrl + " " + tokenText,
		tokenUrl:  url,
		tokenText: "b" + parser.MARKER_BOLD_OPEN + "a" + parser.MARKER_BOLD_CLOSE + "r",
	}
	tokenString := tokenLink

	link, err := generator.expandExternalLink(tokenString, tokenMap)
	test.AssertNil(t, err)
	test.AssertEqual(t, "<a href=\""+url+"\">b<b>a</b>r</a>", link)
}
