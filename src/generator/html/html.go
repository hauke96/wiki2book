package html

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"wiki2book/api"
	"wiki2book/parser"
	"wiki2book/util"
)

const HEADER = `<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
<meta charset="utf-8">
<link rel="stylesheet" href="{{STYLE}}">
</head>
<body xmlns:epub="http://www.idpf.org/2007/ops">
`
const FOOTER = `
</body>
</html>
`

const HREF_TEMPLATE = `<a href="%s">%s</a>`
const STYLE_TEMPLATE = `style="%s"`
const IMAGE_SIZE_ALIGN_TEMPLATE = `vertical-align: middle;`
const IMAGE_SIZE_WIDTH_TEMPLATE = `width: %dpx;`
const IMAGE_SIZE_HEIGHT_TEMPLATE = `height: %dpx;`
const IMAGE_INLINE_TEMPLATE = `<img alt="image" class="inline" src="./%s" %s>`
const IMAGE_TEMPLATE = `<div class="figure">
<img alt="image" src="./%s" %s>
<div class="caption">
%s
</div>
</div>`
const MATH_TEMPLATE = `<img alt="image" src="./%s" style="width: %s; height: %s; %s">`
const TABLE_TEMPLATE = `<div class="figure">
<table>
%s
</table>
<div class="caption">
%s
</div>
</div>`
const TABLE_TEMPLATE_HEAD = `<th%s>
%s
</th>
`
const TABLE_TEMPLATE_ROW = `<tr>
%s
</tr>
`
const TABLE_TEMPLATE_COL = `<td%s>
%s
</td>
`
const TEMPLATE_UL = `<ul>
%s</ul>`
const TEMPLATE_OL = `<ol>
%s</ol>`
const TEMPLATE_DL = `<div class="description-list">
%s</div>` // Use bare div-tags instead of <dl> due to eBook-reader incompatibilities :(
const TEMPLATE_LI = `<li>
%s
</li>
`
const TEMPLATE_DT = `<div class="dt">
%s
</div>
`
const TEMPLATE_DD = `<div class="dd">
%s
</div>
`
const TEMPLATE_HEADING = "<h%d>%s</h%d>"
const TEMPLATE_REF_DEF = "[%d] %s<br>"
const TEMPLATE_REF_USAGE = "[%d]"

var (
	tokenRegex                  = regexp.MustCompile(parser.TOKEN_REGEX)
	tableColAttributeTokenRegex = regexp.MustCompile(`\$\$TOKEN_` + parser.TOKEN_TABLE_COL_ATTRIBUTES + `_\d+\$\$`)
)

type HtmlGenerator struct {
	imageCacheFolder   string
	mathCacheFolder    string
	articleCacheFolder string
}

// Generate creates the HTML for the given article and returns either the HTML file path or an error.
func (g *HtmlGenerator) Generate(wikiArticle *parser.Article, outputFolder string, styleFile string, imgFolder string, mathFolder string, articleFolder string) (string, error) {
	g.imageCacheFolder = imgFolder
	g.mathCacheFolder = mathFolder
	g.articleCacheFolder = articleFolder

	content := strings.ReplaceAll(HEADER, "{{STYLE}}", styleFile)
	content += "\n<h1>" + wikiArticle.Title + "</h1>\n"
	expandedContent, err := g.expand(wikiArticle.Content, wikiArticle.TokenMap)
	if err != nil {
		return "", err
	}
	content += expandedContent
	content += FOOTER
	return write(wikiArticle.Title, outputFolder, content)
}

func (g *HtmlGenerator) expand(content interface{}, tokenMap map[string]interface{}) (string, error) {
	switch content.(type) {
	case string:
		return g.expandString(content.(string), tokenMap)
	case parser.Token:
		return g.expandToken(content.(parser.Token), tokenMap)
	}

	return "", errors.New(fmt.Sprintf("Unsupported type to expand: %T", content))
}

func (g *HtmlGenerator) expandToken(token parser.Token, tokenMap map[string]interface{}) (string, error) {
	var err error = nil
	var html = ""

	switch token.(type) {
	case *parser.HeadingToken:
	case *parser.InlineImageToken:
		html, err = g.expandInlineImage(token.(*parser.InlineImageToken))
	case *parser.ImageToken:
		html, err = g.expandImage(token.(*parser.ImageToken), tokenMap)
	}

	if err != nil {
		return "", err
	}

	return html, nil
}

