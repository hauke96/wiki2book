package parser

import (
	"github.com/hauke96/sigolo"
)

/*
Sorry for the name but that's what it is:
Some Wikpedia-specific stuff is just to weird or is language specific and has to be removed by these hack-functions.
*/

// hackGermanRailwayTemplates takes the content and removed the combination "{{BS-table}} ... |}", because apparently
// the "{{BS-table}}" template generates the head of a table, which simply is closed by a "|}". This is very specific to
// this template, requires knowledge about its evaluation/use and is therefore considered a hack.
func hackGermanRailwayTemplates(content string, startIndex int) string {

	// TODO It can happen that a template looks like this: "{{template|args|}}" and the "|}" part confuses this hack function.
	// This is because it thinks that "|}" is the end of a table, which it isn't in this case.

	startToken := "{{BS-table}}"
	endToken := "|}"
	slidingWindowSize := len(startToken)
	endTokenSize := len(endToken)

	for i := startIndex; i < len(content)-slidingWindowSize; i++ {
		cursor := content[i : i+slidingWindowSize]

		if cursor == startToken {
			// Recursively handle all succeeding occurrences. This solves the problem of nexted template-table-thingies.
			// This nesting is a problem because "{{BS-table}}" and "{|" are both starting tokens and "|}" is the only
			// end token. This cannot be handled by "findCorrespondingCloseToken()". Therefore, recursion is used here
			// to ensure that no "{{BS-table}}" occurs after the current sliding window.
			content = hackGermanRailwayTemplates(content, i+slidingWindowSize)

			// Find the end "|}" to remove the both lines
			endIndex := findCorrespondingCloseToken(content, i+slidingWindowSize, startToken, endToken)
			if endIndex == -1 {
				// TODO return an error instead and do not ignore anything!
				sigolo.Error("Found %s but no corresponding %s. I'll ignore this but something's wrong with the input wikitext!", startToken, endToken)
				return content
			}

			// Remove end token
			contentBeforeEndToken := content[0:endIndex]
			contentAfterEndToken := content[endIndex+endTokenSize:]
			content = contentBeforeEndToken + contentAfterEndToken

			// Remove start token
			contentBeforeWindow := content[0:i]
			contentAfterWindow := content[i+slidingWindowSize:]
			content = contentBeforeWindow + contentAfterWindow
		}
	}

	return content
}
