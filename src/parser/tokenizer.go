package parser

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const TOKEN_REGEX = `\$\$TOKEN_([A-Z_0-9]*)_\d+\$\$`
const TOKEN_TEMPLATE = "$$TOKEN_%s_%d$$"

const TOKEN_HEADING_TEMPLATE = "HEADING_%d"
const TOKEN_HEADING_1 = "HEADING_1"
const TOKEN_HEADING_2 = "HEADING_2"
const TOKEN_HEADING_3 = "HEADING_3"
const TOKEN_HEADING_4 = "HEADING_4"
const TOKEN_HEADING_5 = "HEADING_5"
const TOKEN_HEADING_6 = "HEADING_6"

const TOKEN_INTERNAL_LINK = "INTERNAL_LINK"
const TOKEN_INTERNAL_LINK_ARTICLE = "INTERNAL_LINK_ARTICLE"
const TOKEN_INTERNAL_LINK_TEXT = "INTERNAL_LINK_TEXT"

const TOKEN_EXTERNAL_LINK = "EXTERNAL_LINK"
const TOKEN_EXTERNAL_LINK_URL = "EXTERNAL_LINK_URL"
const TOKEN_EXTERNAL_LINK_TEXT = "EXTERNAL_LINK_TEXT"

const TOKEN_TABLE = "TABLE"
const TOKEN_TABLE_CAPTION = "TABLE_CAPTION"
const TOKEN_TABLE_HEAD = "TABLE_HEAD"
const TOKEN_TABLE_ROW = "TABLE_ROW"
const TOKEN_TABLE_COL = "TABLE_COL"
const TOKEN_TABLE_COL_ATTRIBUTES = "TABLE_COL_ATTRIB"

const TOKEN_UNORDERED_LIST = "UNORDERED_LIST"
const TOKEN_ORDERED_LIST = "ORDERED_LIST"
const TOKEN_LIST_ITEM = "LIST_ITEM"

const TOKEN_DESCRIPTION_LIST = "DESCRIPTION_LIST"
const TOKEN_DESCRIPTION_LIST_ITEM = "DESCRIPTION_LIST_ITEM"

const TOKEN_IMAGE = "IMAGE"
const TOKEN_IMAGE_INLINE = "IMAGE_INLINE"
const TOKEN_IMAGE_FILENAME = "IMAGE_FILENAME"
const TOKEN_IMAGE_CAPTION = "IMAGE_CAPTION"
const TOKEN_IMAGE_SIZE = "IMAGE_SIZE"

const TOKEN_REF_USAGE = "REF_USAGE"
const TOKEN_REF_DEF = "REF_DEF"

const TOKEN_MATH = "REF_MATH"

// Marker do not appear in the token map. A marker does not contain further information, it just marks e.g. the start
// and end of a primitive block of content (like a block of bold text)
const MARKER_BOLD_OPEN = "$$MARKER_BOLD_OPEN$$"
const MARKER_BOLD_CLOSE = "$$MARKER_BOLD_CLOSE$$"
const MARKER_ITALIC_OPEN = "$$MARKER_ITALIC_OPEN$$"
const MARKER_ITALIC_CLOSE = "$$MARKER_ITALIC_CLOSE$$"
const MARKER_PARAGRAPH = "$$MARKER_PARAGRAPH$$"

var imageIgnoreParameters = []string{
	"left",
	"right",
	"top",
	"text-top",
	"bottom",
	"text-bottom",
	"center",
	"none",
	"upright",
	"baseline",
	"sub",
	"super",
	"middle",
	"link",
	"alt",
	"page",
	"class",
	"lang",
	"zentriert",
}
var imageNonInlineParameters = []string{
	"mini",
	"thumb",
}

type ITokenizer interface {
	tokenize(content string) string
	getTokenMap() map[string]string
}

type Tokenizer struct {
	tokenMap       map[string]string
	tokenCounter   int
	imageFolder    string
	templateFolder string

	tokenizeContent func(tokenizer *Tokenizer, content string) string
}

func NewTokenizer(imageFolder string, templateFolder string) Tokenizer {
	return Tokenizer{
		tokenMap:       map[string]string{},
		tokenCounter:   0,
		imageFolder:    imageFolder,
		templateFolder: templateFolder,

		tokenizeContent: tokenizeContent,
	}
}

