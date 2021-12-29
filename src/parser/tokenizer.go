package parser

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
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

var tokenCounter = 0

func getToken(tokenType string) string {
	token := fmt.Sprintf(TOKEN_TEMPLATE, tokenType, tokenCounter)
	tokenCounter++
	return token
}

// https://www.mediawiki.org/wiki/Markup_spec
func tokenize(content string, tokenMap map[string]string) string {
	content = parseBoldAndItalic(content, tokenMap)
	content = parseHeadings(content, tokenMap)
	content = parseReferences(content, tokenMap)

	for {
		originalContent := content

		content = clean(content)
		content = evaluateTemplates(content)
		content = escapeImages(content)

		content = parseInternalLinks(content, tokenMap)
		content = parseImages(content, tokenMap)
		content = parseExternalLinks(content, tokenMap)
		content = parseTables(content, tokenMap)
		content = parseLists(content, tokenMap)
		content = parseMath(content, tokenMap)

		if content == originalContent {
			break
		}
	}

	return content
}

// tokenizeInline is meant for strings that are known to be inline string. Example: The text of an internal link cannot
// contain tables and lists, so we do not want to parse them.
func tokenizeInline(content string, tokenMap map[string]string) string {
	content = parseBoldAndItalic(content, tokenMap)
	content = parseHeadings(content, tokenMap)
	content = parseReferences(content, tokenMap)

	for {
		content = parseInternalLinks(content, tokenMap)
		content = parseImages(content, tokenMap)
		content = parseExternalLinks(content, tokenMap)
		break
	}

	return content
}

func parseHeadings(content string, tokenMap map[string]string) string {
	for i := 1; i < 7; i++ {
		headingPrefixSuffix := strings.Repeat("=", i)
		matches := regexp.MustCompile(`(?m)^`+headingPrefixSuffix+` (.*) `+headingPrefixSuffix+`$`).FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			token := getToken(fmt.Sprintf(TOKEN_HEADING_TEMPLATE, i))
			tokenMap[token] = match[1]
			content = strings.Replace(content, match[0], token, 1)
		}
	}

	return content
}

func parseBoldAndItalic(content string, tokenMap map[string]string) string {
	content, _, _, _ = tokenizeBoldAndItalic(content, 0, tokenMap, false, false)
	return content
}

