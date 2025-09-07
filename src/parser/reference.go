package parser

import (
	"strings"
	"wiki2book/util"

	"github.com/hauke96/sigolo/v2"
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

// This is the default group in which all ungrouped references fall
const defaultReferenceGroup = "__wiki2book_ungrouped_references_group__"

func (t *Tokenizer) parseReferences(content string) string {
	/*
		There are two types of tags we parse here: Reference definitions and references placeholders. Definitions look
		like "<ref ...>...</ref>" and placeholders like "<references />". There are also usages of already defined
		references, which have a "name" tag in them like "<ref name="foo" />".

		Structure and wording for reference definitions:

		<ref name="foo" > Foo </ref>
		|-------A------|B|-C-|--D--|

		A = Start
		B = Start closing
		C = Content
		D = End

		For reference usages like "<ref name="foo" />", there is no start closing and the reference end is "/>".
	*/
	refDefStart := "<ref"
	refDefLongEnd := "</ref>"
	refPlaceholderEnd := "</re" // Only four characters just as the refDefStart, which defined the cursor size.
	xmlClosing := ">"
	refDefStartLen := len(refDefStart)
	refDefLongEndLen := len(refDefLongEnd)

	// Store content of references (i.e. their actual text), ref-name to ref-number mapping and the ref-number counters
	// all per group. Every ref without explicit group is part of the default group.
	refNumberToContent := map[string]map[int]string{}
	nameToRefNumber := map[string]map[string]int{}
	refNumberCounter := map[string]int{}

	// Whether the current cursor is within a "<references>...</references>" block. Within this block, further reference
	// definitions might occur. These references will not be turned into any usage-token because they are not used at
	// that location but just defined.
	cursorWithinReferencePlaceholder := false
	// The group of the current placeholder, e.g. "foo" for "<references group=foo>...".
	currentPlaceholderGroup := ""

	for i := 0; i < len(content)-refDefStartLen; i++ {
		cursor := content[i : i+refDefStartLen]
		if cursor != refDefStart && cursor != refPlaceholderEnd {
			// Cursor is not on the beginning of any reference related tag.
			continue
		}

		startEndIndex := findCorrespondingCloseToken(content, i+refDefStartLen, refDefStart, xmlClosing)
		if startEndIndex == -1 {
			// XML for <ref not closed -> broken wikitext
			sigolo.Errorf("XML element for reference start '%s' not closed (i.e. missing '%s'). Text around this location: ...%s...", refDefStart, xmlClosing, util.GetTextAround(content, i, 50))
		}

		if referencePlaceholderEndRegex.MatchString(content[i:startEndIndex+1]) || referencePlaceholderShortRegex.MatchString(content[i:startEndIndex+1]) {
			// Tag like "</references>" or "<references />" found
			if currentPlaceholderGroup == "" {
				currentPlaceholderGroup = t.getGroupOrDefault(content[i:startEndIndex])
			}

			if refNumberToContent[currentPlaceholderGroup] == nil {
				refNumberToContent[currentPlaceholderGroup] = map[int]string{}
			}

			refNumberCounterForCurrentGroup := refNumberCounter[currentPlaceholderGroup]
			refNumberToContentForCurrentGroup := refNumberToContent[currentPlaceholderGroup]

			content = t.parseReferenceEndPlaceholder(content, i, startEndIndex, refNumberCounterForCurrentGroup, refNumberToContentForCurrentGroup)
			cursorWithinReferencePlaceholder = false
			currentPlaceholderGroup = ""
		} else if referencePlaceholderStartRegex.MatchString(content[i : startEndIndex+1]) {
			// Tag like "<references group=foo >" found
			// TODO indices correct?
			currentPlaceholderGroup = t.getGroupOrDefault(content[i:startEndIndex])
			content = content[0:i] + content[startEndIndex+1:] // Remove tag from content
			cursorWithinReferencePlaceholder = true
		} else {
			// Tag like "<ref name=..." or "<ref>..." found
			nameAttributeValue := t.getNameAttribute(content[i+refDefStartLen : startEndIndex])
			groupName := t.getGroupOrDefault(content[i+refDefStartLen : startEndIndex])

			if nameToRefNumber[groupName] == nil {
				nameToRefNumber[groupName] = map[string]int{}
			}

			if refNumberToContent[groupName] == nil {
				refNumberToContent[groupName] = map[int]string{}
			}

			nameToRefNumberForCurrentGroup := nameToRefNumber[groupName]
			refNumberToContentForCurrentGroup := refNumberToContent[groupName]
			refNumberCounterForCurrentGroup := refNumberCounter[groupName]

			isReferenceUsage := content[startEndIndex-1] == '/' // Reference definitions end with "/>" instead of "</ref>"
			if isReferenceUsage {
				// Reference usage like "<ref name=foo />"
				refNumberCounterForCurrentGroup, content = t.parseNamedReferenceUsage(content, i, nameAttributeValue, nameToRefNumberForCurrentGroup, refNumberCounterForCurrentGroup, cursorWithinReferencePlaceholder, startEndIndex)
			} else {
				// Reference definition like "<ref name=...>Foobar</ref".
				refEndIndex := findCorrespondingCloseToken(content, startEndIndex, refDefStart, refDefLongEnd)
				if refEndIndex == -1 {
					// No end token found -> probably unsupported wikitext syntax (like nested refs)
					sigolo.Errorf("No end-part for reference start '%s' found. Text around this location: ...%s...", refDefStart, util.GetTextAround(content, i, 50))
				}
				refNumberCounterForCurrentGroup, content = t.parseReferenceDefinition(content, i, startEndIndex, refEndIndex, refNumberCounterForCurrentGroup, nameAttributeValue, nameToRefNumberForCurrentGroup, refNumberToContentForCurrentGroup, cursorWithinReferencePlaceholder, refDefLongEndLen)
			}

			refNumberCounter[groupName] = refNumberCounterForCurrentGroup
		}
	}

	return content
}

// parseReferenceEndPlaceholder replaces the end of the given reference placeholder at index i, such as "<references />"
// or "</references>" with a list of all references that occurred so far. It removes elements from the
// refNumberToContent map. It returns the new content in which the reference end-token has been replaces by a newline-
// separated list of reference definition token.
func (t *Tokenizer) parseReferenceEndPlaceholder(content string, i int, startEndIndex int, refNumberCounter int, refNumberToContent map[int]string) string {
	// Remove tag from content
	contentBefore := strings.TrimRight(content[0:i], "\n") + "\n" // ensure this part ends with a newline
	contentAfter := content[startEndIndex+1:]

	// Generate list of references

	for refNumber := 0; refNumber < refNumberCounter; refNumber++ {
		if _, ok := refNumberToContent[refNumber]; !ok {
			continue
		}

		tokenKey := t.getToken(TOKEN_REF_DEF)
		t.setRawToken(tokenKey, RefDefinitionToken{
			Index:   refNumber,
			Content: refNumberToContent[refNumber],
		})
		contentBefore += tokenKey + "\n"

		// Delete entry to prevent it from being used at the next placeholder again.
		delete(refNumberToContent, refNumber)
	}

	content = strings.TrimRight(contentBefore, "\n") + contentAfter
	return content
}

// parseNamedReferenceUsage replaces the occurrence of a named reference usage, such as "<ref name=foo />" at the given
// index i with a reference usage token. It might increase the refNumberCounter, in case the reference appeared for the
// first time, might change the nameToRefNumber map and returns the new content containing the key of the new reference
// usage token.
func (t *Tokenizer) parseNamedReferenceUsage(content string, i int, nameAttributeValue string, nameToRefNumber map[string]int, refNumberCounter int, cursorWithinReferencePlaceholder bool, startEndIndex int) (int, string) {
	if nameAttributeValue != "" {
		// Names reference usage
		refNumber, ok := nameToRefNumber[nameAttributeValue]
		if !ok {
			// Name appears the first time, the definition might come later
			refNumber = refNumberCounter
			nameToRefNumber[nameAttributeValue] = refNumber
			refNumberCounter++
		}

		if !cursorWithinReferencePlaceholder {
			tokenKey := t.getToken(TOKEN_REF_USAGE)
			t.setRawToken(tokenKey, RefUsageToken{
				Index: refNumber,
			})
			content = content[0:i] + tokenKey + content[startEndIndex+1:]
		} else {
			content = content[0:i] + content[startEndIndex+1:]
		}
	} else {
		sigolo.Warnf("Named reference usage without name-attribute found: %s", content[i:startEndIndex])
	}
	return refNumberCounter, content
}

// parseReferenceDefinition replaces the occurrence of a reference definition, such as "<ref>foo</ref>" at the given
// index i which a reference usage token. The definition and its content is stored to the refNumberToContent map. In
// case the reference definition has a name attribute, an entry is added to the nameToRefNumber. When the reference is
// new, the refNumberCounter will be incremented and its new value returned. In case of an already known named reference,
// this counter will not change and its current value will be returned.
func (t *Tokenizer) parseReferenceDefinition(content string, i int, startEndIndex int, refEndIndex int, refNumberCounter int, nameAttributeValue string, nameToRefNumber map[string]int, refNumberToContent map[int]string, cursorWithinReferencePlaceholder bool, refDefLongEndLen int) (int, string) {
	refNumber := refNumberCounter
	if nameAttributeValue != "" {
		if _, ok := nameToRefNumber[nameAttributeValue]; ok {
			// Ref name already used before, so we use the number of this existing ref usage.
			refNumber = nameToRefNumber[nameAttributeValue]
		} else {
			// Ref name appears for the first time, so we save the current counter value for later
			// usages of this ref name.
			nameToRefNumber[nameAttributeValue] = refNumberCounter
		}
	}

	refNumberToContent[refNumber] = t.tokenizeContent(t, content[startEndIndex+1:refEndIndex])

	if !cursorWithinReferencePlaceholder {
		tokenKey := t.getToken(TOKEN_REF_USAGE)
		t.setRawToken(tokenKey, RefUsageToken{
			Index: refNumber,
		})
		content = content[0:i] + tokenKey + content[refEndIndex+refDefLongEndLen:]
	} else {
		content = content[0:i] + content[refEndIndex+refDefLongEndLen:]
	}

	if refNumber == refNumberCounter {
		// We actually used the current count value, so we increase it for the next token.
		refNumberCounter++
	}

	return refNumberCounter, content
}

func (t *Tokenizer) getNameAttribute(content string) string {
	return t.getAttribute(content, "name")
}

func (t *Tokenizer) getGroupOrDefault(content string) string {
	groupAttributeValue := t.getAttribute(content, "group")
	if groupAttributeValue != "" {
		return groupAttributeValue
	}
	return defaultReferenceGroup
}

// getAttribute determines the values after "{attributeName}=" (so e.g. "name=") and supports quoted and unquoted
// attributes. When unquoted attributes are used (e.g. as in name=foobar), the value is only interpreted until a space
// of slash. For quoted attributes (e.g. as in name="foo bar") everything until the next quote is interpreted as name value.
func (t *Tokenizer) getAttribute(content string, attributeName string) string {
	if strings.Contains(content, " "+attributeName+"=\"") {
		// Attribute with quotation
		parts := strings.Split(content, "\"")
		for i, part := range parts {
			if strings.HasSuffix(part, " "+attributeName+"=") {
				// Found attribute key, next item is the value which can be returned
				return parts[i+1]
			}
		}
	} else if strings.Contains(content, " "+attributeName+"=") {
		// Attribute without quotation like "name=foo"
		// The value ends with a space (separator for additional attributes) or slash (and of <ref.../>-token).
		parts := strings.SplitN(content, " "+attributeName+"=", 2)
		var letter rune
		var attributeValue string
		for _, letter = range parts[1] {
			if letter == '/' || letter == ' ' || letter == '>' {
				break
			}
			attributeValue += string(letter)
		}
		return attributeValue
	}

	// Attribut not found
	return ""
}