func (t *Tokenizer) getTokenMap() map[string]string {
	return t.tokenMap
}

func (t *Tokenizer) getToken(tokenType string) string {
	token := fmt.Sprintf(TOKEN_TEMPLATE, tokenType, t.tokenCounter)
	t.tokenCounter++
	return token
}

func (t *Tokenizer) setToken(key string, tokenContent string) {
	t.setRawToken(key, t.tokenize(tokenContent))
}

func (t *Tokenizer) setRawToken(key string, tokenContent string) {
	t.tokenMap[key] = tokenContent
}

// https://www.mediawiki.org/wiki/Markup_spec
func (t *Tokenizer) tokenize(content string) string {
	return t.tokenizeContent(t, content)
}

func tokenizeContent(t *Tokenizer, content string) string {
	content = clean(content)
	content = t.parseBoldAndItalic(content)
	content = t.parseHeadings(content)
	content = t.parseReferences(content)

	for {
		originalContent := content

		content = evaluateTemplates(content, t.templateFolder)
		content = clean(content)
		content = escapeImages(content)

		content = t.parseInternalLinks(content)
		content = t.parseImages(content)
		content = t.parseExternalLinks(content)
		content = t.parseMath(content)
		//content = t.parseParagraphs(content)
		content = t.parseTables(content)
		content = t.parseLists(content)

		if content == originalContent {
			break
		}
	}

	return content
}

// tokenizeInline is meant for strings that are known to be inline string. Example: The text of an internal link cannot
// contain tables and lists, so we do not want to parse them.
func (t *Tokenizer) tokenizeInline(content string) string {
	content = t.parseBoldAndItalic(content)
	content = t.parseHeadings(content)
	content = t.parseReferences(content)

	for {
		content = t.parseInternalLinks(content)
		content = t.parseImages(content)
		content = t.parseExternalLinks(content)
		break
	}

	return content
}

func (t *Tokenizer) parseHeadings(content string) string {
	for i := 1; i < 7; i++ {
		headingPrefixSuffix := strings.Repeat("=", i)
		matches := regexp.MustCompile(`(?m)^`+headingPrefixSuffix+` (.*) `+headingPrefixSuffix+`$`).FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			token := t.getToken(fmt.Sprintf(TOKEN_HEADING_TEMPLATE, i))
			t.setToken(token, match[1])
			content = strings.Replace(content, match[0], token, 1)
		}
	}

	return content
}

func (t *Tokenizer) parseBoldAndItalic(content string) string {
	content, _, _, _ = t.tokenizeBoldAndItalic(content, 0, false, false)
	return content
}

func (t *Tokenizer) tokenizeBoldAndItalic(content string, index int, isBoldOpen bool, isItalicOpen bool) (string, int, bool, bool) {
	for index < len(content) {
		// In case of last opened italic marker
		if index+3 <= len(content) && content[index:index+3] == "'''" {
			if !isBoldOpen {
				// -3 +3 to replace the ''' as well
				content = strings.Replace(content, content[index:index+3], MARKER_BOLD_OPEN, 1)
				index = index + len(MARKER_BOLD_OPEN)

				// Check for further nested italic tags
				content, index, isBoldOpen, isItalicOpen = t.tokenizeBoldAndItalic(content, index, true, isItalicOpen)
			} else {
				// +3 to replace the '''
				content = strings.Replace(content, content[index:index+3], MARKER_BOLD_CLOSE, 1)

				// -3 because of the ''' we replaced above
				return content, index + len(MARKER_BOLD_CLOSE), false, isItalicOpen
			}
		} else if index+2 <= len(content) && content[index:index+2] == "''" {
			if !isItalicOpen {
				// +2 to replace the ''
				content = strings.Replace(content, content[index:index+2], MARKER_ITALIC_OPEN, 1)
				index = index + len(MARKER_ITALIC_OPEN)

				// Check for further nested italic tags
				content, index, isBoldOpen, isItalicOpen = t.tokenizeBoldAndItalic(content, index, isBoldOpen, true)
			} else {
				// +2 to replace the ''
				content = strings.Replace(content, content[index:index+2], MARKER_ITALIC_CLOSE, 1)

				// -2 because of the '' we replaced above
				return content, index + len(MARKER_ITALIC_CLOSE), isBoldOpen, false
			}
		} else {
			index++
		}
	}

	return content, index, false, false
}