func tokenizeBoldAndItalic(content string, index int, tokenMap map[string]string, isBoldOpen bool, isItalicOpen bool) (string, int, bool, bool) {
	for index < len(content) {
		// iIn case of last opened italic marker
		if index+3 <= len(content) && content[index:index+3] == "'''" {
			if !isBoldOpen {
				// -3 +3 to replace the ''' as well
				content = strings.Replace(content, content[index:index+3], MARKER_BOLD_OPEN, 1)
				index = index + len(MARKER_BOLD_OPEN)

				// Check for further nested italic tags
				content, index, isBoldOpen, isItalicOpen = tokenizeBoldAndItalic(content, index, tokenMap, true, isItalicOpen)
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
				content, index, isBoldOpen, isItalicOpen = tokenizeBoldAndItalic(content, index, tokenMap, isBoldOpen, true)
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

func parseImages(content string, tokenMap map[string]string) string {
	regex := regexp.MustCompile(IMAGE_REGEX)
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		filename := submatch[3]
		filenameToken := getToken(TOKEN_IMAGE_FILENAME)
		tokenMap[filenameToken] = filename

		tokenString := TOKEN_IMAGE_INLINE
		imageSizeToken := ""
		captionToken := ""
		if len(submatch) >= 4 {
			options := strings.Split(submatch[5], "|")
			ignorePrefixes := []string{
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

			nonInlinePrefix := []string{
				"mini",
				"thumb",
			}

			for i, option := range options {
				if strings.HasSuffix(option, "px") {
					option = strings.TrimSuffix(option, "px")
					sizes := strings.Split(option, "x")

					xSize := sizes[0]
					ySize := xSize
					if len(sizes) == 2 {
						ySize = sizes[1]
					}

					imageSizeString := fmt.Sprintf("%sx%s", xSize, ySize)
					imageSizeToken = getToken(TOKEN_IMAGE_SIZE)
					tokenMap[imageSizeToken] = imageSizeString
				} else if elemetHasPrefix(option, ignorePrefixes) {
					continue
				} else if elemetHasPrefix(option, nonInlinePrefix) {
					tokenString = TOKEN_IMAGE
				} else if i == len(options)-1 && tokenString == TOKEN_IMAGE {
					// last remaining option is caption as long as we do NOT have an inline image
					captionToken = getToken(TOKEN_IMAGE_CAPTION)
					tokenMap[captionToken] = option
				}
			}
		}

		token := getToken(tokenString)
		tokenMap[token] = filenameToken + " " + captionToken + " " + imageSizeToken

		content = strings.Replace(content, submatch[0], token, 1)
	}

	return content
}

func elemetHasPrefix(element string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(element, prefix) {
			return true
		}
	}
	return false
}

func parseInternalLinks(content string, tokenMap map[string]string) string {
	regex := regexp.MustCompile(`\[\[([^|^\]^\[]*)(\|([^|^\]^\[]*))?]]`)
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		if strings.HasPrefix(submatch[0], "[[Datei:") || strings.HasPrefix(submatch[1], "[[File:") {
			continue
		}

		tokenArticle := getToken(TOKEN_INTERNAL_LINK_ARTICLE)
		tokenMap[tokenArticle] = submatch[1]

		linkText := submatch[1]
		if submatch[3] != "" {
			// Use article as text
			linkText = submatch[3]
		}

		text := tokenizeInline(linkText, tokenMap)
		tokenText := getToken(TOKEN_INTERNAL_LINK_TEXT)
		tokenMap[tokenText] = text

		token := getToken(TOKEN_INTERNAL_LINK)
		tokenMap[token] = tokenArticle + " " + tokenText

		content = strings.Replace(content, submatch[0], token, 1)
	}

	return content
}

func parseExternalLinks(content string, tokenMap map[string]string) string {
	regex := regexp.MustCompile(`([^\[])\[(http[^]]*?)( ([^]]*?))?]([^]])`)
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		tokenUrl := getToken(TOKEN_EXTERNAL_LINK_URL)
		tokenMap[tokenUrl] = submatch[2]

		linkText := submatch[2]
		if len(submatch) >= 5 {
			linkText = submatch[4]
		}

		linkText = tokenize(linkText, tokenMap)
		tokenText := getToken(TOKEN_EXTERNAL_LINK_TEXT)
		tokenMap[tokenText] = linkText

		token := getToken(TOKEN_EXTERNAL_LINK)
		tokenMap[token] = tokenUrl + " " + tokenText

		// Remove last characters as it's the first character after the closing  ]]  of the file tag.
		totalMatch := submatch[0][:len(submatch[0])-1]
		content = strings.Replace(content, totalMatch, submatch[1]+token, 1)
	}

	return content
}

func parseTables(content string, tokenMap map[string]string) string {
	lines := strings.Split(content, "\n")
	regex := regexp.MustCompile(`^(:*)(\{\|.*)`)

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if regex.MatchString(line) {
			submatch := regex.FindStringSubmatch(line)
			listPrefix := submatch[1]
			line = submatch[2]

			// table starts in this line.
			token, newIndex := tokenizeTables(lines, i, tokenMap)

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
func tokenizeTables(lines []string, i int, tokenMap map[string]string) (string, int) {
	tableLines := []string{}
	tableLines = append(tableLines, lines[i])
	i++

	// collect all lines from this table
	for ; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(line, "{|") || strings.HasPrefix(line, ":{|") {
			// another table starts
			tableToken := ""
			tableToken, i = tokenizeTables(lines, i, tokenMap)
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
	token := tokenizeTable(tableContent, tokenMap)
	return token, i
}

// tokenizeTable expects content to be all lines of a table.
func tokenizeTable(content string, tokenMap map[string]string) string {
	// ensure that each columns starts in a new row
	content = strings.ReplaceAll(content, "||", "\n|")
	content = strings.ReplaceAll(content, "!!", "\n!")
	lines := strings.Split(content, "\n")

	tableTokens := ""
	captionToken := ""

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "|+") {
			caption := strings.TrimPrefix(line, "|+")
			captionToken = getToken(TOKEN_TABLE_CAPTION)
			tokenMap[captionToken] = caption
			tableTokens += captionToken + " "
		} else if strings.HasPrefix(line, "|-") {
			rowToken := ""
			if strings.HasPrefix(lines[i+1], "!") {
				// this table row is a heading
				rowToken, i = tokenizeTableRow(lines, i+1, "!", tokenMap)
			} else if strings.HasPrefix(lines[i+1], "|") {
				// this table row is a normal row
				rowToken, i = tokenizeTableRow(lines, i+1, "|", tokenMap)
			}

			tableTokens += rowToken + " "
		} else if strings.HasPrefix(line, "|}") {
			// table ends with this line
			break
		}
	}

	token := getToken(TOKEN_TABLE)
	tokenMap[token] = tableTokens

	return token
}

// tokenizeTableRow expects i to be the line with the first text item (i.e. the line after |- ). The returned index
// points to the last text line of this table row.
func tokenizeTableRow(lines []string, i int, sep string, tokenMap map[string]string) (string, int) {
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
		line, attributes = tokenizeTableColumn(line, tokenMap)

		attributeToken := ""
		if attributes != "" {
			attributeToken = getToken(TOKEN_TABLE_COL_ATTRIBUTES)
			tokenMap[attributeToken] = attributes
		}

		token := ""
		if sep == "!" {
			token = getToken(TOKEN_TABLE_HEAD)
		} else {
			token = getToken(TOKEN_TABLE_COL)
		}
		tokenMap[token] = attributeToken + line

		rowLines = append(rowLines, token)
	}

	tokenContent := strings.Join(rowLines, " ")

	token := getToken(TOKEN_TABLE_ROW)
	tokenMap[token] = tokenContent

	// return i-1 so that i is on the last line of the row when returning
	return token, i - 1
}

// tokenizeTableColumn returns the tokenized text of the column and as second return value relevant CSS attributes (might be empty).
func tokenizeTableColumn(content string, tokenMap map[string]string) (string, string) {
	splittedContent := strings.Split(content, "|")
	if len(splittedContent) < 2 {
		return tokenize(content, tokenMap), ""
	}

	attributeString := splittedContent[0]
	columnText := tokenize(splittedContent[1], tokenMap)

	relevantTags := []string{}

	colspanMatch := regexp.MustCompile(`colspan="(\d+)"`).FindStringSubmatch(attributeString)
	if len(colspanMatch) > 1 {
		relevantTags = append(relevantTags, colspanMatch[0])
	}

	alignmentMatch := regexp.MustCompile(`text-align=".*?";`).FindStringSubmatch(attributeString)
	if len(alignmentMatch) > 1 {
		relevantTags = append(relevantTags, `style="`+alignmentMatch[0]+`"`)
	}

	return columnText, strings.Join(relevantTags, " ")
}

func parseLists(content string, tokenMap map[string]string) string {
	lines := strings.Split(content, "\n")

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		regex := regexp.MustCompile(`^([*#:])`)
		lineStartCharacter := regex.FindStringSubmatch(line)

		if len(lineStartCharacter) > 0 && lineStartCharacter[1] != "" {
			listTokenString := getListTokenString(lineStartCharacter[1])

			// a new unordered list starts here
			token, newIndex := tokenizeList(lines, i, tokenMap, lineStartCharacter[1], listTokenString)

			length := newIndex - i

			newLines := []string{}
			newLines = append(newLines, lines[:i]...)
			newLines = append(newLines, token)
			if i+length+1 < len(lines) {
				newLines = append(newLines, lines[i+length:]...)
			}
			lines = newLines
		}
	}

	content = strings.Join(lines, "\n")
	return content
}

