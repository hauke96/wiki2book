package parser

// findCorrespondingCloseToken determines the index on which the given openingToken at the startIndex is closed.
func findCorrespondingCloseToken(content string, startIndex int, openingToken string, closingToken string) int {
	// Used as a primitive stack to count the degree of nesting the cursor is in. If a closing token has been found
	// and the nesting degree is 0, then the correct closing token has been found.
	closeTokenCounter := 0

	// The tokens are considered to be of equal size
	openingTokenSize := len(openingToken)
	closingTokenSize := len(closingToken)
	contentSize := len(content)

	for i := startIndex; i < contentSize; i++ {
		cursorOpeningToken := ""
		cursorClosingToken := ""

		if i < contentSize-openingTokenSize+1 {
			cursorOpeningToken = content[i : i+openingTokenSize]
		}

		if i < contentSize-closingTokenSize+1 {
			cursorClosingToken = content[i : i+closingTokenSize]
		}

		if cursorOpeningToken == openingToken {
			closeTokenCounter++
		} else if cursorClosingToken == closingToken {
			if closeTokenCounter == 0 {
				return i
			} else {
				closeTokenCounter--
			}
		}
	}

	return -1
}