func (t *Tokenizer) parseImages(content string) string {
	regex := regexp.MustCompile(IMAGE_REGEX)
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		filename := filepath.Join(t.imageFolder, submatch[3])
		filenameToken := t.getToken(TOKEN_IMAGE_FILENAME)
		t.setToken(filenameToken, filename)

		tokenString := TOKEN_IMAGE_INLINE
		imageSizeToken := ""
		captionToken := ""
		if len(submatch) >= 4 {
			options := strings.Split(submatch[5], "|")

			for i, option := range options {
				if t.elementHasPrefix(option, imageNonInlineParameters) {
					tokenString = TOKEN_IMAGE
				} else if t.elementHasPrefix(option, imageIgnoreParameters) {
					continue
				} else if strings.HasSuffix(option, "px") && tokenString != TOKEN_IMAGE {
					option = strings.TrimSuffix(option, "px")
					sizes := strings.Split(option, "x")

					xSize := sizes[0]
					ySize := xSize
					if len(sizes) == 2 {
						ySize = sizes[1]
					}

					imageSizeString := fmt.Sprintf("%sx%s", xSize, ySize)
					imageSizeToken = t.getToken(TOKEN_IMAGE_SIZE)
					t.setToken(imageSizeToken, imageSizeString)
				} else if i == len(options)-1 && tokenString == TOKEN_IMAGE {
					// last remaining option is caption as long as we do NOT have an inline image
					captionToken = t.getToken(TOKEN_IMAGE_CAPTION)
					t.setToken(captionToken, option)
				}
			}
		}

		token := t.getToken(tokenString)
		t.setToken(token, filenameToken+" "+captionToken+" "+imageSizeToken)

		content = strings.Replace(content, submatch[0], token, 1)
	}

	return content
}

func (t *Tokenizer) elementHasPrefix(element string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(element, prefix) {
			return true
		}
	}
	return false
}

func (t *Tokenizer) parseInternalLinks(content string) string {
	regex := regexp.MustCompile(`\[\[([^|^\]^\[]*)(\|([^|^\]^\[]*))?]]`)
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		if strings.HasPrefix(submatch[0], "[[Datei:") || strings.HasPrefix(submatch[1], "[[File:") {
			continue
		}

		tokenArticle := t.getToken(TOKEN_INTERNAL_LINK_ARTICLE)
		t.setToken(tokenArticle, submatch[1])

		linkText := submatch[1]
		if submatch[3] != "" {
			// Use article as text
			linkText = submatch[3]
		}

		text := t.tokenizeInline(linkText)
		tokenText := t.getToken(TOKEN_INTERNAL_LINK_TEXT)
		t.setToken(tokenText, text)

		token := t.getToken(TOKEN_INTERNAL_LINK)
		t.setToken(token, tokenArticle+" "+tokenText)

		content = strings.Replace(content, submatch[0], token, 1)
	}

	return content
}

func (t *Tokenizer) parseExternalLinks(content string) string {
	regex := regexp.MustCompile(`([^\[])\[(http[^]]*?)( ([^]]*?))?](([^]])|$)`)
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		tokenUrl := t.getToken(TOKEN_EXTERNAL_LINK_URL)
		t.setToken(tokenUrl, submatch[2])

		linkText := submatch[2]
		if len(submatch) >= 5 {
			linkText = submatch[4]
		}

		tokenText := t.getToken(TOKEN_EXTERNAL_LINK_TEXT)
		t.setToken(tokenText, linkText)

		token := t.getToken(TOKEN_EXTERNAL_LINK)
		t.setToken(token, tokenUrl+" "+tokenText)

		// Remove last characters as it's the first character after the closing  ]  of the file tag.
		totalMatch := submatch[0]
		if totalMatch[len(totalMatch)-1] != ']' {
			totalMatch = totalMatch[:len(totalMatch)-1]
		}
		content = strings.Replace(content, totalMatch, submatch[1]+token, 1)
	}

	return content
}

