package parser

import (
	"fmt"
	"github.com/hauke96/wiki2book/src/test"
	"testing"
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
	test.AssertEqual(t, map[string]int{
		"foo":   1,
		"blubb": 2,
		"bar":   3,
	}, refNameToIndex)
}

func TestReplaceNamedReferences(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	head := `some<ref>foo</ref> text with refs<ref name="barbar">bar</ref> some other <ref>foo</ref>`
	content := head + ` and even more<ref>blubb</ref>text`
	referenceDefinitions := map[string]string{}
	newHead := tokenizer.replaceNamedReferences(content, referenceDefinitions, head)

	test.AssertMapEqual(t, map[string]string{
		"barbar": `<ref name="barbar">bar</ref>`,
	}, referenceDefinitions)
	test.AssertEqual(t, `some<ref>foo</ref> text with refs<ref name="barbar" /> some other <ref>foo</ref>`, newHead)
}

func TestReplaceUnnamedReferences(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	head := `some<ref some="param">foo</ref> text with refs<ref name="barbar">bar</ref>`
	content := head + ` and even more<ref>blubb</ref>text`
	referenceDefinitions := map[string]string{}
	newHead := tokenizer.replaceUnnamedReferences(content, referenceDefinitions, head)

	test.AssertMapEqual(t, map[string]string{
		"2ae457b665ef5955b2fc685cdaaa879c96c14801": `<ref some="param">foo</ref>`,
		"2839c654c615b6833625f76f38b609c71b74ada4": "<ref>blubb</ref>",
	}, referenceDefinitions)
	test.AssertEqual(t, `some<ref name="2ae457b665ef5955b2fc685cdaaa879c96c14801" /> text with refs<ref name="barbar">bar</ref>`, newHead)
}

func TestGetReferenceUsages(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `some<ref name="bar" /> text with refs<ref name="foo" />`
	usages, _ := tokenizer.getReferenceUsages(content)

	test.AssertEqual(t, map[string]string{
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
}