func tokenizeList(lines []string, i int, tokenMap map[string]string, itemPrefix string, tokenString string) (string, int) {
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

	regex := regexp.MustCompile(`(^|\n)[` + itemPrefix + `]([^` + itemPrefix + `])`)
	submatches := regex.FindAllStringSubmatch(content, -1)
	// Ignore first item as it's always empty
	// Each element contains the whole item including all sub-lists and everything
	completeListItems := regex.Split(content, -1)[1:]

	for itemIndex, item := range completeListItems {
		// re-add the non-prefix characters which was removed by .Split() above
		token := tokenizeListItem(submatches[itemIndex][2]+item, tokenMap, itemPrefix)
		completeListItems[itemIndex] = token
	}

	tokenContent := strings.Join(completeListItems, " ")
	token := getToken(tokenString)
	tokenMap[token] = tokenContent

	return token, i
}

func tokenizeListItem(content string, tokenMap map[string]string, itemPrefix string) string {
	content = strings.TrimPrefix(content, itemPrefix+" ")
	lines := strings.Split(content, "\n")

	itemContent := ""
	subListString := ""

	// collect all lines of this list item which do not belong to a nested item
	for i, line := range lines {
		if hasListItemPrefix(line) {
			// a sub-item starts
			subListString = strings.Join(lines[i:], "\n")
			break
		}
		itemContent += line + "\n"
	}

	token := getToken(getListItemTokenString(itemPrefix))
	tokenMap[token] = tokenize(itemContent, tokenMap)

	if subListString != "" {
		subItemPrefix := subListString[0:1]
		regex := regexp.MustCompile(`(^|\n)[` + subItemPrefix + `]`)

		subListString = regex.ReplaceAllString(subListString, "\n")
		// Ignore first item as it's always empty (due to newline from replacement)
		subListItemLines := strings.Split(subListString, "\n")

		subListToken, _ := tokenizeList(subListItemLines, 0, tokenMap, subItemPrefix, getListTokenString(subItemPrefix))
		tokenMap[token] += " " + subListToken
	}

	return token
}