func (g *HtmlGenerator) expandString(content string, tokenMap map[string]interface{}) (string, error) {
	content = g.expandMarker(content)

	submatches := tokenRegex.FindAllStringSubmatch(content, -1)

	if len(submatches) == 0 {
		// no token in content
		return content, nil
	}

	for _, submatch := range submatches {
		// TODO Only print when --trace is active (-> #35)
		//sigolo.Debug("Found token %s", submatch[1])

		if _, ok := tokenMap[submatch[0]]; !ok {
			return "", errors.New(fmt.Sprintf("Token key %s not found in token map", submatch[0]))
		}

		html := submatch[0]
		var err error = nil

		switch submatch[1] {
		case parser.TOKEN_EXTERNAL_LINK:
			html, err = g.expandExternalLink(submatch[0], tokenMap)
		case parser.TOKEN_INTERNAL_LINK:
			html, err = g.expandInternalLink(submatch[0], tokenMap)
		case parser.TOKEN_TABLE:
			html, err = g.expandTable(submatch[0], tokenMap)
		case parser.TOKEN_TABLE_HEAD:
			html, err = g.expandTableColumn(submatch[0], tokenMap, TABLE_TEMPLATE_HEAD)
		case parser.TOKEN_TABLE_ROW:
			html, err = g.expandTableRow(submatch[0], tokenMap)
		case parser.TOKEN_TABLE_COL:
			html, err = g.expandTableColumn(submatch[0], tokenMap, TABLE_TEMPLATE_COL)
		case parser.TOKEN_UNORDERED_LIST:
			html, err = g.expandUnorderedList(submatch[0], tokenMap)
		case parser.TOKEN_ORDERED_LIST:
			html, err = g.expandOrderedList(submatch[0], tokenMap)
		case parser.TOKEN_DESCRIPTION_LIST:
			html, err = g.expandDescriptionList(submatch[0], tokenMap)
		case parser.TOKEN_LIST_ITEM:
			html, err = g.expandListItem(submatch[0], tokenMap)
		case parser.TOKEN_DESCRIPTION_LIST_HEAD:
			html, err = g.expandDescriptionHead(submatch[0], tokenMap)
		case parser.TOKEN_DESCRIPTION_LIST_ITEM:
			html, err = g.expandDescriptionItem(submatch[0], tokenMap)
		case parser.TOKEN_IMAGE_INLINE:
			html, err = g.expandInlineImage(tokenMap[submatch[0]].(*parser.InlineImageToken))
		case parser.TOKEN_IMAGE:
			html, err = g.expandImage(tokenMap[submatch[0]].(*parser.ImageToken), tokenMap)
		case parser.TOKEN_MATH:
			html, err = g.expandMath(submatch[0], tokenMap)
		case parser.TOKEN_HEADING:
			html, err = g.expandHeadings(tokenMap[submatch[0]].(*parser.HeadingToken), tokenMap)
		case parser.TOKEN_REF_DEF:
			html, err = g.expandRefDefinition(submatch[0], tokenMap)
		case parser.TOKEN_REF_USAGE:
			html, err = g.expandRefUsage(submatch[0], tokenMap)
		}
		if err != nil {
			return "", err
		}

		content = strings.Replace(content, submatch[0], html, 1)
	}

	return content, nil
}

func (g *HtmlGenerator) expandMarker(content string) string {
	content = strings.ReplaceAll(content, parser.MARKER_BOLD_OPEN, "<b>")
	content = strings.ReplaceAll(content, parser.MARKER_BOLD_CLOSE, "</b>")
	content = strings.ReplaceAll(content, parser.MARKER_ITALIC_OPEN, "<i>")
	content = strings.ReplaceAll(content, parser.MARKER_ITALIC_CLOSE, "</i>")

	// TODO Maybe use <p> and </p> instead of <br>? That would need additional parsing to find exact
	// passages of the paragraph (to create the two separate start/end markers for <p> and </p>).
	content = strings.ReplaceAll(content, parser.MARKER_PARAGRAPH, "<br><br>")
	return content
}

