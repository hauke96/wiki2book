package parser

import (
	"fmt"
	"strings"
	"wiki2book/util"
)

type RefDefinitionToken struct {
	Token
	Index   int
	Content string
}

type RefUsageToken struct {
	Token
	Index int
}

func (t *Tokenizer) parseReferences(content string) string {
	// Split content in head, body and foot.
	// Head is everything before the <references/> or </references> tag.
	// Body is everything between <references> and </references>.
	// Foot is everything after <references/> or </references>.
	head, body, foot, noRefListFound := t.getReferenceHeadBodyFoot(content)
	if noRefListFound {
		return content
	}

	refIndexToContent := map[int]string{}
	nameToRefIndex := map[string]int{}
	refIndexCounter := 0

	refIndexCounter, newHead := t.parseStringForReferences(head, nameToRefIndex, refIndexToContent, refIndexCounter, true)
	refIndexCounter, _ = t.parseStringForReferences(body, nameToRefIndex, refIndexToContent, refIndexCounter, false)

	// Append ref definitions to head
	for refIndex := 0; refIndex < refIndexCounter; refIndex++ {
		tokenKey := t.getToken(TOKEN_REF_DEF)
		t.setRawToken(tokenKey, RefDefinitionToken{
			Index:   refIndex,
			Content: t.tokenizeContent(t, refIndexToContent[refIndex]),
		})
		newHead += tokenKey + "\n"
	}

	return newHead + foot
}

// parseStringForReferences determines all reference usage and definition in the given text and stores that information
// in the given maps. The appendTokenToContent flag determines whether tokens should be generated and appended
// to the output string. The output int is the next (and not yet used ref index), the output string is the parsed
// input with removed refs and optionally with token keys in it.
func (t *Tokenizer) parseStringForReferences(stringToParse string, nameToRefIndex map[string]int, refIndexToContent map[int]string, refIndexCounter int, appendTokenToContent bool) (int, string) {
	/*
		Strategy:

		Go through stringToParse and search for all occurrences of any reference stuff (usage, usage with name, usage
		with definition). Then remember all the definitions and usages in their order to create correct numbering and
		footnote entries later.

		This is done by first splitting the content by "<ref" to obtain segments which all start with a reference. There
		is some additional processing needed to only get relevant lines. The <reference/> tag also starts with <ref but
		should not be considered. There might be more tags starting with <ref.

		Having all truly valid reference segments, they are parsed to obtain the name (<ref name="some name"...>) and the
		content/body of the reference (<ref>some content</ref>).

		So later join all segments back together into one string with tokens in it, the given maps and segmentToRefIndex
		are used to store this information which reference with which index belongs to which segment.

		In the end (if wanted, s. appendTokenToContent flag), tokens are created and the segments are joined back
		together.

		Some additional information:

		Cases of reference definitions:
		<ref>foo</ref>
		<ref name="something">foo</ref>

		Case of reference usage:
		<ref name="something" />
	*/

	contentSegments := strings.Split(stringToParse, "<ref")

	// Determine segments that actually start with a reference.
	validContentSegments := []string{contentSegments[0]}
	for i, segment := range contentSegments {
		if i == 0 {
			// First element is not a ref (it's in front of the first ref) and is already added above.
			continue
		} else if !strings.HasPrefix(segment, " ") && !strings.HasPrefix(segment, ">") {
			// Segment starts with "<ref" but is not a reference -> merge with prior segment
			validContentSegments[len(validContentSegments)-1] = validContentSegments[len(validContentSegments)-1] + "<ref" + segment
		} else {
			validContentSegments = append(validContentSegments, segment)
		}
	}

	contentSegments = validContentSegments

	// This map stores the reference index per segment. A reference can be used multiple times, so an index might be
	// used in multiple segments.
	segmentToRefIndex := map[int]int{}

	for i, segment := range contentSegments {
		if i == 0 {
			// First element of string.Split() doesn't contain a ref (it's located before the first ref)
			continue
		}

		// Try to find normal ref definition
		// 0 = everything before </ref> (so everything between <ref... and ...</ref>)
		// 1 = everything after </ref>
		segmentParts := strings.SplitN(segment, "</ref>", 2)

		// <ref refAttributes>refContent</ref>contentAfterRef
		var refAttributes string
		var refContent string
		var contentAfterRef string //

		if len(segmentParts) == 2 {
			contentAfterRef = segmentParts[1]
			split := strings.SplitN(segmentParts[0], ">", 2)
			refAttributes = split[0]
			refContent = split[1]
		} else if len(segmentParts) < 2 {
			// Found ref usage instead, so we have to split a bit differently
			// 0 = everything before /> (so everything between <ref... and .../>)
			// 1 = everything after />
			segmentParts = strings.SplitN(segment, "/>", 2)
			refAttributes = segmentParts[0]
			contentAfterRef = segmentParts[1]
		}

		refName := getNameAttribute(refAttributes)
		if refName == "" {
			// Nameless ref found -> Create randomized ref name
			refName = util.Hash(fmt.Sprintf("%d%s", i, segment))
		}

		// Store only the part without the </ref> and /> snippets. The element [1] contains the content after the ref,
		// so it's easier for later to join all the segments back together into a new string with reference token in it.
		contentSegments[i] = contentAfterRef

		if existingRefIndex, ok := nameToRefIndex[refName]; ok {
			// Ref with same refName was defined earlier, so we reuse the ref index counter value
			segmentToRefIndex[i] = existingRefIndex
			if refContent != "" {
				refIndexToContent[existingRefIndex] = refContent
			}
		} else {
			nameToRefIndex[refName] = refIndexCounter
			segmentToRefIndex[i] = refIndexCounter
			refIndexToContent[refIndexCounter] = refContent
			refIndexCounter++
		}
	}

	if !appendTokenToContent {
		return refIndexCounter, ""
	}

	// The first segment if before the first ref and can therefore be used unchanged.
	newHead := contentSegments[0]

	for i := 1; i <= len(segmentToRefIndex); i++ {
		refIndex := segmentToRefIndex[i]
		tokenKey := t.getToken(TOKEN_REF_USAGE)
		t.setRawToken(tokenKey, RefUsageToken{
			Index: refIndex,
		})
		newHead += tokenKey + contentSegments[i] // +1 because segment 0 was already added above
	}

	return refIndexCounter, newHead
}

