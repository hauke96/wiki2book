package html

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/api"
	"github.com/hauke96/wiki2book/src/parser"
	"github.com/hauke96/wiki2book/src/util"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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

const HREF_TEMPLATE = "<a href=\"%s\">%s</a>"
const IMAGE_SIZE_TEMPLATE = `style="vertical-align: middle; width: %spx; height: %spx;"`
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
const TEMPLATE_DL = `<div class="list">
%s</div>`
const TEMPLATE_LI = `<li>
%s
</li>
`
const TEMPLATE_DD = `<div>
%s
</div>
`
const TEMPLATE_HEADING = "<h%d>%s</h%d>"
const TEMPLATE_REF_DEF = "[%d] %s<br>"
const TEMPLATE_REF_USAGE = "[%d]"

type HtmlGenerator struct {
	imageCacheFolder   string
	mathCacheFolder    string
	articleCacheFolder string
}

func (g *HtmlGenerator) Generate(wikiArticle parser.Article, outputFolder string, styleFile string, imgFolder string, mathFolder string, articleFolder string) (string, error) {
	g.imageCacheFolder = imgFolder
	g.mathCacheFolder = mathFolder
	g.articleCacheFolder = articleFolder

	err := api.DownloadImages(wikiArticle.Images, g.imageCacheFolder, g.articleCacheFolder)
	sigolo.FatalCheck(err)

	content := strings.ReplaceAll(HEADER, "{{STYLE}}", styleFile)
	content += "\n<h1>" + wikiArticle.Title + "</h1>"
	expandedContent, err := g.expand(wikiArticle.Content, wikiArticle.TokenMap)
	if err != nil {
		return "", err
	}
	content += expandedContent
	content += FOOTER
	return write(wikiArticle.Title, outputFolder, content)
}

