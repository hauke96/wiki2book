package parser

import (
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/util"
	"strings"
)

type BoldItalicStackItemType int

const (
	BOLD_OPEN BoldItalicStackItemType = iota
	BOLD_CLOSE
	ITALIC_OPEN
	ITALIC_CLOSE
)

type BoldItalicStackItem struct {
	itemType BoldItalicStackItemType
	index    int
	length   int
}

func (t *Tokenizer) parseBoldAndItalic(content string) string {
	var success bool
	index := 0

	// The idea: Search for the first  '  character in the content, this means a new bold or italic block might start.
	// This block then gets parsed and we create the according markers. Then we search for the next block. This is
	// based on the assumption that only a small percentage of the content is actually part of an italic or bold block.
	// Because this approach is based on recursion, this per-block parsing reduces the recursion depth.
	for {
		// Skip uninteresting characters
		for index < len(content)-1 && content[index:index+2] != "''" {
			index++
		}

		// Reached the end of text? -> break
		if index >= len(content)-1 {
			break
		}

		var stack []BoldItalicStackItem
		// First try everything without repairing crossovers. Crossovers can happen even if a normal solution is possible.
		success, stack = t.tokenizeBoldAndItalic(content, index, stack, false, false, false)
		if !success {
			sigolo.Error("Unable to parse bold and italic tags WITHOUT repairing crossovers in: %s. I'll try it again with repairing crossovers enabled.", util.TruncString(content))
			stack = []BoldItalicStackItem{}
			// Okay, not try to repair crossovers.
			success, stack = t.tokenizeBoldAndItalic(content, index, stack, true, false, false)

			if !success {
				sigolo.Error("Unable to parse bold and italic tags EVEN WITH repairing crossovers in: %s", util.TruncString(content))
				return content
			}
		}

		// the index increase when inserting the markers as the content gets longer
		offset := 0
		chars := []byte(content)
		for _, item := range stack {
			itemIndex := item.index + offset
			lenBefore := len(chars)
			charsBefore := chars[:itemIndex]
			var charsAfter []byte

			switch item.itemType {
			case BOLD_OPEN:
				charsAfter = append([]byte(nil), chars[itemIndex+item.length:]...)
				chars = append(charsBefore, []byte(MARKER_BOLD_OPEN)...)
			case BOLD_CLOSE:
				charsAfter = append([]byte(nil), chars[itemIndex+item.length:]...)
				chars = append(charsBefore, []byte(MARKER_BOLD_CLOSE)...)
			case ITALIC_OPEN:
				charsAfter = append([]byte(nil), chars[itemIndex+item.length:]...)
				chars = append(charsBefore, []byte(MARKER_ITALIC_OPEN)...)
			case ITALIC_CLOSE:
				charsAfter = append([]byte(nil), chars[itemIndex+item.length:]...)
				chars = append(charsBefore, []byte(MARKER_ITALIC_CLOSE)...)
			}

			index = len(chars)
			chars = append(chars, charsAfter...)
			offset += len(chars) - lenBefore
		}

		content = string(chars)
	}

	return content
}

