package parser

import (
	"github.com/hauke96/sigolo"
	"strings"
)

func (t *Tokenizer) parseLists(content string) string {
	lines := strings.Split(content, "\n")

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		lineStartCharacter := listPrefixRegex.FindStringSubmatch(line)

		if len(lineStartCharacter) > 0 && lineStartCharacter[1] != "" {
			listTokenString := t.getListTokenString(lineStartCharacter[1])

			// a new list starts here
			token, newIndex := t.tokenizeList(lines, i, lineStartCharacter[1], listTokenString)

			length := newIndex - i

			var newLines []string
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

// tokenizeList takes several lines of text and searches for lists to tokenize. It returns the token to a list and the
// new index. The new index is the index in the input lines and points to the first line that was *not* processed by
// this function.
func (t *Tokenizer) tokenizeList(lines []string, i int, itemPrefix string, tokenString string) (string, int) {
	/*
		Implementation idea:

		We know:
			1. We're at the beginning of a list
			2. We know the list type is "itemPrefix" (so e.g. "*" for unordered lists)

		Strategy:
			1. Find all lines that start with the prefix, i.e. that are relevant for parsing
			2. Remove that prefix from all lines (only the first character to not destroy sub-lists)
			3. Parse each line 'l' and check if it's a normal test line (and therefore a normal list item) or an item of
			   a sub list. In the latter case, recursively tokenize that sub-list and store its token in the token of 'l'.
	*/

	// Step 1: Find all relevant lines for the given list type.
	var listLines []string
	for ; i < len(lines); i++ {
		line := lines[i]

		if !belongsToListPrefix(line, itemPrefix) {
			break
		}

		// Store relevant line without the prefix (step 2 of strategy described above). We know the prefix here (due to
		// the "HasPrefix" call) and don't want the prefix for further line processing.
		listLines = append(listLines, strings.TrimPrefix(line, itemPrefix))
	}

	lineIndex := 0
	var allListItemTokens []string

	// Step 3: Parse each line.
	for ; lineIndex < len(listLines); lineIndex++ {
		line := listLines[lineIndex]
		if line == "" {
			continue
		}

		linePrefix := string(line[0])

		token := ""

		// Does this line has a line prefix? Or in other words: Does this line start a new sub-list? This is the case
		// when the regex matches. When we're parsing a description list, then the current line must be a non
		// description list line in order to start a sub-list.
		if listPrefixRegex.MatchString(line) &&
			!(linePrefix == ":" && itemPrefix == ";") &&
			!(linePrefix == ";" && itemPrefix == ":") {
			// Yes, line is a sub-list beginning: Parse this sub-list recursively
			token, lineIndex = t.tokenizeList(listLines, lineIndex, linePrefix, t.getListTokenString(linePrefix))
		} else {
			// No sub-list starts, line is just text:
			// First check if the next line starts a new sub-list. If so, we have to parse that first, because the
			// token of that sub-list must be within the token content of the current line token. Otherwise, the sub-
			// list will be part of a new and empty item of the current list we're in. That's not what we want.

			subListToken := ""
			// Is there a next line and does it start a sub-list?
			if lineIndex < len(listLines)-1 && listPrefixRegex.MatchString(listLines[lineIndex+1]) {
				// Yes -> Get the prefix of the next line (to see what kind of list it starts)
				nextLinePrefix := string(listLines[lineIndex+1][0])

				// Is it a list of description list items? If so, we don't consider it a new list start, because a
				// description list starts with a ";" (semicolon).
				if nextLinePrefix != ":" {
					// When there's a sub-list, then this sub list token will be part of the current list item token. Therefore,
					// we have to start parsing the sub-list before creating the token for the current list item.
					subListToken, lineIndex = t.tokenizeList(listLines, lineIndex+1, nextLinePrefix, t.getListTokenString(nextLinePrefix))
					lineIndex-- // compensates the  lineIndex++  from the for-loop
				}
			}

			tokenItemPrefix := itemPrefix
			if linePrefix == ":" {
				// If this line is a description list item, we remove its prefix as it wasn't removed at beginning (s.
				// above: Description lists usually start with ";" (semicolon) but items are marked with ":" (colon)
				// which hasn't been removed to distinguish description list headings and items.
				tokenItemPrefix = linePrefix
				line = line[1:]
			}

			tokenContent := t.tokenizeContent(t, line)

			if subListToken != "" {
				// Store the token of the sub list, that was found above, in the token of the current list item. This
				// ensures that no new dummy-list item is needed to store the token of the sub-list.
				tokenContent += " " + subListToken
			}

			tokenString := t.getListItemTokenString(tokenItemPrefix)
			if tokenString == TOKEN_DESCRIPTION_LIST_HEAD {
				// A description list may contain content within the line of a heading. Something like this: "; foo: bar"

				// Split the heading-part from potential content
				lineParts := strings.Split(tokenContent, ":")

				headPart := lineParts[0]
				token = t.getToken(tokenString)
				t.setRawToken(token, headPart)

				if len(lineParts) > 1 {
					// Add heading token to that the variable can be overridden without losing the heading token
					allListItemTokens = append(allListItemTokens, token)

					// There's content after the heading -> Create separate token for that
					contentPart := strings.Join(lineParts[1:], ":")
					token = t.getToken(TOKEN_DESCRIPTION_LIST_ITEM)
					t.setRawToken(token, contentPart)
				}
			} else {
				token = t.getToken(tokenString)
				t.setRawToken(token, tokenContent)
			}
		}

		allListItemTokens = append(allListItemTokens, token)
	}

	tokenContent := strings.Join(allListItemTokens, " ")
	token := t.getToken(tokenString)
	t.setRawToken(token, tokenContent)

	return token, i
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

func (t *Tokenizer) getListTokenString(listItemPrefix string) string {
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
	sigolo.Error("Unable to get list token string: Unknown list item prefix %s", listItemPrefix)
	return "UNKNOWN_LIST_TYPE_" + listItemPrefix
}

func (t *Tokenizer) getListItemTokenString(listItemPrefix string) string {
	switch listItemPrefix {
	case "*":
		return TOKEN_LIST_ITEM
	case "#":
		return TOKEN_LIST_ITEM
	case ";":
		return TOKEN_DESCRIPTION_LIST_HEAD
	case ":":
		return TOKEN_DESCRIPTION_LIST_ITEM
	}
	sigolo.Error("Unable to get list item token string: Unknown list item prefix %s", listItemPrefix)
	return "UNKNOWN_LIST_ITEM_TYPE_" + listItemPrefix
}
