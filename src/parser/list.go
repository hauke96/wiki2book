package parser

import (
	"fmt"
	"github.com/hauke96/sigolo/v2"
	"strings"
)

type ListToken interface {
	Token
}

type OrderedListToken struct {
	ListToken
	Items []ListItemToken
}

type UnorderedListToken struct {
	ListToken
	Items []ListItemToken
}

type ListItemTokenType int

const (
	NORMAL_ITEM ListItemTokenType = iota
	DESCRIPTION_ITEM
	DESCRIPTION_HEAD
)

type ListItemToken struct {
	ListToken
	Type     ListItemTokenType
	Content  string
	SubLists []ListToken
}

type DescriptionListToken struct {
	ListToken
	Items []ListItemToken
}

func (t *Tokenizer) parseLists(content string) string {
	lines := strings.Split(content, "\n")

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if line == "" || !isListStartingCharacter(line[0]) {
			continue
		}

		lineStartCharacter := string(line[0])
		_, tokenKey, newIndex := t.tokenizeList(lines, i, lineStartCharacter)

		length := newIndex - i

		var newLines []string
		newLines = append(newLines, lines[:i]...)
		newLines = append(newLines, tokenKey)
		if i+length < len(lines) {
			newLines = append(newLines, lines[i+length:]...)
		}
		lines = newLines
	}

	content = strings.Join(lines, "\n")
	return content
}

// tokenizeList takes several lines of text and searches for lists to tokenize. It returns the token to a list and the
// new index. The new index is the index in the input lines and points to the first line that was *not* processed by
// this function.
func (t *Tokenizer) tokenizeList(lines []string, startLineIndex int, listPrefix string) (ListToken, string, int) {
	/*
		Implementation idea:

		We know:
			1. We're at the beginning of a list
			2. We know the list type is "listPrefix" (so e.g. "*" for unordered lists)

		Strategy:
			1. Find all lines that start with the prefix, i.e. that are relevant for parsing
			2. Remove that prefix from all lines (only the first character to not destroy sub-lists)
			3. Parse each line 'l' and check if it's a normal text line (and therefore a normal list item) or an item of
			   a sub list. In the latter case, recursively tokenize that sub-list and store its token in the token of 'l'.
	*/

	// Step 1: Find all relevant lines for the given list type.
	var listLines []string
	for i := startLineIndex; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if !belongsToListPrefix(line, listPrefix) {
			break
		}

		// Store relevant line without the prefix (step 2 of strategy described above). We know the prefix here (due to
		// the "belongsToListPrefix" call) and don't want the prefix for further line processing.
		listLines = append(listLines, line)
	}

	var allListItemTokens []ListItemToken

	// Step 3: Parse each line.
	for lineIndex := 0; lineIndex < len(listLines); lineIndex++ {
		line := listLines[lineIndex]
		if line == "" {
			continue
		}

		// line[0] is the prefix of the current list nesting level. The "linePrefix" is - except for description list
		// items - equal to the list prefix.
		linePrefix := string(line[0])

		// Does this line has a line prefix? Or in other words: Does this line start a new sub-list? This is the case
		// when the regex matches. When we're parsing a description list, then the current line must be a non
		// description list line in order to start a sub-list.
		if len(line) > 1 && isListStartingCharacter(line[1]) {
			// Yes, line is a sub-list beginning: Parse this sub-list recursively

			subListPrefix := string(line[1])

			// Remove the line prefix to ensure that we're not stuck in an endless loop of list parsing.
			var trimmedListLines []string
			for j := lineIndex; j < len(listLines); j++ {
				if !belongsToListPrefix(listLines[j][1:], subListPrefix) {
					break
				}
				trimmedListLines = append(trimmedListLines, listLines[j][1:])
			}

			subListToken, _, _ := t.tokenizeList(trimmedListLines, 0, subListPrefix)

			// Add the sub-list token to the previous list item. Of no previous list item exists, create one, which will
			// only serve as a hull (or dummy) for the actual sub list.
			hasPreviousListItem := len(allListItemTokens) > 0
			if !hasPreviousListItem {
				listItemType := t.getListItemTypeForList(listPrefix)
				subListDummyToken := ListItemToken{
					Type:     listItemType,
					Content:  "",
					SubLists: []ListToken{subListToken},
				}
				allListItemTokens = append(allListItemTokens, subListDummyToken)
			} else {
				previousListItem := &allListItemTokens[len(allListItemTokens)-1]
				previousListItem.SubLists = appendOrCreate(previousListItem.SubLists, subListToken)
			}

			lineIndex += len(trimmedListLines)
			// Compensate "lineIndex++" from the for-loop
			lineIndex--
		} else {
			// No sub-list starts, line is just text:
			// First check if the next line starts a new sub-list. If so, we have to parse that first, because the
			// token of that sub-list must be within the token content of the current line token. Otherwise, the sub-
			// list will be part of a new and empty item of the current list we're in. That's not what we want.

			if linePrefix == ";" {
				// Description list head -> also check for description list item in the same line
				lineParts := strings.SplitN(line[1:], ":", 2)

				headPart := t.tokenizeContent(t, lineParts[0])
				headToken := ListItemToken{
					Type:    DESCRIPTION_HEAD,
					Content: headPart,
				}
				allListItemTokens = append(allListItemTokens, headToken)

				if len(lineParts) > 1 {
					// There's a description list item after the heading -> Create separate token for that
					itemPart := t.tokenizeContent(t, lineParts[1])
					itemToken := ListItemToken{
						Type:    DESCRIPTION_ITEM,
						Content: itemPart,
					}
					allListItemTokens = append(allListItemTokens, itemToken)
				}
			} else if linePrefix == ":" {
				// Description list item
				itemToken := ListItemToken{
					Type:    DESCRIPTION_ITEM,
					Content: t.tokenizeContent(t, line[1:]),
				}
				allListItemTokens = append(allListItemTokens, itemToken)
			} else {
				// Normal list item
				tokenizedLine := t.tokenizeContent(t, line[1:])
				lineToken := ListItemToken{
					Type:    NORMAL_ITEM,
					Content: tokenizedLine,
				}
				allListItemTokens = append(allListItemTokens, lineToken)
			}
		}
	}

	listTokenKind := t.getListTokenKey(listPrefix)
	listTokenKey := t.getToken(listTokenKind)
	var listToken ListToken
	switch listPrefix {
	case "*":
		listToken = UnorderedListToken{
			Items: allListItemTokens,
		}
	case "#":
		listToken = OrderedListToken{
			Items: allListItemTokens,
		}
	case ":":
		fallthrough
	case ";":
		listToken = DescriptionListToken{
			Items: allListItemTokens,
		}
	}
	t.setRawToken(listTokenKey, listToken)
	return listToken, listTokenKey, startLineIndex + len(listLines)
}