// tokenizeBoldAndItalic takes the content, an index in that content, if this index is currently in a bold or italic
// block and a stack to parse the content for bold and italic blocks. It returns a success-bool and the resulting stack.
func (t *Tokenizer) tokenizeBoldAndItalic(content string, index int, stack []BoldItalicStackItem, repairCrossovers, isBoldOpen, isItalicOpen bool) (bool, []BoldItalicStackItem) {
	// The idea of this approach:
	// Find the next possible match for a start/stop of an italic/bold block. Put that item on the stack (e.g.
	// BOLD_START), increase the index and recursively move on to further parse the content. Whenever something's
	// "flaky", stop using the new stack and proceed with the old one which has one item less.
	// Example of such an exit-condition: Say  ''  have been interpreted as ITALIC_OPEN but one  '  is left -> this
	// should probably a BOLD_OPEN or BOLD_CLOSE as there are three  '''. Move one step back, interpret the  '''  as a
	// BOLD-block and move on from there.

	for index < len(content) {
		nextTokenMightBeItalic := index+2 <= len(content) && content[index:index+2] == "''"
		nextTokenMightBeBold := index+3 <= len(content) && content[index:index+3] == "'''"
		mightBelongToPrevToken := index > 0 && content[index-1:index+1] == "''"

		if nextTokenMightBeItalic {
			continueParsingAsItalicToken := true
			itemType := ITALIC_OPEN
			if isItalicOpen {
				itemType = ITALIC_CLOSE
			}

			newStack := append(stack, BoldItalicStackItem{itemType: itemType, index: index, length: 2})

			// Crossover = The combination of opening bold followed by closing italic is invalid in most markup languages
			// -> insert dummy-items to resolve this issue
			hasCrossover := len(stack) > 0 && stack[len(stack)-1].itemType == BOLD_OPEN && itemType == ITALIC_CLOSE
			if hasCrossover {
				if repairCrossovers {
					newStack = append(stack, BoldItalicStackItem{itemType: BOLD_CLOSE, index: index, length: 0})
					newStack = append(newStack, BoldItalicStackItem{itemType: itemType, index: index, length: 2})
					newStack = append(newStack, BoldItalicStackItem{itemType: BOLD_OPEN, index: index + 2, length: 0})
				} else {
					if !nextTokenMightBeBold {
						// Return if parsing failed and continuing with parsing the current token als bold-token is
						// useless as this token is definitely not a bold one.
						return false, stack
					}
					// Parsing the current token as italic didn't work, but maybe parsing it as a bold token does work.
					continueParsingAsItalicToken = false
				}
			}

			if continueParsingAsItalicToken {
				success, newStack := t.tokenizeBoldAndItalic(content, index+2, newStack, repairCrossovers, isBoldOpen, !isItalicOpen)

				if success {
					// path went well -> end recursion and return
					return true, newStack
				}
			}

			// path went not well -> use old stack and try to match it with a bold item
		}
		if nextTokenMightBeBold {
			itemType := BOLD_OPEN
			if isBoldOpen {
				itemType = BOLD_CLOSE
			}

			newStack := append(stack, BoldItalicStackItem{itemType: itemType, index: index, length: 3})

			// Crossover = The combination of opening italic followed by closing bold is invalid in most markup languages
			// -> insert dummy-items to resolve this issue
			hasCrossover := len(stack) > 0 && stack[len(stack)-1].itemType == ITALIC_OPEN && itemType == BOLD_CLOSE
			if hasCrossover {
				if repairCrossovers {
					newStack = append(stack, BoldItalicStackItem{itemType: ITALIC_CLOSE, index: index, length: 0})
					newStack = append(newStack, BoldItalicStackItem{itemType: itemType, index: index, length: 3})
					newStack = append(newStack, BoldItalicStackItem{itemType: ITALIC_OPEN, index: index + 3, length: 0})
				} else {
					// Characters are  '''  and parsing them as italic (happened before this part of the code) didn't
					// work. Therefor we must abort here as this crossover cannot be resolved without dummy-items.
					return false, stack
				}
			}

			success, newStack := t.tokenizeBoldAndItalic(content, index+3, newStack, repairCrossovers, !isBoldOpen, isItalicOpen)

			if success {
				// path went well -> end recursion and return
				return true, newStack
			}

			// path went also not well -> tried italic *and* bold = all possibilities failed -> abort
			return false, stack
		}

		if mightBelongToPrevToken {
			// The current character belongs to an italic/bold token but did not match the above -> wrong token chosen
			// in previous call (e.g. an italic token was chosen when there were ''' but instead a bold token makes more
			// sense here).
			return false, stack
		}

		if nextTokenMightBeItalic {
			// When this line is reached it means parsing wasn't successful. Parsing italic and bold tokens (including
			// return statements) happens above, so something went wrong.
			return false, stack
		}

		// This is true when for all opening tags a closing tag was found and therefore no opening tags are left. This
		// means all following characters (maybe even whole sentences or paragraphs) are not relevant for us. So we
		// abort here.
		closedBlock := len(stack) > 0 && hasOddNumberOfItems(stack)
		if closedBlock {
			break
		}

		index++
	}

	return hasOddNumberOfItems(stack), stack
}

// parseParagraphs replaces two directly following newlines by a `MARKER_PARAGRAPH` marker. When the line before the two
// newlines is a line containing a token, the two consecutive newlines do NOT count as a paragraph. This is because we
// assume a token to be self-contained without the need ot extra space below it.
func (t *Tokenizer) parseParagraphs(content string) string {
	oldLines := strings.Split(content, "\n")
	var resultLines []string

	ignoreEmptyLinesMode := true

	// Setting this to "true" marks a wish to create a marker. If a marker is created depends on the surrounding
	// lines. No paragraphs are added before and after a token line.
	createParagraphWhenAllowed := false

	for i := 0; i < len(oldLines); i++ {
		line := oldLines[i]
		lineBefore := ""
		if i-1 >= 0 {
			lineBefore = oldLines[i-1]
		}

		if tokenLineRegex.MatchString(line) {
			ignoreEmptyLinesMode = true
			createParagraphWhenAllowed = false
			resultLines = append(resultLines, line)
			continue
		}

		if ignoreEmptyLinesMode {
			if line != "" {
				// End mode of ignoring blank lines, because we reached a non-blank and non-token line
				ignoreEmptyLinesMode = false

				if createParagraphWhenAllowed {
					// If there was a normal line before (and not a token-line), then create a paragraph marker
					resultLines = append(resultLines, MARKER_PARAGRAPH)
				}

				createParagraphWhenAllowed = false
				resultLines = append(resultLines, line)
			}
		} else {
			if line == "" && lineBefore != "" {
				// Previous line was not blank & line before was not a token line & we're not in blank-ignore-mode
				// -> start blank-ignore-mode and make a wish to add a marker
				ignoreEmptyLinesMode = true
				createParagraphWhenAllowed = true
			} else {
				// Handling of normal lines
				resultLines = append(resultLines, line)
			}
		}
	}

	return strings.Join(resultLines, "\n")
}

// hasOddNumberOfItems determines if e.g. there are more opening bold tags than closing ones (-> number of bold tags is odd).
func hasOddNumberOfItems(stack []BoldItalicStackItem) bool {
	numberOfItalics := 0
	numberOfBolds := 0

	for _, item := range stack {
		switch item.itemType {
		case BOLD_OPEN:
			numberOfBolds++
		case BOLD_CLOSE:
			numberOfBolds--
		case ITALIC_OPEN:
			numberOfItalics++
		case ITALIC_CLOSE:
			numberOfItalics--
		}
	}

	return numberOfItalics == 0 && numberOfBolds == 0
}
