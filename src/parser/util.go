package parser

import "wiki2book/util"

// FindCorrespondingCloseToken determines the index on which the given openingToken at the startIndex is closed. If the
// closing token has not been found, -1 is returned. This function is case-sensitive.
func FindCorrespondingCloseToken(content string, startIndex int, openingToken string, closingToken string) int {
	return findCorrespondingCloseToken(content, startIndex, openingToken, closingToken, false)
}

// FindCorrespondingCloseTokenIgnoreCase behaves like FindCorrespondingCloseToken but ignores the case of letters. This
// function is case-insensitive.
func FindCorrespondingCloseTokenIgnoreCase(content string, startIndex int, openingToken string, closingToken string) int {
	return findCorrespondingCloseToken(content, startIndex, openingToken, closingToken, true)
}

// findCorrespondingCloseToken determines the index on which the given openingToken at the startIndex is closed. If the
// closing token has not been found, -1 is returned.
func findCorrespondingCloseToken(content string, startIndex int, openingToken string, closingToken string, ignoreCase bool) int {
	// Used as a primitive stack to count the degree of nesting the cursor is in. Every opening token increments the
	// counter, every closing token decrements it. If a closing token has been found and the nesting degree is 0, then
	// the correct closing token has been found.
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

		openingAndClosingTokenAreDifferent := false
		cursorIsOnOpeningToken := false
		if ignoreCase {
			openingAndClosingTokenAreDifferent = !util.EqualsIgnoreCase(openingToken, closingToken)
			cursorIsOnOpeningToken = util.EqualsIgnoreCase(cursorOpeningToken, openingToken)
		} else {
			openingAndClosingTokenAreDifferent = openingToken != closingToken
			cursorIsOnOpeningToken = cursorOpeningToken == openingToken
		}

		cursorIsOnClosingToken := false
		if ignoreCase {
			cursorIsOnClosingToken = util.EqualsIgnoreCase(cursorClosingToken, closingToken)
		} else {
			cursorIsOnClosingToken = cursorClosingToken == closingToken
		}

		foundNewOpeningToken := openingAndClosingTokenAreDifferent && cursorIsOnOpeningToken
		if foundNewOpeningToken {
			closeTokenCounter++

			// Skip the found opening token. Use the "-1" to compensate the "+1" by the loop
			i += openingTokenSize - 1
		} else if cursorIsOnClosingToken {
			if closeTokenCounter == 0 {
				return i
			} else {
				closeTokenCounter--

				// Skip the found closing token. Use the "-1" to compensate the "+1" by the loop
				i += closingTokenSize - 1
			}
		}
	}

	return -1
}
