package parser

import (
	"fmt"
	"testing"
	"wiki2book/test"
)

func TestGetReferenceHeadAndFoot(t *testing.T) {
	head := "some text<ref>foo</ref> with refs<ref name=\"barbar\">bar</ref>.\n"
	foot := "foooooooter"
	content := head + "<references />\n" + foot

	tokenizer := NewTokenizer("foo", "bar")
	newHead, newFoot, newContent, noRefListFound := tokenizer.getReferenceHeadAndFoot(content)

	test.AssertEqual(t, head, newHead)
	test.AssertEqual(t, foot, newFoot)
	test.AssertEqual(t, content, newContent)
	test.AssertFalse(t, noRefListFound)
}

func TestGetSortedReferenceNames(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	refIndexToName := map[int]string{
		3: "bar",
		1: "foo",
		2: "blubb",
	}

	sortedName, refNameToIndex := tokenizer.getSortedReferenceNames(refIndexToName)

	test.AssertEqual(t, []string{"foo", "blubb", "bar"}, sortedName)
	test.AssertMapEqual(t, map[string]int{
		"foo":   1,
		"blubb": 2,
		"bar":   3,
	}, refNameToIndex)
}

func TestReplaceNamedReferences(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	head := `some<ref>foo</ref> text with refs<ref name="barbar">bar</ref> some <ref name=other-name>other</ref> <ref>foo</ref>`
	content := head + ` and even more<ref>blubb</ref>text`
	referenceDefinitions := map[string]string{}
	newHead := tokenizer.replaceNamedReferences(content, referenceDefinitions, head)

	test.AssertMapEqual(t, map[string]string{
		"barbar":     `bar`,
		"other-name": `other`,
	}, referenceDefinitions)
	test.AssertEqual(t, `some<ref>foo</ref> text with refs<ref name="barbar" /> some <ref name="other-name" /> <ref>foo</ref>`, newHead)
}

func TestReplaceNamedReferences_withSpecialCharacters(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	head := `Refs can contain special chars <ref name="foo/bar.blubb">bar</ref>.`
	content := head + ` and even more<ref>blubb</ref>text`
	referenceDefinitions := map[string]string{}
	newHead := tokenizer.replaceNamedReferences(content, referenceDefinitions, head)

	test.AssertMapEqual(t, map[string]string{
		"foo/bar.blubb": `bar`,
	}, referenceDefinitions)
	test.AssertEqual(t, `Refs can contain special chars <ref name="foo/bar.blubb" />.`, newHead)
}

func TestReplaceUnnamedReferences(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	head := `some<ref some="param">foo</ref> text with refs<ref name="barbar">bar</ref> and <ref some="sla/shes">slashes</ref>`
	content := head + ` and even more<ref>blubb</ref>text`
	referenceDefinitions := map[string]string{}
	newHead := tokenizer.replaceUnnamedReferences(content, referenceDefinitions, head)

	test.AssertMapEqual(t, map[string]string{
		"2ae457b665ef5955b2fc685cdaaa879c96c14801": `foo`,
		"eeada6edccd48f48f3d8c8968c1878a994cbf23e": `slashes`,
		"74e7903564d066a6c4c76d9c0b9835938d0ae829": "blubb",
	}, referenceDefinitions)
	test.AssertEqual(t, `some<ref name="2ae457b665ef5955b2fc685cdaaa879c96c14801" /> text with refs<ref name="barbar">bar</ref> and <ref name="eeada6edccd48f48f3d8c8968c1878a994cbf23e" />`, newHead)
}

func TestReplaceUnnamedReferences_ignoreReferenceUsages(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `foo <ref name="refname" /> bar <ref some="barbar">bar</ref>`
	referenceDefinitions := map[string]string{}
	newHead := tokenizer.replaceUnnamedReferences(content, referenceDefinitions, content)

	test.AssertMapEqual(t, map[string]string{
		"c8d3521ed18935eb577600c6c0e9fd278b296264": `bar`,
	}, referenceDefinitions)
	test.AssertEqual(t, `foo <ref name="refname" /> bar <ref name="c8d3521ed18935eb577600c6c0e9fd278b296264" />`, newHead)
}

func TestGetReferenceUsages(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `some<ref name="bar" /> text with refs<ref name="foo" />`
	usages, _ := tokenizer.getReferenceUsages(content)

	test.AssertMapEqual(t, map[string]string{
		"bar": `<ref name="bar" />`,
		"foo": `<ref name="foo" />`,
	}, usages)
}

func TestParseReferences(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `some text<ref>bar</ref>
some<ref name="blubb">blubbeldy</ref> other<ref name="fooref" /> text
<references responsive>
<ref name="fooref">foo</ref>
</references>
some footer`
	expectedContent := "some text" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 0) + "\n" +
		"some" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 1) + " other" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 2) + " text\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 3) + "\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 4) + "\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 5) + "\n" +
		"some footer"

	newContent := tokenizer.parseReferences(content)

	test.AssertEqual(t, expectedContent, newContent)
	test.AssertMapEqual(t, map[string]Token{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 0): RefUsageToken{Index: 1},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 1): RefUsageToken{Index: 2},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 2): RefUsageToken{Index: 3},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 3):   RefDefinitionToken{Index: 1, Content: "bar"},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 4):   RefDefinitionToken{Index: 2, Content: "blubbeldy"},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 5):   RefDefinitionToken{Index: 3, Content: "foo"},
	}, tokenizer.getTokenMap())
}