func (t *Tokenizer) parseTables(content string) string {
	lines := strings.Split(content, "\n")
	regex := regexp.MustCompile(`^(:*)(\{\|.*)`)

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if regex.MatchString(line) {
			submatch := regex.FindStringSubmatch(line)
			listPrefix := submatch[1]
			line = submatch[2]

			// table starts in this line.
			token, newIndex := t.tokenizeTables(lines, i)

			length := newIndex - i

			newLines := []string{}
			newLines = append(newLines, lines[:i]...)
			newLines = append(newLines, listPrefix+token)
			if i+length+1 < len(lines) {
				newLines = append(newLines, lines[i+length+1:]...)
			}
			lines = newLines

		}
	}

	content = strings.Join(lines, "\n")
	return content
}

// tokenizeTable returns the token of the table and the index of the row where this table ended.
func (t *Tokenizer) tokenizeTables(lines []string, i int) (string, int) {
	tableLines := []string{}
	tableLines = append(tableLines, lines[i])
	i++

	// collect all lines from this table
	for ; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(line, "{|") || strings.HasPrefix(line, ":{|") {
			// another table starts
			tableToken := ""
			tableToken, i = t.tokenizeTables(lines, i)
			tableLines = append(tableLines, tableToken)
		} else if strings.HasPrefix(line, "|}") {
			// the table ends with this line
			tableLines = append(tableLines, lines[i])
			break
		} else {
			tableLines = append(tableLines, line)
		}
	}

	tableContent := strings.Join(tableLines, "\n")
	token := t.tokenizeTable(tableContent)
	return token, i
}

// tokenizeTable expects content to be all lines of a table.
func (t *Tokenizer) tokenizeTable(content string) string {
	// ensure that each columns starts in a new row
	content = strings.ReplaceAll(content, "||", "\n|")
	content = strings.ReplaceAll(content, "!!", "\n!")
	lines := strings.Split(content, "\n")

	tableTokens := []string{}
	captionToken := ""

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "|+") {
			caption := strings.TrimPrefix(line, "|+")
			captionToken = t.getToken(TOKEN_TABLE_CAPTION)
			t.setToken(captionToken, caption)
			tableTokens = append(tableTokens, captionToken)
		} else if strings.HasPrefix(line, "|-") {
			rowToken := ""
			if strings.HasPrefix(lines[i+1], "!") {
				// this table row is a heading
				rowToken, i = t.tokenizeTableRow(lines, i+1, "!")
			} else if strings.HasPrefix(lines[i+1], "|") {
				// this table row is a normal row
				rowToken, i = t.tokenizeTableRow(lines, i+1, "|")
			}

			tableTokens = append(tableTokens, rowToken)
		} else if strings.HasPrefix(line, "|}") {
			// table ends with this line
			break
		}
	}

	token := t.getToken(TOKEN_TABLE)
	t.setToken(token, strings.Join(tableTokens, " "))

	return token
}

// tokenizeTableRow expects i to be the line with the first text item (i.e. the line after |- ). Furthermore this
// function expects that each column starts in a new line starting with "| " (or whatever "sep" is). The returned index
// points to the last text line of this table row.
func (t *Tokenizer) tokenizeTableRow(lines []string, i int, sep string) (string, int) {
	rowLines := []string{}

	// collect all lines from this row
	for ; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(line, "|-") || strings.HasPrefix(line, "|}") {
			// table row ended
			break
		}

		line = strings.TrimPrefix(line, sep)

		// one column may consist of multiple text rows -> all text lines until the next column or row and tokenize them
		i++
		for ; !strings.HasPrefix(lines[i], sep) && !strings.HasPrefix(lines[i], "|"); i++ {
			line += "\n" + lines[i]
			fmt.Println(lines[i])
		}
		// now the index is at the start of the next column/row -> reduce by 1 for later parsing
		i -= 1

		attributes := ""
		line, attributes = t.tokenizeTableColumn(line)

		attributeToken := ""
		if attributes != "" {
			attributeToken = t.getToken(TOKEN_TABLE_COL_ATTRIBUTES)
			t.setToken(attributeToken, attributes)
		}

		token := ""
		if sep == "!" {
			token = t.getToken(TOKEN_TABLE_HEAD)
		} else {
			token = t.getToken(TOKEN_TABLE_COL)
		}
		t.setToken(token, attributeToken+line)

		rowLines = append(rowLines, token)
	}

	tokenContent := strings.Join(rowLines, " ")

	token := t.getToken(TOKEN_TABLE_ROW)
	t.setToken(token, tokenContent)

	// return i-1 so that i is on the last line of the row when returning
	return token, i - 1
}