func getListTokenString(listItemPrefix string) string {
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

func getListItemTokenString(listItemPrefix string) string {
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

func hasListItemPrefix(line string) bool {
	return regexp.MustCompile(`^[*#:]`).MatchString(line)
}

func parseReferences(content string, tokenMap map[string]string) string {
	referenceDefinitions := map[string]string{}
	referenceUsages := map[string]string{}

	// These map take the index of the reference in "content" as determined by  strings.Index()  as key/value.
	contentIndexToRefName := map[int]string{}

	// Maps the reference name to the actual index starting at 1 as they will appear in the generated result later on.
	refNameToIndex := map[string]int{}

	// Split the content into section before, at and after the reference list
	regex := regexp.MustCompile(`</?references.*?/?>\n?`)
	if !regex.MatchString(content) {
		return content
	}
	contentParts := regex.Split(content, -1)
	// In case of dedicated <references>...</references> block
	//   part 0: everything before <references...>
	//   part 1: everything between <references> and </references>
	//   part 2: everything after </references>
	// In case of <references/>
	//   part 0: everything before <references/>
	//   part 1: everything after <references/>
	// Completely remove the reference section as we already parsed it above with the regex.
	head := contentParts[0]
	foot := ""
	if len(contentParts) == 2 {
		foot = contentParts[1]
	} else if len(contentParts) == 3 {
		foot = contentParts[2]
	}

	// For definition <ref name="...">...</ref>   -> create usage with the given name
	regex = regexp.MustCompile(`<ref name="?([^"^>]*?)"?>((.|\n)*?)</ref>`)
	// Go throught "content" to also parse the definitions in the reference section below "head"
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		name := submatch[1]
		totalRef := submatch[0]
		referenceDefinitions[name] = totalRef
		head = strings.ReplaceAll(head, totalRef, fmt.Sprintf("<ref name=\"%s\" />", name))
	}

	// For definition <ref>...</ref>   -> create usage with a new random name
	regex = regexp.MustCompile(`<ref>((.|\n)*?)</ref>`)
	// Go throught "content" to also parse the definitions in the reference section below "head"
	submatches = regex.FindAllStringSubmatch(content, -1)
	for i, submatch := range submatches {
		totalRef := submatch[0]

		// Generate more or less random but unique name
		hash := sha1.New()
		hash.Write([]byte(fmt.Sprintf("%d", i)))
		hash.Write([]byte(totalRef))
		name := hex.EncodeToString(hash.Sum(nil))

		referenceDefinitions[name] = totalRef
		head = strings.ReplaceAll(head, totalRef, fmt.Sprintf("<ref name=\"%s\" />", name))
	}

	// For usage <ref name="..." />
	regex = regexp.MustCompile(`<ref name="([^"]*?)" ?/>`)
	submatches = regex.FindAllStringSubmatch(head, -1)
	for _, submatch := range submatches {
		name := submatch[1]
		referenceUsages[name] = submatch[0]
		contentIndexToRefName[strings.Index(head, submatch[0])] = name
	}

	contentIndexToReferenceKeys := make([]int, 0, len(contentIndexToRefName))
	for key := range contentIndexToRefName {
		contentIndexToReferenceKeys = append(contentIndexToReferenceKeys, key)
	}
	sort.Ints(contentIndexToReferenceKeys)

	// Assign increasing index to each reference based on their occurrence in "content"
	refCounter := 1
	sortedRefNames := []string{}
	for _, refContentIndex := range contentIndexToReferenceKeys {
		refName := contentIndexToRefName[refContentIndex]
		refNameToIndex[refName] = refCounter
		sortedRefNames = append(sortedRefNames, refName)
		refCounter++
	}

	// Create usage token for ref usages like <ref name="foo" />
	for name, ref := range referenceUsages {
		refIndex := refNameToIndex[name]
		token := getToken(TOKEN_REF_USAGE)
		tokenMap[token] = fmt.Sprintf("%d %s", refIndex, ref)
		head = strings.ReplaceAll(head, ref, token)
	}

	// Append ref definitions to head
	for _, name := range sortedRefNames {
		ref := referenceDefinitions[name]
		token := getToken(TOKEN_REF_DEF)
		tokenMap[token] = fmt.Sprintf("%d %s", refNameToIndex[name], tokenize(ref, tokenMap))
		head += token + "\n"
	}

	return head + foot
}

func parseMath(content string, tokenMap map[string]string) string {
	regex := regexp.MustCompile(`<math>(.*?)</math>`)
	matches := regex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		token := getToken(TOKEN_MATH)
		tokenMap[token] = match[1]
		content = strings.Replace(content, match[0], token, 1)
	}
	return content
}
