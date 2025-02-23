package parser

import (
	"fmt"
	"testing"
	"wiki2book/test"
)

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
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 0): RefUsageToken{Index: 0},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 1): RefUsageToken{Index: 1},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 2): RefUsageToken{Index: 2},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 3):   RefDefinitionToken{Index: 0, Content: "bar"},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 4):   RefDefinitionToken{Index: 1, Content: "blubbeldy"},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 5):   RefDefinitionToken{Index: 2, Content: "foo"},
	}, tokenizer.getTokenMap())
}

func TestParseReferences_tokenizeRefContent(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `some text<ref>foo [[bar|Bar]]</ref>.`
	expectedContent := "some text" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 1) + "."

	newContent := tokenizer.parseReferences(content)

	test.AssertEqual(t, expectedContent, newContent)
	test.AssertMapEqual(t, map[string]Token{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_INTERNAL_LINK, 0): InternalLinkToken{ArticleName: "bar", LinkText: "Bar"},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 1):     RefUsageToken{Index: 0},
	}, tokenizer.getTokenMap())
}

func TestParseReferences_mixedQuotations(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `Foo<ref name="foo">This is a ref for foo.</ref>
Bar<ref name=bar>This is a quoteless ref for bar.</ref>
<references/>`
	expectedContent := "Foo" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 0) + "\n" +
		"Bar" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 1) + "\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 2) + "\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 3)

	newContent := tokenizer.parseReferences(content)

	test.AssertEqual(t, expectedContent, newContent)
	test.AssertMapEqual(t, map[string]Token{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 0): RefUsageToken{Index: 0},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 1): RefUsageToken{Index: 1},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 2):   RefDefinitionToken{Index: 0, Content: "This is a ref for foo."},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 3):   RefDefinitionToken{Index: 1, Content: "This is a quoteless ref for bar."},
	}, tokenizer.getTokenMap())
}

func TestParseReferences_multipleUsagesOfRefName(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `Foo<ref name="foo">some ref</ref>
Bar<ref name=foo />
Foobar<ref name="foo"" />
<references/>`
	expectedContent := "Foo" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 0) + "\n" +
		"Bar" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 1) + "\n" +
		"Foobar" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 2) + "\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 3)

	newContent := tokenizer.parseReferences(content)

	test.AssertEqual(t, expectedContent, newContent)
	test.AssertMapEqual(t, map[string]Token{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 0): RefUsageToken{Index: 0},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 1): RefUsageToken{Index: 0},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 2): RefUsageToken{Index: 0},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 3):   RefDefinitionToken{Index: 0, Content: "some ref"},
	}, tokenizer.getTokenMap())
}

func TestParseReferences_multipleRefBodyDefinitions(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `Foo<ref name="foo">some ref</ref>
Bar<ref name=foo>some ref but for bar</ref>
<references/>`
	expectedContent := "Foo" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 0) + "\n" +
		"Bar" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 1) + "\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 2)

	newContent := tokenizer.parseReferences(content)

	test.AssertEqual(t, expectedContent, newContent)
	test.AssertMapEqual(t, map[string]Token{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 0): RefUsageToken{Index: 0},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 1): RefUsageToken{Index: 0},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 2):   RefDefinitionToken{Index: 0, Content: "some ref but for bar"},
	}, tokenizer.getTokenMap())
}

func TestParseReferences_multipleReferencesTags(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `Foo<ref>some ref</ref>
<references/>
Bar<ref>some other ref</ref>
<references/>`
	expectedContent := "Foo" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 0) + "\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 1) + "\n" +
		"Bar" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 2) + "\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 3)

	newContent := tokenizer.parseReferences(content)

	test.AssertEqual(t, expectedContent, newContent)
	test.AssertMapEqual(t, map[string]Token{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 0): RefUsageToken{Index: 0},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 1):   RefDefinitionToken{Index: 0, Content: "some ref"},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 2): RefUsageToken{Index: 1},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 3):   RefDefinitionToken{Index: 1, Content: "some other ref"},
	}, tokenizer.getTokenMap())
}

func TestGetNameAttribute(t *testing.T) {
	content := "some name=foo bar"
	name := getNameAttribute(content)
	test.AssertEqual(t, "foo", name)

	content = "some name=\"foo\" bar"
	name = getNameAttribute(content)
	test.AssertEqual(t, "foo", name)

	content = "some name=foo"
	name = getNameAttribute(content)
	test.AssertEqual(t, "foo", name)

	content = "some name=\"foo\""
	name = getNameAttribute(content)
	test.AssertEqual(t, "foo", name)

	content = "some <ref name=\"foo\"> bar"
	name = getNameAttribute(content)
	test.AssertEqual(t, "foo", name)

	content = "some <ref name=foo> bar"
	name = getNameAttribute(content)
	test.AssertEqual(t, "foo", name)

	content = "some noname=foo bar"
	name = getNameAttribute(content)
	test.AssertEqual(t, "", name)

	content = "some noname=\"foo\" bar"
	name = getNameAttribute(content)
	test.AssertEqual(t, "", name)
}