// expandHeadings expands a heading with the given leven (e.g. 4 for <h4> headings)
func (g *HtmlGenerator) expandHeadings(token *parser.HeadingToken, tokenMap map[string]interface{}) (string, error) {
	expandedHeadingText, err := g.expand(token.Content, tokenMap)
	if err != nil {
		return "", err
	}
	return g.expand(fmt.Sprintf(TEMPLATE_HEADING, token.Depth, expandedHeadingText, token.Depth), tokenMap)
}

func (g *HtmlGenerator) expandInlineImage(token *parser.InlineImageToken) (string, error) {
	sizeTemplate := expandSizeTemplate(token.SizeX, token.SizeY)
	return fmt.Sprintf(IMAGE_INLINE_TEMPLATE, token.Filename, sizeTemplate), nil
}

func (g *HtmlGenerator) expandImage(token *parser.ImageToken, tokenMap map[string]interface{}) (string, error) {
	caption := ""
	var err error = nil

	captionToken, hasCaptionToken := tokenMap[token.CaptionTokenKey]
	if hasCaptionToken {
		caption, err = g.expand(captionToken, tokenMap)
		if err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("Error while parsing image token %#v", token))
		}
	}

	sizeTemplate := expandSizeTemplate(token.SizeX, token.SizeY)

	return fmt.Sprintf(IMAGE_TEMPLATE, token.Filename, sizeTemplate, caption), nil
}

func expandSizeTemplate(xSize int, ySize int) string {
	sizeTemplate := ""
	if xSize != -1 || ySize != -1 {
		styles := []string{IMAGE_SIZE_ALIGN_TEMPLATE}
		if xSize != -1 {
			styles = append(styles, fmt.Sprintf(IMAGE_SIZE_WIDTH_TEMPLATE, int(xSize)))
		}
		if ySize != -1 {
			styles = append(styles, fmt.Sprintf(IMAGE_SIZE_HEIGHT_TEMPLATE, int(ySize)))
		}
		sizeTemplate = fmt.Sprintf(STYLE_TEMPLATE, strings.Join(styles, " "))
	}
	return sizeTemplate
}

func (g *HtmlGenerator) expandInternalLink(token string, tokenMap map[string]interface{}) (string, error) {
	tokenContentParts := strings.Split(tokenMap[token].(string), " ")
	// Currently links are not added to the eBook, even though it's possible. Maybe this will be made configurable in
	// the future.
	return g.expand(tokenMap[tokenContentParts[1]], tokenMap)
}

func (g *HtmlGenerator) expandExternalLink(token string, tokenMap map[string]interface{}) (string, error) {
	splitToken := strings.Split(tokenMap[token].(string), " ")
	url := tokenMap[splitToken[0]]
	text, err := g.expand(tokenMap[splitToken[1]], tokenMap)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(HREF_TEMPLATE, url, text), nil
}

func (g *HtmlGenerator) expandTable(token string, tokenMap map[string]interface{}) (string, error) {
	tokenContent := tokenMap[token]
	expandedTokenContent, err := g.expand(tokenContent, tokenMap)
	if err != nil {
		return "", err
	}

	caption := ""
	for _, subToken := range strings.Split(expandedTokenContent, " ") {
		match := tokenRegex.FindStringSubmatch(subToken)
		if len(match) < 2 {
			continue
		}
		subTokenKey := match[1]
		hasCaption := subTokenKey == parser.TOKEN_TABLE_CAPTION
		if hasCaption {
			caption, err = g.expand(tokenMap[match[0]], tokenMap)
			if err != nil {
				return "", err
			}
			expandedTokenContent = strings.Replace(expandedTokenContent, match[0], "", 1)
			break
		}
	}

	return fmt.Sprintf(TABLE_TEMPLATE, expandedTokenContent, caption), nil
}

func (g *HtmlGenerator) expandTableRow(token string, tokenMap map[string]interface{}) (string, error) {
	return g.expandSimple(token, tokenMap, TABLE_TEMPLATE_ROW)
}

