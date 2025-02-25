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

func TestParseReferences_multipleReferencesPlaceholder(t *testing.T) {
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

func TestParseReferences_multipleGroupedReferences(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `Foo<ref group="foo">some ref</ref>
Blubb<ref>some ungrouped ref</ref>
Bar<ref group="bar">some other ref</ref>

Foo references:
<references group="foo" />

Bar references:
<references group="bar" />

Other references:
<references/>`
	expectedContent := "Foo" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 0) + "\n" +
		"Blubb" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 1) + "\n" +
		"Bar" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 2) + "\n\n" +
		"Foo references:\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 3) + "\n\n" +
		"Bar references:\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 4) + "\n\n" +
		"Other references:\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 5)

	newContent := tokenizer.parseReferences(content)

	test.AssertEqual(t, expectedContent, newContent)
	test.AssertMapEqual(t, map[string]Token{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 0): RefUsageToken{Index: 0},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 1): RefUsageToken{Index: 0},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 2): RefUsageToken{Index: 0},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 3):   RefDefinitionToken{Index: 0, Content: "some ref"},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 4):   RefDefinitionToken{Index: 0, Content: "some other ref"},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 5):   RefDefinitionToken{Index: 0, Content: "some ungrouped ref"},
	}, tokenizer.getTokenMap())
}

func TestParseReferences_groupedAndNamedReferences(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `Foo<ref group="foo">some ref</ref>
Blubb<ref group="foo" name="some-name" />
Bar<ref group="foo" name="some-name">some grouped and named ref</ref>

Foo references:
<references group="foo" />`
	expectedContent := "Foo" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 0) + "\n" +
		"Blubb" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 1) + "\n" +
		"Bar" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 2) + "\n\n" +
		"Foo references:\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 3) + "\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 4)

	newContent := tokenizer.parseReferences(content)

	test.AssertEqual(t, expectedContent, newContent)
	test.AssertMapEqual(t, map[string]Token{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 0): RefUsageToken{Index: 0},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 1): RefUsageToken{Index: 1},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 2): RefUsageToken{Index: 1},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 3):   RefDefinitionToken{Index: 0, Content: "some ref"},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 4):   RefDefinitionToken{Index: 1, Content: "some grouped and named ref"},
	}, tokenizer.getTokenMap())
}

func TestParseReferences_multipleReferencesPerGroup(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `Foo<ref group="foo">some ref</ref>
Blubb<ref>some ungrouped ref</ref>
Bar<ref group="bar">some other ref</ref>

Foo2<ref group="foo">some ref2</ref>
Blubb2<ref>some ungrouped ref2</ref>
Bar2<ref group="bar">some other ref2</ref>

Foo references:
<references group="foo" />

Bar references:
<references group="bar" />

Other references:
<references/>`
	expectedContent := "Foo" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 0) + "\n" +
		"Blubb" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 1) + "\n" +
		"Bar" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 2) + "\n\n" +
		"Foo2" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 3) + "\n" +
		"Blubb2" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 4) + "\n" +
		"Bar2" + fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 5) + "\n\n" +
		"Foo references:\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 6) + "\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 7) + "\n\n" +
		"Bar references:\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 8) + "\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 9) + "\n\n" +
		"Other references:\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 10) + "\n" +
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 11)

	newContent := tokenizer.parseReferences(content)

	test.AssertEqual(t, expectedContent, newContent)
	test.AssertMapEqual(t, map[string]Token{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 0): RefUsageToken{Index: 0},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 1): RefUsageToken{Index: 0},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 2): RefUsageToken{Index: 0},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 3): RefUsageToken{Index: 1},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 4): RefUsageToken{Index: 1},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_USAGE, 5): RefUsageToken{Index: 1},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 6):   RefDefinitionToken{Index: 0, Content: "some ref"},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 7):   RefDefinitionToken{Index: 1, Content: "some ref2"},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 8):   RefDefinitionToken{Index: 0, Content: "some other ref"},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 9):   RefDefinitionToken{Index: 1, Content: "some other ref2"},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 10):  RefDefinitionToken{Index: 0, Content: "some ungrouped ref"},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_REF_DEF, 11):  RefDefinitionToken{Index: 1, Content: "some ungrouped ref2"},
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

func TestGetGroupOrDefaultAttribute(t *testing.T) {
	content := "some group=foo bar"
	group := getGroupOrDefault(content)
	test.AssertEqual(t, "foo", group)

	content = "some name=foo bar"
	group = getGroupOrDefault(content)
	test.AssertEqual(t, defaultReferenceGroup, group)
}