func belongsToListPrefix(line string, itemPrefix string) bool {
	if len(line) == 0 {
		return false
	}

	linePrefix := string(line[0])

	// Special case for description lists: They have two separate prefixes: ";" and ":" for the heading and items themselves
	if itemPrefix == ";" {
		return linePrefix == ":" || linePrefix == ";"
	}

	return linePrefix == itemPrefix
}

func isListStartingCharacter(c uint8) bool {
	return c == '*' || c == '#' || c == ';' || c == ':'
}

func removeListPrefix(line string, itemPrefix string) string {
	if len(line) == 0 {
		return line
	}

	linePrefix := string(line[0])

	// Special case for description lists: They have two separate prefixes: ";" and ":" for the heading and items themselves
	if itemPrefix == ";" {
		if linePrefix == ";" {
			return strings.TrimPrefix(line, linePrefix)
		} else if linePrefix == ":" {
			return strings.TrimPrefix(line, linePrefix)
		}
		return line
	}

	return strings.TrimPrefix(line, itemPrefix)
}

func (t *Tokenizer) getListTokenKey(listItemPrefix string) string {
	switch listItemPrefix {
	case "*":
		return TOKEN_UNORDERED_LIST
	case "#":
		return TOKEN_ORDERED_LIST
	case ";":
		return TOKEN_DESCRIPTION_LIST
	case ":":
		return TOKEN_DESCRIPTION_LIST
	}
	sigolo.Errorf("Unable to get list token key: Unknown list item prefix %s", listItemPrefix)
	return fmt.Sprintf(TOKEN_UNKNOWN_LIST_ITEM, listItemPrefix)
}

func (t *Tokenizer) getListItemTypeForList(listItemPrefix string) ListItemTokenType {
	switch listItemPrefix {
	case "*":
		return NORMAL_ITEM
	case "#":
		return NORMAL_ITEM
	case ";":
		return DESCRIPTION_HEAD
	case ":":
		return DESCRIPTION_ITEM
	}
	sigolo.Errorf("Unable to get list item type from list prefix '%s'", listItemPrefix)
	return -1
}

func appendOrCreate[T interface{}](list []T, item T) []T {
	if list == nil {
		return []T{item}
	}
	return append(list, item)
}
