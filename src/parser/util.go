package parser

import "wiki2book/util"

// FindCorrespondingCloseToken determines the index on which the given openingToken at the startIndex is closed. If the
// closing token has not been found, -1 is returned. This function is case-sensitive.
func FindCorrespondingCloseToken(content string, startIndex int, openingToken string, closingToken string) int {
	return findCorrespondingCloseToken(content, startIndex, openingToken, false, closingToken)
}

// FindCorrespondingCloseTokenIgnoreCase behaves like FindCorrespondingCloseToken but ignores the case of letters. This
// function is case-insensitive.
func FindCorrespondingCloseTokenIgnoreCase(content string, startIndex int, openingToken string, closingTokens ...string) int {
	return findCorrespondingCloseToken(content, startIndex, openingToken, true, closingTokens...)
}

// findCorrespondingCloseToken determines the index on which the given openingToken at the startIndex is closed. If the
// closing token has not been found, -1 is returned.
func findCorrespondingCloseToken(content string, startIndex int, openingToken string, ignoreCase bool, closingTokens ...string) int {
	// Used as a primitive stack to count the degree of nesting the cursor is in. Every opening token increments the
	// counter, every closing token decrements it. If a closing token has been found and the nesting degree is 0, then
	// the correct closing token has been found.
	closeTokenCounter := 0

	// The tokens are considered to be of equal size
	openingTokenSize := len(openingToken)
	contentSize := len(content)

	closingTokenSizes := make([]int, len(closingTokens))
	for i := 0; i < len(closingTokens); i++ {
		closingTokenSizes[i] = len(closingTokens[i])
	}

	for i := startIndex; i < contentSize; i++ {
		cursorOpeningToken := ""

		if i < contentSize-openingTokenSize+1 {
			cursorOpeningToken = content[i : i+openingTokenSize]
		}

		for j := 0; j < len(closingTokens); j++ {
			closingToken := closingTokens[j]
			closingTokenSize := len(closingToken)

			foundNewOpeningToken, cursorIsOnClosingToken := getStatesForCursor(content, i, closingTokenSize, ignoreCase, openingToken, closingToken, cursorOpeningToken)

			if foundNewOpeningToken {
				closeTokenCounter++

				// Skip the found opening token. Use the "-1" to compensate the "+1" by the loop
				i += openingTokenSize - 1

				// Do not check any other closing tokens since the cursor is obviously not sitting on a closing token
				// but on an opening token. There we can assume that all other closing-token-checks will fail.
				break
			} else if cursorIsOnClosingToken {
				if closeTokenCounter == 0 {
					// Found last closing token and therefore the search for exactly this token (specifically its index) ends.
					return i
				} else {
					closeTokenCounter--

					// Skip the found closing token. Use the "-1" to compensate the "+1" by the loop
					i += closingTokenSize - 1

					// Do not search for any more potential close token, since we already found a matching one
					break
				}
			}
		}
	}

	return -1
}

// getStatesForCursor determines the states for the given cursor position i in the content. The first boolean is true,
// when the cursor sits on a new opening token. The second boolean is true, when the cursor is sitting on an closing
// token.
func getStatesForCursor(content string, i int, closingTokenSize int, ignoreCase bool, openingToken string, closingToken string, cursorOpeningToken string) (bool, bool) {
	cursorClosingToken := ""

	if i < len(content)-closingTokenSize+1 {
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

	return foundNewOpeningToken, cursorIsOnClosingToken
}