func (g *HtmlGenerator) expand(content string, tokenMap map[string]string) (string, error) {
	content = g.expandMarker(content)

	regex := regexp.MustCompile(parser.TOKEN_REGEX)
	submatches := regex.FindAllStringSubmatch(content, -1)

	if len(submatches) == 0 {
		// no token in content
		return content, nil
	}

	for _, submatch := range submatches {
		sigolo.Debug("Found token %s", submatch[1])

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
		case parser.TOKEN_DESCRIPTION_LIST_ITEM:
			html, err = g.expandDescriptionItem(submatch[0], tokenMap)
		case parser.TOKEN_IMAGE_INLINE:
			html, err = g.expandImage(submatch[0], tokenMap)
		case parser.TOKEN_IMAGE:
			html, err = g.expandImage(submatch[0], tokenMap)
		case parser.TOKEN_MATH:
			html, err = g.expandMath(submatch[0], tokenMap)
		case parser.TOKEN_HEADING_1:
			html, err = g.expandHeadings(submatch[0], tokenMap, 1)
		case parser.TOKEN_HEADING_2:
			html, err = g.expandHeadings(submatch[0], tokenMap, 2)
		case parser.TOKEN_HEADING_3:
			html, err = g.expandHeadings(submatch[0], tokenMap, 3)
		case parser.TOKEN_HEADING_4:
			html, err = g.expandHeadings(submatch[0], tokenMap, 4)
		case parser.TOKEN_HEADING_5:
			html, err = g.expandHeadings(submatch[0], tokenMap, 5)
		case parser.TOKEN_HEADING_6:
			html, err = g.expandHeadings(submatch[0], tokenMap, 6)
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
func (g *HtmlGenerator) expandHeadings(tokenString string, tokenMap map[string]string, level int) (string, error) {
	title := tokenMap[tokenString]
	return g.expand(fmt.Sprintf(TEMPLATE_HEADING, level, title, level), tokenMap)
}

func (g *HtmlGenerator) expandImage(tokenString string, tokenMap map[string]string) (string, error) {
	filename := ""
	xSize := ""
	ySize := ""
	caption := ""
	var err error = nil

	regex := regexp.MustCompile(parser.TOKEN_REGEX)
	tokenName := regex.FindStringSubmatch(tokenString)[1]
	inline := tokenName == parser.TOKEN_IMAGE_INLINE

	submatches := regex.FindAllStringSubmatch(tokenMap[tokenString], -1)

	if len(submatches) == 0 {
		return "", errors.New("No token found in image token: " + tokenString)
	}

	for _, submatch := range submatches {
		sigolo.Debug("Found sub-token %s in image token %s", submatch[1], tokenString)

		subTokenString := submatch[0]

		switch submatch[1] {
		case parser.TOKEN_IMAGE_FILENAME:
			filename = tokenMap[subTokenString]
		case parser.TOKEN_IMAGE_CAPTION:
			caption, err = g.expand(tokenMap[subTokenString], tokenMap)
		case parser.TOKEN_IMAGE_SIZE:
			sizes := strings.Split(tokenMap[subTokenString], "x")
			xSize = sizes[0]
			ySize = sizes[1]
		}
	}

	if err != nil {
		return "", errors.Wrap(err, "Error while parsing image token "+tokenString)
	}

	sizeTemplate := ""
	if xSize != "" && ySize != "" {
		sizeTemplate = fmt.Sprintf(IMAGE_SIZE_TEMPLATE, xSize, ySize)
	}

	if inline {
		return fmt.Sprintf(IMAGE_INLINE_TEMPLATE, filename, sizeTemplate), nil
	}

	return fmt.Sprintf(IMAGE_TEMPLATE, filename, sizeTemplate, caption), nil
}

func (g *HtmlGenerator) expandInternalLink(tokenString string, tokenMap map[string]string) (string, error) {
	splittedToken := strings.Split(tokenMap[tokenString], " ")
	// Yeah, let's not add an link to the article in an eBook. Maybe make it configurable some day...
	return g.expand(tokenMap[splittedToken[1]], tokenMap)
}

func (g *HtmlGenerator) expandExternalLink(tokenString string, tokenMap map[string]string) (string, error) {
	splittedToken := strings.Split(tokenMap[tokenString], " ")
	url := tokenMap[splittedToken[0]]
	text, err := g.expand(tokenMap[splittedToken[1]], tokenMap)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(HREF_TEMPLATE, url, text), nil
}

func (g *HtmlGenerator) expandTable(tokenString string, tokenMap map[string]string) (string, error) {
	tokenContent := tokenMap[tokenString]
	tokenizedContent, err := g.expand(tokenContent, tokenMap)
	if err != nil {
		return "", err
	}

	regex := regexp.MustCompile(parser.TOKEN_REGEX)
	caption := ""
	for _, subToken := range strings.Split(tokenizedContent, " ") {
		match := regex.FindStringSubmatch(subToken)
		if len(match) < 2 {
			continue
		}
		tokenName := match[1]
		hasCaption := tokenName == parser.TOKEN_TABLE_CAPTION
		if hasCaption {
			caption, err = g.expand(tokenMap[match[0]], tokenMap)
			if err != nil {
				return "", err
			}
			tokenizedContent = strings.Replace(tokenizedContent, match[0], "", 1)
			break
		}
	}

	return fmt.Sprintf(TABLE_TEMPLATE, tokenizedContent, caption), nil
}

func (g *HtmlGenerator) expandTableRow(tokenString string, tokenMap map[string]string) (string, error) {
	tokenContent := tokenMap[tokenString]
	tokenizedContent, err := g.expand(tokenContent, tokenMap)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(TABLE_TEMPLATE_ROW, tokenizedContent), nil
}

func (g *HtmlGenerator) expandTableColumn(tokenString string, tokenMap map[string]string, template string) (string, error) {
	tokenContent := tokenMap[tokenString]

	attributes := ""
	if strings.Contains(tokenContent, parser.TOKEN_TABLE_COL_ATTRIBUTES) {
		regex := regexp.MustCompile(`\$\$TOKEN_` + parser.TOKEN_TABLE_COL_ATTRIBUTES + `_\d+\$\$`)
		attributeToken := regex.FindString(tokenContent)
		attributes = " " + tokenMap[attributeToken]
		tokenContent = strings.Replace(tokenContent, attributeToken, "", 1)
	}

	tokenizedContent, err := g.expand(tokenContent, tokenMap)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(template, attributes, tokenizedContent), nil
}

func (g *HtmlGenerator) expandUnorderedList(tokenString string, tokenMap map[string]string) (string, error) {
	tokenContent := tokenMap[tokenString]
	tokenizedContent, err := g.expand(tokenContent, tokenMap)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(TEMPLATE_UL, tokenizedContent), nil
}

func (g *HtmlGenerator) expandOrderedList(tokenString string, tokenMap map[string]string) (string, error) {
	tokenContent := tokenMap[tokenString]
	tokenizedContent, err := g.expand(tokenContent, tokenMap)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(TEMPLATE_OL, tokenizedContent), nil
}

func (g *HtmlGenerator) expandDescriptionList(tokenString string, tokenMap map[string]string) (string, error) {
	tokenContent := tokenMap[tokenString]
	tokenizedContent, err := g.expand(tokenContent, tokenMap)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(TEMPLATE_DL, tokenizedContent), nil
}

func (g *HtmlGenerator) expandListItem(tokenString string, tokenMap map[string]string) (string, error) {
	tokenContent := tokenMap[tokenString]
	tokenizedContent, err := g.expand(tokenContent, tokenMap)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(TEMPLATE_LI, tokenizedContent), nil
}

func (g *HtmlGenerator) expandDescriptionItem(tokenString string, tokenMap map[string]string) (string, error) {
	tokenContent := tokenMap[tokenString]
	tokenizedContent, err := g.expand(tokenContent, tokenMap)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(TEMPLATE_DD, tokenizedContent), nil
}

func (g *HtmlGenerator) expandRefDefinition(tokenString string, tokenMap map[string]string) (string, error) {
	tokenContent := tokenMap[tokenString]
	splittedToken := strings.SplitN(tokenContent, " ", 2)
	refIndex, err := strconv.Atoi(splittedToken[0])
	if err != nil {
		return "", err
	}

	tokenizedContent, err := g.expand(splittedToken[1], tokenMap)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(TEMPLATE_REF_DEF, refIndex, tokenizedContent), nil
}

func (g *HtmlGenerator) expandRefUsage(tokenString string, tokenMap map[string]string) (string, error) {
	tokenContent := tokenMap[tokenString]
	splittedToken := strings.SplitN(tokenContent, " ", 2)
	refIndex, err := strconv.Atoi(splittedToken[0])
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(TEMPLATE_REF_USAGE, refIndex), nil
}

// TODO Create service class with public interface for the api functions (like RenderMath) to be able to mock that service.
func (g *HtmlGenerator) expandMath(tokenString string, tokenMap map[string]string) (string, error) {
	svgFilename, pngFilename, err := api.RenderMath(tokenMap[tokenString], g.imageCacheFolder, g.mathCacheFolder)
	if err != nil {
		return "", err
	}

	svg, err := util.ReadSvg(svgFilename)
	if err != nil {
		return "", err
	}

	sigolo.Debug("File: %s, Width: %s, Height: %s, Style: %s", pngFilename, svg.Width, svg.Height, svg.Style)

	return fmt.Sprintf(MATH_TEMPLATE, pngFilename, svg.Width, svg.Height, svg.Style), nil
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
	sigolo.Info("Write to %s", outputFilepath)
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
