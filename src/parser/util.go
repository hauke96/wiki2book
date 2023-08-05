package parser

// findCorrespondingCloseToken determines the index on which the given openingToken at the startIndex is closed.
func findCorrespondingCloseToken(content string, startIndex int, openingToken string, closingToken string) int {
	// Used as a primitive stack to count the degree of nesting the cursor is in. If a closing token has been found
	// and the nesting degree is 0, then the correct closing token has been found.
	closeTokenCounter := 0

	// The tokens are considered to be of equal size
	tokenSize := len(openingToken)

	for i := startIndex; i < len(content)-1; i++ {
		cursor := content[i : i+tokenSize]

		if cursor == openingToken {
			closeTokenCounter++
		} else if cursor == closingToken {
			if closeTokenCounter == 0 {
				return i
			} else {
				closeTokenCounter--
			}
		}
	}

	return -1
}
