package parser

import (
	"regexp"
	"strings"
)

// FindCorrespondingCloseToken determines the index on which the given openingToken at the startIndex is closed. If the
// closing token has not been found, -1 is returned. This function is case-sensitive.
func FindCorrespondingCloseToken(content string, startIndex int, openingToken string, closingToken string) int {
	return findCorrespondingCloseToken(content, startIndex, openingToken, closingToken)
}

// findCorrespondingCloseToken determines the index on which the given openingToken at the startIndex is closed. If the
// closing token has not been found, -1 is returned.
func findCorrespondingCloseToken(content string, startIndex int, openingToken string, closingToken string) int {
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

		openingAndClosingTokenAreDifferent := openingToken != closingToken
		cursorIsOnOpeningToken := cursorOpeningToken == openingToken
		cursorIsOnClosingToken := cursorClosingToken == closingToken

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

// FindXmlCloseToken determines the index of the closing XML element. It handles nested XML structures, i.e. all
// opening and closing XML items. It also assumes that the startIndex is current within an XML element. The returned
// index corresponds to this at startIndex opened XML item. If the closing token has not been found, -1 is returned.
func FindXmlCloseToken(content string, startIndex int) int {
	// Used as a primitive stack to count the degree of nesting the cursor is in. Every opening token increments the
	// counter, every closing token decrements it. If a closing token has been found and the nesting degree is 0, then
	// the correct closing token has been found.
	closeTokenCounter := 0

	// Does not comply with XML standard but is simpler and will probably work all the time. When there are real world
	// edge-cases, then this expression should be adjusted.
	// Matches to i.e. "<ref " and "<xds:some-item>" but not to "</ref>" or "3 < 4"
	isXmlItemStart := regexp.MustCompile("^<[^ />]+")

	for i := startIndex; i < len(content); i++ {
		// There's only one way of XML items opening, e.g. "<ref"
		cursorIsOnOpeningToken := isXmlItemStart.MatchString(content[i:])
		// There are multiple ways for XML items to close, e.g. "... />" or "</ref>
		cursorIsOnClosingToken := strings.HasPrefix(content[i:], "/>") || strings.HasPrefix(content[i:], "</")

		if cursorIsOnOpeningToken {
			closeTokenCounter++

			// Skip the found opening token. Use the "-1" to compensate the "+1" by the loop
			i++
		} else if cursorIsOnClosingToken {
			if closeTokenCounter == 0 {
				return i
			} else {
				closeTokenCounter--

				// Skip the found closing token. Use the "-1" to compensate the "+1" by the loop
				i++
			}
		}
	}

	return -1
}