// tokenizeTableColumn returns the tokenized text of the column and as second return value relevant CSS attributes (might be empty).
func (t *Tokenizer) tokenizeTableColumn(content string) (string, string) {
	splittedContent := strings.Split(content, "|")
	if len(splittedContent) < 2 {
		return t.tokenize(content), ""
	}

	attributeString := splittedContent[0]
	columnText := t.tokenize(splittedContent[1])

	relevantTags := []string{}

	colspanMatch := regexp.MustCompile(`colspan="(\d+)"`).FindStringSubmatch(attributeString)
	if len(colspanMatch) > 1 {
		relevantTags = append(relevantTags, colspanMatch[0])
	}

	alignmentMatch := regexp.MustCompile(`text-align:.+?;`).FindStringSubmatch(attributeString)
	if len(alignmentMatch) > 0 {
		relevantTags = append(relevantTags, `style="`+alignmentMatch[0]+`"`)
	}

	return columnText, strings.Join(relevantTags, " ")
}

func (t *Tokenizer) parseLists(content string) string {
	lines := strings.Split(content, "\n")

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		regex := regexp.MustCompile(`^([*#:])`)
		lineStartCharacter := regex.FindStringSubmatch(line)

		if len(lineStartCharacter) > 0 && lineStartCharacter[1] != "" {
			listTokenString := t.getListTokenString(lineStartCharacter[1])

			// a new list starts here
			token, newIndex := t.tokenizeList(lines, i, lineStartCharacter[1], listTokenString)

			length := newIndex - i

			newLines := []string{}
			newLines = append(newLines, lines[:i]...)
			newLines = append(newLines, token)
			if i+length < len(lines) {
				newLines = append(newLines, lines[i+length:]...)
			}
			lines = newLines
		}
	}

	content = strings.Join(lines, "\n")
	return content
}

func (t *Tokenizer) tokenizeList(lines []string, i int, itemPrefix string, tokenString string) (string, int) {
	listLines := []string{}
	for ; i < len(lines); i++ {
		line := lines[i]

		if !strings.HasPrefix(line, itemPrefix) && line != "" {
			break
		}

		listLines = append(listLines, line)
	}

	// lines may start directly with higher level is nesting, like "*** item one" as first item
	// To be able to parse such list beginning, we here add some dummy list items
	for {
		regex := regexp.MustCompile(`^[` + itemPrefix + `]([` + itemPrefix + `]+)`)
		submatch := regex.FindStringSubmatch(listLines[0])
		if len(submatch) == 2 {
			newLine := submatch[1] + " "
			listLines = append([]string{newLine}, listLines...)
		} else {
			break
		}
	}

	content := strings.Join(listLines, "\n")

	regex := regexp.MustCompile(`(^|\n)[` + itemPrefix + `]([^*^#^:])`)
	submatches := regex.FindAllStringSubmatch(content, -1)
	// Ignore first item as it's always empty
	// Each element contains the whole item including all sub-lists and everything
	completeListItems := regex.Split(content, -1)[1:]

	for itemIndex, item := range completeListItems {
		// re-add the non-prefix characters which was removed by .Split() above
		item = submatches[itemIndex][2] + item
		token := t.tokenizeListItem(item, itemPrefix)
		completeListItems[itemIndex] = token
	}

	tokenContent := strings.Join(completeListItems, " ")
	token := t.getToken(tokenString)
	t.setToken(token, tokenContent)

	return token, i
}

