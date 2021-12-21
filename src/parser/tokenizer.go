package parser

import (
	"fmt"
	"regexp"
	"strings"
)

const TOKEN_REGEX = `\$\$TOKEN_([A-Z_]*)_\d*\$\$`
const TOKEN_TEMPLATE = "$$TOKEN_%s_%d$$"

const TOKEN_INTERNAL_LINK = "INTERNAL_LINK"
const TOKEN_INTERNAL_LINK_ARTICLE = "INTERNAL_LINK_ARTICLE"
const TOKEN_INTERNAL_LINK_TEXT = "INTERNAL_LINK_TEXT"

const TOKEN_EXTERNAL_LINK = "EXTERNAL_LINK"
const TOKEN_EXTERNAL_LINK_URL = "EXTERNAL_LINK_URL"
const TOKEN_EXTERNAL_LINK_TEXT = "EXTERNAL_LINK_TEXT"

const TOKEN_TABLE = "TABLE"
const TOKEN_TABLE_HEAD = "TABLE_HEAD"
const TOKEN_TABLE_ROW = "TABLE_ROW"
const TOKEN_TABLE_COL = "TABLE_COL"

const TOKEN_UNORDERED_LIST = "UNORDERED_LIST"
const TOKEN_ORDERED_LIST = "ORDERED_LIST"
const TOKEN_LIST_ITEM = "LIST_ITEM"

const TOKEN_DESCRIPTION_LIST = "DESCRIPTION_LIST"
const TOKEN_DESCRIPTION_LIST_ITEM = "DESCRIPTION_LIST_ITEM"

// Marker do not appear in the token map. A marker does not contain further information, it just marks e.g. the start
// and end of a primitive block of content (like a block of bold text)
const MARKER_BOLD_OPEN = "$$MARKER_BOLD_OPEN$$"
const MARKER_BOLD_CLOSE = "$$MARKER_BOLD_CLOSE$$"
const MARKER_ITALIC_OPEN = "$$MARKER_ITALIC_OPEN$$"
const MARKER_ITALIC_CLOSE = "$$MARKER_ITALIC_CLOSE$$"
const MARKER_NEW_LINE = "$$MARKER_NEW_LINE$$"

var tokenCounter = 0

func getToken(tokenType string) string {
	token := fmt.Sprintf(TOKEN_TEMPLATE, tokenType, tokenCounter)
	tokenCounter++
	return token
}

// https://www.mediawiki.org/wiki/Markup_spec
func tokenize(content string, tokenMap map[string]string) string {
	for {
		content = parseBoldAndItalic(content, tokenMap)
		content = parseInternalLinks(content, tokenMap)
		content = parseExternalLinks(content, tokenMap)
		content = parseTables(content, tokenMap)
		content = parseLists(content, tokenMap)
		break
	}

	return content
}

func parseBoldAndItalic(content string, tokenMap map[string]string) string {
	index := strings.Index(content, "''")
	if index != -1 {
		content, _, _, _ = tokenizeBoldAndItalic(content, index, tokenMap, false, false)
		return content
	}
	return content
}

// tokenizeByRegex applies the regex which must have exactly one group. The tokenized content is returned and a flag
// saying if something changed (i.e. is a tokenization happened).
//func tokenizeByRegex(content string, tokenMap map[string]string, regexString string, tokenType string) (string, bool) {
//	regex := regexp.MustCompile(regexString)
//	matches := regex.FindStringSubmatch(content)
//	if len(matches) != 0 {
//		content = processMatch(content, tokenMap, matches[0], matches[1], tokenType)
//		return content, true
//	}
//	return content, false
//}
//
//func processMatch(content string, tokenMap map[string]string, wholeMatch string, untokenizedMatch string, tokenType string) string {
//	token := getToken(tokenType)
//	tokenizedString := Tokenize(untokenizedMatch, tokenMap)
//	tokenMap[token] = tokenizedString
//	return strings.Replace(content, wholeMatch, token, 1)
//}

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

		if !isBoldOpen && !isItalicOpen {
			break
		}
	}

	return content, index, false, false
}

func parseInternalLinks(content string, tokenMap map[string]string) string {
	regex := regexp.MustCompile(`\[\[([^|^\]]*)\|?(.*?)]]`)
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		tokenArticle := getToken(TOKEN_INTERNAL_LINK_ARTICLE)
		tokenMap[tokenArticle] = submatch[1]

		if submatch[2] == "" {
			// Use article as text
			submatch[2] = submatch[1]
		}

		text := tokenize(submatch[2], tokenMap)
		tokenText := getToken(TOKEN_INTERNAL_LINK_TEXT)
		tokenMap[tokenText] = text

		token := getToken(TOKEN_INTERNAL_LINK)
		tokenMap[token] = tokenArticle + " " + tokenText

		content = strings.Replace(content, submatch[0], token, 1)
	}

	return content
}

func parseExternalLinks(content string, tokenMap map[string]string) string {
	regex := regexp.MustCompile(`([^\[])\[([^ ]*) ?(.*)\]`)
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		tokenUrl := getToken(TOKEN_EXTERNAL_LINK_URL)
		tokenMap[tokenUrl] = submatch[2]

		if submatch[3] == "" {
			// Use URL as text
			submatch[3] = submatch[2]
		}

		text := tokenize(submatch[3], tokenMap)
		tokenText := getToken(TOKEN_EXTERNAL_LINK_TEXT)
		tokenMap[tokenText] = text

		token := getToken(TOKEN_EXTERNAL_LINK)
		tokenMap[token] = tokenUrl + " " + tokenText

		content = strings.Replace(content, submatch[0], submatch[1]+token, 1)
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
			// TODO create token and return
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

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "|-") {
			rowToken := ""
			if strings.HasPrefix(lines[i+1], "!") {
				// this table row is a heading
				rowToken, i = tokenizeTableRow(lines, i+1, "!", tokenMap)
			} else if strings.HasPrefix(lines[i+1], "|") {
				// this table row is a normal row
				rowToken, i = tokenizeTableRow(lines, i+1, "|", tokenMap)
			} else {
				// TODO throw error
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

		splittedLine := strings.SplitN(line, sep, 2)
		if len(splittedLine) == 2 {
			line = splittedLine[1]
		}

		// one column may consist of multiple text rows -> all text lines until the next column or row and tokenize them
		i++
		for ; !strings.HasPrefix(lines[i], sep) && !strings.HasPrefix(lines[i], "|"); i++ {
			line += "\n" + lines[i]
		}
		// now the index is at the start of the next column/row -> reduce by 1 for later parsing
		i -= 1

		line = tokenize(line, tokenMap)
		token := ""
		if sep == "!" {
			token = getToken(TOKEN_TABLE_HEAD)
		} else {
			token = getToken(TOKEN_TABLE_COL)
		}
		tokenMap[token] = line

		rowLines = append(rowLines, token)
	}

	tokenContent := strings.Join(rowLines, " ")

	token := getToken(TOKEN_TABLE_ROW)
	tokenMap[token] = tokenContent

	// return i-1 so that i is on the last line of the row when returning
	return token, i - 1
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