// getNameAttribute determines the values after "name=" and supports quoted and unquoted attributes. When unquoted
// attributes are used (e.g. as in name=foobar), the value is only interpreted until a space of slash. For quoted
// attributes (e.g. as in name="foo bar") everything until the next quote is interpreted as name value.
func getNameAttribute(content string) string {
	if strings.Contains(content, " name=\"") {
		// Name with quotation
		parts := strings.Split(content, "\"")
		for i, part := range parts {
			if strings.HasSuffix(part, " name=") {
				// Found name key, next item is the value which can be returned
				return parts[i+1]
			}
		}
	} else if strings.Contains(content, " name=") {
		// Name without quotation like name=foo
		// The value ends with a space (separator for additional attributes) or slash (and of <ref.../>-token).
		parts := strings.SplitN(content, " name=", 2)
		var letter rune
		var resultName string
		for _, letter = range []rune(parts[1]) {
			if letter == '/' || letter == ' ' || letter == '>' {
				break
			}
			resultName += string(letter)
		}
		return resultName
	}

	// No name found
	return ""
}

// getReferenceHeadBodyFoot splits the content into section before, within and after the reference list tag. The return
// values are head, body, foot and a boolean which is true if there's no reference list in content.
func (t *Tokenizer) getReferenceHeadBodyFoot(content string) (string, string, string, bool) {
	// No reference list found -> abort
	if !referenceBlockStartRegex.MatchString(content) {
		return "", "", "", true
	}

	contentParts := referenceBlockStartRegex.Split(content, -1)
	// In case of dedicated <references>...</references> block
	//   part 0 = head : everything before <references...>
	//   part 1 = body : everything between <references> and </references>
	//   part 2 = foot : everything after </references>
	// In case of <references/>
	//   part 0 = head: everything before <references/>
	//   part 1 = foot: everything after <references/>
	// Completely remove the reference section as we already parsed it above with the regex.
	head := contentParts[0]
	body := ""
	foot := ""
	if len(contentParts) == 2 {
		foot = contentParts[1]
	} else if len(contentParts) == 3 {
		body = contentParts[1]
		foot = contentParts[2]
	}
	return head, body, foot, false
}