func (t *Tokenizer) tokenizeListItem(content string, itemPrefix string) string {
	content = strings.TrimPrefix(content, itemPrefix+" ")
	lines := strings.Split(content, "\n")

	itemContent := ""
	subListString := ""

	// collect all lines of this list item which do not belong to a nested item
	for i, line := range lines {
		if t.hasListItemPrefix(line) {
			// a sub-item starts
			subListString = strings.Join(lines[i:], "\n")
			break
		}
		itemContent += line + "\n"
	}

	token := t.getToken(t.getListItemTokenString(itemPrefix))
	tokenContent := t.tokenize(itemContent)

	if subListString != "" {
		regex := regexp.MustCompile(`(^|\n)[` + itemPrefix + `]`)

		subListItemLines := regex.Split(subListString, -1)[1:]
		// Ignore first item as it's always empty (due to newline from replacement)
		subItemPrefix := subListItemLines[0][0:1]

		listTokenString := t.getListTokenString(subItemPrefix)
		subListToken, _ := t.tokenizeList(subListItemLines, 0, subItemPrefix, listTokenString)
		tokenContent += " " + subListToken
	}

	t.setToken(token, tokenContent)
	return token
}

func (t *Tokenizer) getListTokenString(listItemPrefix string) string {
	switch listItemPrefix {
	case "*":
		return TOKEN_UNORDERED_LIST
	case "#":
		return TOKEN_ORDERED_LIST
	case ":":
		return TOKEN_DESCRIPTION_LIST
	}
	return ""
}

func (t *Tokenizer) getListItemTokenString(listItemPrefix string) string {
	switch listItemPrefix {
	case "*":
		return TOKEN_LIST_ITEM
	case "#":
		return TOKEN_LIST_ITEM
	case ":":
		return TOKEN_DESCRIPTION_LIST_ITEM
	}
	return ""
}

func (t *Tokenizer) hasListItemPrefix(line string) bool {
	return regexp.MustCompile(`^[*#:]`).MatchString(line)
}

func (t *Tokenizer) parseReferences(content string) string {
	/*
		Idea of this parsing step:

		0. Split the content into a head and foot part (head is above the reference list, foot below it)
		1. Collect reference definitions
			1.1. Collect all named references definitions like <ref name="foo">bar</ref> and replace them by usages like <ref name="foo" />
			1.2. Collect all unnamed definitions like <ref>foo</ref> and replace them by usages like <ref name="foo" />
		2. Get a sorted list of all such usages to assign numbers (reference counter value) to them. This number will be shown to the user, that's why it needs to be sorted.
		3. Generate tokens for reference usages and definitions. The definitions will be between head and foot.
	*/

	referenceDefinitions := map[string]string{}

	// Step 0
	head, foot, content, noRefListFound := t.getReferenceHeadAndFoot(content)
	if noRefListFound {
		return content
	}

	// Step 1
	// Step 1.1
	// For definition <ref name="...">...</ref>   -> create usage with the given name
	head = t.replaceNamedReferences(content, referenceDefinitions, head)

	// Step 1.2
	// For definition <ref>...</ref>   -> create usage with a new random name
	head = t.replaceUnnamedReferences(content, referenceDefinitions, head)

	// Step 2
	// For usage <ref name="..." />
	nameToReference, contentIndexToRefName := t.getReferenceUsages(head)

	sortedRefNames, refNameToIndex := t.getSortedReferenceNames(contentIndexToRefName)

	// Step 3
	// Create usage token for ref usages like <ref name="foo" />
	for _, name := range sortedRefNames {
		ref := nameToReference[name]
		refIndex := refNameToIndex[name]
		token := t.getToken(TOKEN_REF_USAGE)
		t.setToken(token, fmt.Sprintf("%d %s", refIndex, ref))
		head = strings.ReplaceAll(head, ref, token)
	}

	// Append ref definitions to head
	for _, name := range sortedRefNames {
		ref := referenceDefinitions[name]
		token := t.getToken(TOKEN_REF_DEF)
		t.setToken(token, fmt.Sprintf("%d %s", refNameToIndex[name], t.tokenize(ref)))
		head += token + "\n"
	}

	return head + foot
}

// getReferenceHeadAndFoot splits the content into section before, at and after the reference list.
// The return values are head, foot, content and a boolean which is true if there's no reference list in content.
func (t *Tokenizer) getReferenceHeadAndFoot(content string) (string, string, string, bool) {
	regex := regexp.MustCompile(`</?references.*?/?>\n?`)

	// No reference list found -> abort
	if !regex.MatchString(content) {
		return "", "", content, true
	}

	contentParts := regex.Split(content, -1)
	// In case of dedicated <references>...</references> block
	//   part 0 = head   : everything before <references...>
	//   part 1 (ignored): everything between <references> and </references>
	//   part 2 = foot   : everything after </references>
	// In case of <references/>
	//   part 0 = head: everything before <references/>
	//   part 1 = foot: everything after <references/>
	// Completely remove the reference section as we already parsed it above with the regex.
	head := contentParts[0]
	foot := ""
	if len(contentParts) == 2 {
		foot = contentParts[1]
	} else if len(contentParts) == 3 {
		foot = contentParts[2]
	}
	return head, foot, content, false
}

