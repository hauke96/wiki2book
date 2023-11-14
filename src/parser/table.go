package parser

import "strings"

type TableToken struct {
	Token
	Rows []Token
}

type TableHeadToken struct {
	Token
	Attributes TableColAttributeToken
	Content    string
}

type TableRowToken struct {
	Token
	Columns []Token
}

type TableColToken struct {
	Token
	Attributes TableColAttributeToken
	Content    string
}

type TableCaptionToken struct {
	Token
	Attributes TableColAttributeToken
	Content    string
}

type TableColAttributeToken struct {
	Token
	Attributes []string
}

func (t *Tokenizer) parseTables(content string) string {
	lines := strings.Split(content, "\n")

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		submatch := tableStartRegex.FindStringSubmatch(line)
		if submatch != nil {
			listPrefix := submatch[1]

			// table starts in this line.
			token, newIndex := t.tokenizeTables(lines, i)

			length := newIndex - i

			var newLines []string
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
	var tableLines []string
	tableLines = append(tableLines, lines[i])
	i++

	// collect all lines from this table
	for ; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

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
			tableLines = append(tableLines, lines[i])
		}
	}

	tableContent := strings.Join(tableLines, "\n")
	token := t.tokenizeTable(tableContent)

	tokenKey := t.getToken(TOKEN_TABLE)
	t.setRawToken(tokenKey, token)

	return tokenKey, i
}

// tokenizeTable expects content to be all lines of a table.
func (t *Tokenizer) tokenizeTable(content string) TableToken {
	// ensure that each entry starts in a new row
	content = strings.ReplaceAll(content, "||", "\n|")
	content = strings.ReplaceAll(content, "!!", "\n!")
	lines := strings.Split(content, "\n")

	var rowTokens []Token

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		isCaptionStart := strings.HasPrefix(line, "|+")
		isRowStart := strings.HasPrefix(line, "|-")
		isDataCell := strings.HasPrefix(line, "|")
		isHeadingStart := strings.HasPrefix(line, "!")
		isEndOfTable := strings.HasPrefix(line, "|}")

		var rowToken Token
		if isCaptionStart {
			rowToken, i = t.tokenizeTableCaption(lines, i)
		} else if isHeadingStart {
			rowToken, i = t.tokenizeTableRow(lines, i)
		} else if isRowStart {
			if strings.HasPrefix(lines[i+1], "!") {
				// this table row is a heading
				rowToken, i = t.tokenizeTableRow(lines, i+1)
			} else if strings.HasPrefix(lines[i+1], "|") {
				// this table row is a normal row
				rowToken, i = t.tokenizeTableRow(lines, i+1)
			}
		} else if isEndOfTable {
			// table ends with this line, so we can end this loop
			break
		} else if isDataCell {
			// A data cell/row is usually part of a normal row or heading. Here we found a data row but without an
			// explicit start of a new table row. So we just assume that here one or more normal table rows start.
			rowToken, i = t.tokenizeTableRow(lines, i)
		}

		if rowToken != nil {
			rowTokens = append(rowTokens, rowToken)
		}
	}

	token := TableToken{
		Rows: rowTokens,
	}

	return token
}

// tokenizeTableRow expects i to be the line with the first text item (i.e. the line after |- ). Furthermore, this
// function expects that each column of this row starts in a new line starting with  |  or  !  . The returned string is
// never nil and an empty string represents an empty row, that can be ignored. The index points to the last text line of
// this table row.
func (t *Tokenizer) tokenizeTableRow(lines []string, i int) (TableRowToken, int) {
	var rowToken TableRowToken
	var columnTokens []Token

	// collect all lines from this row
	for ; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if strings.HasPrefix(line, "|-") || strings.HasPrefix(line, "|+") || strings.HasPrefix(line, "|}") {
			// Row or whole table ended.
			break
		}

		line = strings.TrimPrefix(line, "|")
		line = strings.TrimPrefix(line, "!")
		line = strings.TrimSpace(line)

		// Collect all normal text rows until the next row or column starts.
		i++
		for l := strings.TrimSpace(lines[i]); !strings.HasPrefix(l, "|") && !strings.HasPrefix(l, "!"); {
			line += "\n" + l

			i++
			l = lines[i]
		}
		// Now the index is at the start of the next column/row -> reduce by 1 for later parsing.
		i -= 1

		tokenizedLine, attributeToken := t.tokenizeTableEntry(line)
		tokenizedLine = strings.TrimSpace(tokenizedLine)

		var token Token
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "!") {
			token = TableHeadToken{
				Attributes: attributeToken,
				Content:    tokenizedLine,
			}
		} else {
			token = TableColToken{
				Attributes: attributeToken,
				Content:    tokenizedLine,
			}
		}

		columnTokens = append(columnTokens, token)
	}

	// TODO Can this be removed? Wouldn't it remove empty rows because "nil" is later interpreted as "nothing to see here"?
	if len(columnTokens) == 0 {
		return rowToken, i - 1
	}

	rowToken = TableRowToken{
		Columns: columnTokens,
	}

	// return i-1 so that i is on the last line of the row when returning
	return rowToken, i - 1
}

// tokenizeTableCaption expects i to be the line in which the caption starts (i.e. the line after |+ ). The return
// values are the tokenized caption and the index pointing to the last text line of the caption.
func (t *Tokenizer) tokenizeTableCaption(lines []string, i int) (TableCaptionToken, int) {
	captionLines := strings.TrimSpace(lines[i])

	// collect all lines from this caption
	i++
	for l := strings.TrimSpace(lines[i]); !strings.HasPrefix(l, "|") && !strings.HasPrefix(l, "!"); {
		captionLines += "\n" + l

		i++
		l = strings.TrimSpace(lines[i])
	}

	captionLines = strings.TrimPrefix(captionLines, "|+")

	tokenizedCaption, styleAttributeToken := t.tokenizeTableEntry(captionLines)
	tokenizedCaption = strings.TrimSpace(tokenizedCaption)
	captionToken := TableCaptionToken{
		Attributes: styleAttributeToken,
		Content:    tokenizedCaption,
	}

	// return i-1 so that i is on the last line of the caption when returning
	return captionToken, i - 1
}

// tokenizeTableEntry returns the tokenized text of the entry (for example a column or caption) and a token containing
// the style attributes (might be empty when no style was found).
func (t *Tokenizer) tokenizeTableEntry(content string) (string, TableColAttributeToken) {
	var attributeToken TableColAttributeToken

	splittedContent := strings.Split(content, "|")
	if len(splittedContent) < 2 {
		return t.tokenizeContent(t, content), attributeToken
	}

	attributeString := strings.TrimSpace(splittedContent[0])
	entryText := t.tokenizeContent(t, splittedContent[1])

	var relevantTags []string

	rowAndColspanMatch := tableRowAndColspanRegex.FindStringSubmatch(attributeString)
	if len(rowAndColspanMatch) > 1 {
		relevantTags = append(relevantTags, rowAndColspanMatch[0])
	}

	alignmentMatch := tableTextAlignRegex.FindStringSubmatch(attributeString)
	if len(alignmentMatch) > 0 {
		relevantTags = append(relevantTags, `style="`+alignmentMatch[0]+`"`)
	}

	if len(relevantTags) > 0 {
		attributeToken = TableColAttributeToken{
			Attributes: relevantTags,
		}
	}

	return entryText, attributeToken
}
