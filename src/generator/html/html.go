package html

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"regexp"
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
const IMAGE_SIZE_WIDTH_AUTO_TEMPLATE = `width: auto;`
const IMAGE_SIZE_HEIGHT_AUTO_TEMPLATE = `height: auto;`
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
</table>%s
</div>`
const TABLE_TEMPLATE_CAPTION = `
<div class="caption">
%s
</div>`
const TABLE_TEMPLATE_HEAD = `<th%s>
%s
</th>
`
const TABLE_TEMPLATE_ROW = `<tr>
%s
</tr>`
const TABLE_TEMPLATE_COL = `<td%s>
%s
</td>`
const TEMPLATE_UL = `<ul>
%s
</ul>`
const TEMPLATE_OL = `<ol>
%s
</ol>`
const TEMPLATE_DL = `<div class="description-list">
%s
</div>` // Use bare div-tags instead of <dl> due to eBook-reader incompatibilities :(
const TEMPLATE_LI = `<li>
%s
</li>`
const TEMPLATE_DT = `<div class="dt">
%s
</div>`
const TEMPLATE_DD = `<div class="dd">
%s
</div>`
const TEMPLATE_HEADING = "<h%d>%s</h%d>"
const TEMPLATE_REF_DEF = "[%d] %s<br>"
const TEMPLATE_REF_USAGE = "[%d]"

var (
	tokenRegex = regexp.MustCompile(parser.TOKEN_REGEX)
)

type HtmlGenerator struct {
	ImageCacheFolder   string
	MathCacheFolder    string
	ArticleCacheFolder string
	TokenMap           map[string]parser.Token
}

// Generate creates the HTML for the given article and returns either the HTML file path or an error.
func (g *HtmlGenerator) Generate(wikiArticle *parser.Article, outputFolder string, styleFile string) (string, error) {
	content := strings.ReplaceAll(HEADER, "{{STYLE}}", styleFile)
	content += "\n<h1>" + wikiArticle.Title + "</h1>\n"
	expandedContent, err := g.expand(wikiArticle.Content)
	if err != nil {
		return "", err
	}
	content += expandedContent
	content += FOOTER
	return write(wikiArticle.Title, outputFolder, content)
}

func (g *HtmlGenerator) expand(content interface{}) (string, error) {
	switch content.(type) {
	case string:
		return g.expandString(content.(string))
	case parser.Token:
		return g.expandToken(content.(parser.Token))
	}

	return "", errors.New(fmt.Sprintf("Unsupported type to expand: %T", content))
}

func (g *HtmlGenerator) expandToken(token parser.Token) (string, error) {
	var err error = nil
	var html = ""

	switch t := token.(type) {
	case parser.HeadingToken:
		html, err = g.expandHeadings(t)
	case parser.InlineImageToken:
		html, err = g.expandInlineImage(t)
	case parser.ImageToken:
		html, err = g.expandImage(t)
	case parser.ExternalLinkToken:
		html, err = g.expandExternalLink(t)
	case parser.InternalLinkToken:
		html, err = g.expandInternalLink(t)
	case parser.UnorderedListToken:
		html, err = g.expandUnorderedList(t)
	case parser.OrderedListToken:
		html, err = g.expandOrderedList(t)
	case parser.DescriptionListToken:
		html, err = g.expandDescriptionList(t)
	case parser.ListItemToken:
		html, err = g.expandListItem(t)
	case parser.TableToken:
		html, err = g.expandTable(t)
	case parser.TableRowToken:
		html, err = g.expandTableRow(t)
	case parser.TableColToken:
		html, err = g.expandTableColumn(t)
	case parser.TableCaptionToken:
		html, err = g.expandTableCaption(t)
	case parser.MathToken:
		html, err = g.expandMath(t)
	case parser.RefDefinitionToken:
		html, err = g.expandRefDefinition(t)
	case parser.RefUsageToken:
		html, err = g.expandRefUsage(t)
	case parser.NowikiToken:
		html, err = g.expandNowiki(t)
	}

	if err != nil {
		return "", err
	}

	return html, nil
}

func (g *HtmlGenerator) expandString(content string) (string, error) {
	content = g.expandMarker(content)

	matches := tokenRegex.FindAllString(content, -1)

	if len(matches) == 0 {
		// no token in content
		return content, nil
	}

	for _, tokenKey := range matches {
		tokenContent, hasTokenKey := g.TokenMap[tokenKey]
		if !hasTokenKey {
			return "", errors.New(fmt.Sprintf("Token key %s not found in token map", tokenKey))
		}
		sigolo.Trace("Found token %s -> %#v", tokenKey, tokenContent)

		html, err := g.expand(tokenContent)
		if err != nil {
			return "", err
		}

		content = strings.Replace(content, tokenKey, html, 1)
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
func (g *HtmlGenerator) expandHeadings(token parser.HeadingToken) (string, error) {
	expandedHeadingText, err := g.expand(token.Content)
	if err != nil {
		return "", err
	}
	return g.expand(fmt.Sprintf(TEMPLATE_HEADING, token.Depth, expandedHeadingText, token.Depth))
}

func (g *HtmlGenerator) expandInlineImage(token parser.InlineImageToken) (string, error) {
	sizeTemplate := expandSizeTemplate(token.SizeX, token.SizeY)
	return fmt.Sprintf(IMAGE_INLINE_TEMPLATE, token.Filename, sizeTemplate), nil
}

func (g *HtmlGenerator) expandImage(token parser.ImageToken) (string, error) {
	caption, err := g.expand(token.Caption.Content)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Error while expanding caption of image %#v", token))
	}

	sizeTemplate := expandSizeTemplate(token.SizeX, token.SizeY)

	return fmt.Sprintf(IMAGE_TEMPLATE, token.Filename, sizeTemplate, caption), nil
}

func expandSizeTemplate(xSize int, ySize int) string {
	sizeTemplate := ""
	if xSize != -1 || ySize != -1 {
		styles := []string{IMAGE_SIZE_ALIGN_TEMPLATE}

		if xSize != -1 {
			styles = append(styles, fmt.Sprintf(IMAGE_SIZE_WIDTH_TEMPLATE, xSize))
		} else {
			// Allow scaling with correct aspect ratio in case there's no width specified
			styles = append(styles, IMAGE_SIZE_WIDTH_AUTO_TEMPLATE)
		}

		if ySize != -1 {
			styles = append(styles, fmt.Sprintf(IMAGE_SIZE_HEIGHT_TEMPLATE, ySize))
		} else {
			// Allow scaling with correct aspect ratio in case there's no height specified
			styles = append(styles, IMAGE_SIZE_HEIGHT_AUTO_TEMPLATE)
		}

		sizeTemplate = fmt.Sprintf(STYLE_TEMPLATE, strings.Join(styles, " "))
	}
	return sizeTemplate
}

func (g *HtmlGenerator) expandInternalLink(token parser.InternalLinkToken) (string, error) {
	// Currently links are not added to the eBook, even though it's possible. Maybe this will be made configurable in
	// the future.
	return g.expand(token.LinkText)
}

func (g *HtmlGenerator) expandExternalLink(token parser.ExternalLinkToken) (string, error) {
	text, err := g.expand(token.LinkText)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(HREF_TEMPLATE, token.URL, text), nil
}

func (g *HtmlGenerator) expandTable(token parser.TableToken) (string, error) {
	var expandedRows []string
	for _, rowToken := range token.Rows {
		expandedRow, err := g.expand(rowToken)
		if err != nil {
			return "", err
		}

		expandedRows = append(expandedRows, expandedRow)
	}

	expandedCaption, err := g.expand(token.Caption)
	if err != nil {
		return "", err
	}

	joinedRows := strings.Join(expandedRows, "\n")
	return fmt.Sprintf(TABLE_TEMPLATE, joinedRows, expandedCaption), nil
}

func (g *HtmlGenerator) expandTableRow(token parser.TableRowToken) (string, error) {
	var expandedCols []string
	for _, colToken := range token.Columns {
		expandedCol, err := g.expand(colToken)
		if err != nil {
			return "", err
		}
		expandedCols = append(expandedCols, expandedCol)
	}

	joinedColumns := strings.Join(expandedCols, "\n")
	return fmt.Sprintf(TABLE_TEMPLATE_ROW, joinedColumns), nil
}

func (g *HtmlGenerator) expandTableColumn(token parser.TableColToken) (string, error) {
	expandedTokenContent, err := g.expand(token.Content)
	if err != nil {
		return "", err
	}
	attributes := strings.Join(token.Attributes.Attributes, " ")
	if attributes != "" {
		attributes = " " + attributes
	}

	template := TABLE_TEMPLATE_COL
	if token.IsHeading {
		template = TABLE_TEMPLATE_HEAD
	}

	return fmt.Sprintf(template, attributes, expandedTokenContent), nil
}

func (g *HtmlGenerator) expandTableCaption(token parser.TableCaptionToken) (string, error) {
	expandedTokenContent, err := g.expand(token.Content)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(TABLE_TEMPLATE_CAPTION, expandedTokenContent), nil
}

func (g *HtmlGenerator) expandUnorderedList(token parser.UnorderedListToken) (string, error) {
	expandedItems, err := g.expandListItems(token.Items)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(TEMPLATE_UL, expandedItems), nil
}

func (g *HtmlGenerator) expandOrderedList(token parser.OrderedListToken) (string, error) {
	expandedItems, err := g.expandListItems(token.Items)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(TEMPLATE_OL, expandedItems), nil
}

func (g *HtmlGenerator) expandDescriptionList(token parser.DescriptionListToken) (string, error) {
	expandedItems, err := g.expandListItems(token.Items)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(TEMPLATE_DL, expandedItems), nil
}

func (g *HtmlGenerator) expandListItems(items []parser.ListItemToken) (string, error) {
	var expandedItems []string

	for _, item := range items {
		expandedItem, err := g.expandListItem(item)
		if err != nil {
			return "", err
		}

		expandedItems = append(expandedItems, expandedItem)
	}

	expandedTokenContent := strings.Join(expandedItems, "\n")
	return expandedTokenContent, nil
}

func (g *HtmlGenerator) expandListItem(token parser.ListItemToken) (string, error) {
	var template string
	switch token.Type {
	case parser.NORMAL_ITEM:
		template = TEMPLATE_LI
	case parser.DESCRIPTION_HEAD:
		template = TEMPLATE_DT
	case parser.DESCRIPTION_ITEM:
		template = TEMPLATE_DD
	default:
		return "", errors.New(fmt.Sprintf("Unknown list item type '%d'", token.Type))
	}

	var listItemContents []string
	listItemContent, err := g.expand(token.Content)
	if err != nil {
		return "", err
	}
	listItemContents = append(listItemContents, listItemContent)

	for _, subListToken := range token.SubLists {
		expandedSubList, err := g.expand(subListToken)
		if err != nil {
			return "", err
		}
		listItemContents = append(listItemContents, expandedSubList)
	}

	return fmt.Sprintf(template, strings.Join(listItemContents, "\n")), nil
}

func (g *HtmlGenerator) expandRefDefinition(token parser.RefDefinitionToken) (string, error) {
	expandedRefContent, err := g.expand(token.Content)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(TEMPLATE_REF_DEF, token.Index+1, expandedRefContent), nil
}

func (g *HtmlGenerator) expandRefUsage(token parser.RefUsageToken) (string, error) {
	return fmt.Sprintf(TEMPLATE_REF_USAGE, token.Index+1), nil
}

// TODO Create service class with public interface for the api functions (like RenderMath) to be able to mock that service.
func (g *HtmlGenerator) expandMath(token parser.MathToken) (string, error) {
	svgFilename, pngFilename, err := api.RenderMath(token.Content, g.ImageCacheFolder, g.MathCacheFolder)
	if err != nil {
		return "", err
	}

	svg, err := util.ReadSimpleAvgAttributes(svgFilename)
	if err != nil {
		return "", err
	}

	sigolo.Debug("Expanded math | file: %s, width: %s, height: %s, style: %s", pngFilename, svg.Width, svg.Height, svg.Style)

	return fmt.Sprintf(MATH_TEMPLATE, pngFilename, svg.Width, svg.Height, svg.Style), nil
}

func (g *HtmlGenerator) expandNowiki(token parser.NowikiToken) (string, error) {
	return token.Content, nil
}

// write returns the output path or an error.
func write(title string, outputFolder string, content string) (string, error) {
	// Create the output folder
	sigolo.Debug("Ensure output folder '%s'", outputFolder)
	err := os.Mkdir(outputFolder, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return "", errors.Wrap(err, fmt.Sprintf("Unable to create output folder %s", outputFolder))
	}

	// Create output file
	outputFilepath := filepath.Join(outputFolder, title+".html")
	sigolo.Debug("Ensure output file '%s'", outputFilepath)
	outputFile, err := os.Create(outputFilepath)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Unable to create output file %s", outputFilepath))
	}
	defer outputFile.Close()

	// Write data to file
	sigolo.Debug("Write to %s", outputFilepath)
	_, err = outputFile.WriteString(content)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Unable write data to file %s", outputFilepath))
	}

	return outputFilepath, nil
}