// getSortedReferenceNames gets all reference names sorted by their occurrence and a map from name to an index (the occurrence counter).
func (t *Tokenizer) getSortedReferenceNames(indexToRefName map[int]string) ([]string, map[string]int) {
	refNameToIndex := map[string]int{}

	referenceIndices := make([]int, 0, len(indexToRefName))
	for key := range indexToRefName {
		referenceIndices = append(referenceIndices, key)
	}
	sort.Ints(referenceIndices)

	// Assign increasing index to each reference based on their occurrence in "content"
	refCounter := 1
	var sortedRefNames []string
	for _, refIndex := range referenceIndices {
		refName := indexToRefName[refIndex]
		refNameToIndex[refName] = refCounter
		sortedRefNames = append(sortedRefNames, refName)
		refCounter++
	}

	return sortedRefNames, refNameToIndex
}

// replaceNamedReferences replaces all occurrences of named reference definitions in "head" by a named reference usage.
func (t *Tokenizer) replaceNamedReferences(content string, nameToRefDef map[string]string, head string) string {
	regex := regexp.MustCompile(`<ref name="?([^"^>]*?)"?>((.|\n)*?)</ref>`)
	// Go through "content" to also parse the definitions inside the <references>...</references> block and below it.
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		name := submatch[1]
		totalRef := submatch[0]
		nameToRefDef[name] = totalRef
		head = strings.ReplaceAll(head, totalRef, fmt.Sprintf("<ref name=\"%s\" />", name))
	}
	return head
}

// replaceNamedReferences replaces all occurrences of unnamed reference definitions by a named reference usage with a random reference name.
func (t *Tokenizer) replaceUnnamedReferences(content string, nameToRefDef map[string]string, head string) string {
	regex := regexp.MustCompile(`<ref>((.|\n)*?)</ref>`)
	// Go throught "content" to also parse the definitions in the reference section below "head"
	submatches := regex.FindAllStringSubmatch(content, -1)
	for i, submatch := range submatches {
		totalRef := submatch[0]

		// Generate more or less random but unique name
		hash := sha1.New()
		hash.Write([]byte(fmt.Sprintf("%d", i)))
		hash.Write([]byte(totalRef))
		name := hex.EncodeToString(hash.Sum(nil))

		nameToRefDef[name] = totalRef
		head = strings.ReplaceAll(head, totalRef, fmt.Sprintf("<ref name=\"%s\" />", name))
	}
	return head
}

// getReferenceUsages gets all reference usages (name to total reference) as well as a map that maps the reference counter to the reference name.
func (t *Tokenizer) getReferenceUsages(head string) (map[string]string, map[int]string) {
	// This map maps the reference name to the actual wikitext of that reference
	nameToRefDef := map[string]string{}
	// This map take the index of the reference in "content" as determined by  strings.Index()  as key/value.
	indexToRefName := map[int]string{}

	regex := regexp.MustCompile(`<ref name="([^"]*?)" ?/>`)
	submatches := regex.FindAllStringSubmatch(head, -1)
	for _, submatch := range submatches {
		name := submatch[1]
		nameToRefDef[name] = submatch[0]
		indexToRefName[strings.Index(head, submatch[0])] = name
	}

	return nameToRefDef, indexToRefName
}

func (t *Tokenizer) parseMath(content string) string {
	regex := regexp.MustCompile(`<math>((.|\n|\r)*?)</math>`)
	matches := regex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		token := t.getToken(TOKEN_MATH)
		t.setRawToken(token, match[1])
		content = strings.Replace(content, match[0], token, 1)
	}
	return content
}

func (t *Tokenizer) parseParagraphs(content string) string {
	return strings.ReplaceAll(content, "\n\n", "\n"+MARKER_PARAGRAPH+"\n")
}
