package parser

import (
	"fmt"
	"github.com/hauke96/wiki2book/src/util"
	"sort"
	"strings"
)

func (t *Tokenizer) parseReferences(content string) string {
	/*
		Idea of this parsing step:

		0. Split the content into a head and foot part (head is above the reference list, foot below it)
		1. Collect reference definitions
			1.1. Collect all named references definitions like <ref name="foo">bar</ref> and replace them by usages like <ref name="foo" />
			1.2. Collect all unnamed definitions like <ref>foo</ref> and replace them by usages like <ref name="foo" />
		2. Get a sorted list of all such usages to assign numbers (reference counter value) to them. This number will be shown to the user, that's why it needs to be sorted.
		3. Generate tokens for reference usages and definitions. The definitions will be between head and foot.
	*/

	referenceDefinitions := map[string]string{}

	// Step 0
	head, foot, content, noRefListFound := t.getReferenceHeadAndFoot(content)
	if noRefListFound {
		return content
	}

	// Step 1
	// Step 1.1
	// For definition <ref name="...">...</ref>   -> create usage with the given name
	head = t.replaceNamedReferences(content, referenceDefinitions, head)

	// Step 1.2
	// For definition <ref>...</ref>   -> create usage with a new random name
	head = t.replaceUnnamedReferences(content, referenceDefinitions, head)

	// Step 2
	// For usage <ref name="..." />
	nameToReference, contentIndexToRefName := t.getReferenceUsages(head)

	sortedRefNames, refNameToIndex := t.getSortedReferenceNames(contentIndexToRefName)

	// Step 3
	// Create usage token for ref usages like <ref name="foo" />
	for _, name := range sortedRefNames {
		ref := nameToReference[name]
		refIndex := refNameToIndex[name]
		token := t.getToken(TOKEN_REF_USAGE)
		t.setToken(token, fmt.Sprintf("%d %s", refIndex, ref))
		head = strings.ReplaceAll(head, ref, token)
	}

	// Append ref definitions to head
	for _, name := range sortedRefNames {
		ref := referenceDefinitions[name]
		token := t.getToken(TOKEN_REF_DEF)
		t.setToken(token, fmt.Sprintf("%d %s", refNameToIndex[name], t.tokenizeContent(t, ref)))
		head += token + "\n"
	}

	return head + foot
}

// getReferenceHeadAndFoot splits the content into section before, at and after the reference list.
// The return values are head, foot, content and a boolean which is true if there's no reference list in content.
func (t *Tokenizer) getReferenceHeadAndFoot(content string) (string, string, string, bool) {
	// No reference list found -> abort
	if !referenceBlockStartRegex.MatchString(content) {
		return "", "", content, true
	}

	contentParts := referenceBlockStartRegex.Split(content, -1)
	// In case of dedicated <references>...</references> block
	//   part 0 = head   : everything before <references...>
	//   part 1 (ignored): everything between <references> and </references>
	//   part 2 = foot   : everything after </references>
	// In case of <references/>
	//   part 0 = head: everything before <references/>
	//   part 1 = foot: everything after <references/>
	// Completely remove the reference section as we already parsed it above with the regex.
	head := contentParts[0]
	foot := ""
	if len(contentParts) == 2 {
		foot = contentParts[1]
	} else if len(contentParts) == 3 {
		foot = contentParts[2]
	}
	return head, foot, content, false
}

// getSortedReferenceNames gets all reference names sorted by their occurrence and a map from name to an index (the occurrence counter).
func (t *Tokenizer) getSortedReferenceNames(indexToRefName map[int]string) ([]string, map[string]int) {
	refNameToIndex := map[string]int{}

	referenceIndices := make([]int, 0, len(indexToRefName))
	for key := range indexToRefName {
		referenceIndices = append(referenceIndices, key)
	}
	sort.Ints(referenceIndices)

	// Assign increasing index to each reference based on their occurrence in "content"
	refCounter := 1
	var sortedRefNames []string
	for _, refIndex := range referenceIndices {
		refName := indexToRefName[refIndex]
		refNameToIndex[refName] = refCounter
		sortedRefNames = append(sortedRefNames, refName)
		refCounter++
	}

	return sortedRefNames, refNameToIndex
}

// replaceNamedReferences replaces all occurrences of named reference definitions in "head" by a named reference usage.
func (t *Tokenizer) replaceNamedReferences(content string, nameToRefDef map[string]string, head string) string {
	// Go through "content" to also parse the definitions inside the <references>...</references> block and below it.
	submatches := namedReferenceRegex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		name := submatch[1]
		totalRef := submatch[0]
		nameToRefDef[name] = totalRef
		head = strings.ReplaceAll(head, totalRef, fmt.Sprintf("<ref name=\"%s\" />", name))
	}
	return head
}

// replaceNamedReferences replaces all occurrences of unnamed reference definitions by a named reference usage with a random reference name.
func (t *Tokenizer) replaceUnnamedReferences(content string, nameToRefDef map[string]string, head string) string {
	// Go throught "content" to also parse the definitions in the reference section below "head"
	submatches := unnamedReferenceRegex.FindAllStringSubmatch(content, -1)
	for i, submatch := range submatches {
		totalRef := submatch[0]

		// Ignore named references, we're just interested in UNnamed ones
		if namedReferenceWithoutGroupsRegex.MatchString(totalRef) {
			continue
		}

		// Generate more or less random but unique name
		name := util.Hash(fmt.Sprintf("%d%s", i, totalRef))

		nameToRefDef[name] = totalRef
		head = strings.ReplaceAll(head, totalRef, fmt.Sprintf("<ref name=\"%s\" />", name))
	}
	return head
}

// getReferenceUsages gets all reference usages (name to total reference) as well as a map that maps the reference counter to the reference name.
func (t *Tokenizer) getReferenceUsages(head string) (map[string]string, map[int]string) {
	// This map maps the reference name to the actual wikitext of that reference
	nameToRefDef := map[string]string{}
	// This map take the index of the reference in "content" as determined by  strings.Index()  as key/value.
	indexToRefName := map[int]string{}

	submatches := namedReferenceUsageRegex.FindAllStringSubmatch(head, -1)
	for _, submatch := range submatches {
		name := submatch[1]
		nameToRefDef[name] = submatch[0]
		indexToRefName[strings.Index(head, submatch[0])] = name
	}

	return nameToRefDef, indexToRefName
}