func (g *HtmlGenerator) expandTableColumn(token string, tokenMap map[string]interface{}, template string) (string, error) {
	tokenContent := tokenMap[token].(string)

	attributes := ""
	if strings.Contains(tokenContent, parser.TOKEN_TABLE_COL_ATTRIBUTES) {
		attributeToken := tableColAttributeTokenRegex.FindString(tokenContent)
		attributes = " " + tokenMap[attributeToken].(string)
		tokenContent = strings.Replace(tokenContent, attributeToken, "", 1)
	}

	expandedTokenContent, err := g.expand(tokenContent, tokenMap)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(template, attributes, expandedTokenContent), nil
}

func (g *HtmlGenerator) expandUnorderedList(token string, tokenMap map[string]interface{}) (string, error) {
	return g.expandSimple(token, tokenMap, TEMPLATE_UL)
}

func (g *HtmlGenerator) expandOrderedList(token string, tokenMap map[string]interface{}) (string, error) {
	return g.expandSimple(token, tokenMap, TEMPLATE_OL)
}

func (g *HtmlGenerator) expandDescriptionList(token string, tokenMap map[string]interface{}) (string, error) {
	return g.expandSimple(token, tokenMap, TEMPLATE_DL)
}

func (g *HtmlGenerator) expandListItem(token string, tokenMap map[string]interface{}) (string, error) {
	return g.expandSimple(token, tokenMap, TEMPLATE_LI)
}

func (g *HtmlGenerator) expandDescriptionHead(token string, tokenMap map[string]interface{}) (string, error) {
	return g.expandSimple(token, tokenMap, TEMPLATE_DT)
}

func (g *HtmlGenerator) expandDescriptionItem(token string, tokenMap map[string]interface{}) (string, error) {
	return g.expandSimple(token, tokenMap, TEMPLATE_DD)
}

func (g *HtmlGenerator) expandRefDefinition(token string, tokenMap map[string]interface{}) (string, error) {
	tokenContent := tokenMap[token].(string)
	tokenContentParts := strings.SplitN(tokenContent, " ", 2)
	refIndex, err := strconv.Atoi(tokenContentParts[0])
	if err != nil {
		return "", err
	}

	expandedRefContent, err := g.expand(tokenContentParts[1], tokenMap)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(TEMPLATE_REF_DEF, refIndex, expandedRefContent), nil
}

func (g *HtmlGenerator) expandRefUsage(token string, tokenMap map[string]interface{}) (string, error) {
	tokenContent := tokenMap[token].(string)
	tokenContentParts := strings.SplitN(tokenContent, " ", 2)
	refIndex, err := strconv.Atoi(tokenContentParts[0])
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(TEMPLATE_REF_USAGE, refIndex), nil
}

// TODO Create service class with public interface for the api functions (like RenderMath) to be able to mock that service.
func (g *HtmlGenerator) expandMath(token string, tokenMap map[string]interface{}) (string, error) {
	svgFilename, pngFilename, err := api.RenderMath(tokenMap[token].(string), g.imageCacheFolder, g.mathCacheFolder)
	if err != nil {
		return "", err
	}

	svg, err := util.ReadSimpleAvgAttributes(svgFilename)
	if err != nil {
		return "", err
	}

	sigolo.Debug("File: %s, Width: %s, Height: %s, Style: %s", pngFilename, svg.Width, svg.Height, svg.Style)

	return fmt.Sprintf(MATH_TEMPLATE, pngFilename, svg.Width, svg.Height, svg.Style), nil
}

func (g *HtmlGenerator) expandSimple(token string, tokenMap map[string]interface{}, template string) (string, error) {
	tokenContent := tokenMap[token]
	expandedTokenContent, err := g.expand(tokenContent, tokenMap)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(template, expandedTokenContent), nil
}

// write returns the output path or an error.
func write(title string, outputFolder string, content string) (string, error) {
	// Create the output folder
	err := os.Mkdir(outputFolder, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return "", errors.Wrap(err, fmt.Sprintf("Unable to create output folder %s", outputFolder))
	}

	// Create output file
	outputFilepath := filepath.Join(outputFolder, title+".html")
	sigolo.Debug("Write to %s", outputFilepath)
	outputFile, err := os.Create(outputFilepath)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Unable to create output file %s", outputFilepath))
	}
	defer outputFile.Close()

	// Write data to file
	_, err = outputFile.WriteString(content)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Unable write data to file %s", outputFilepath))
	}

	return outputFilepath, nil
}
