package parser

import (
	"testing"
	"wiki2book/config"
	"wiki2book/test"
	"wiki2book/wikipedia"
)

func TestRemoveComments(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})

	content := "foo bar\nblubb hi"
	content = tokenizer.removeComments(content)
	test.AssertEqual(t, "foo bar\nblubb hi", content)

	content = "foo"
	content = tokenizer.removeComments(content)
	test.AssertEqual(t, "foo", content)

	content = ""
	content = tokenizer.removeComments(content)
	test.AssertEqual(t, "", content)

	content = "foo <!-- bar --> blubb"
	content = tokenizer.removeComments(content)
	test.AssertEqual(t, "foo  blubb", content)

	content = "foo <!-- <!-- bar --> --> blubb"
	content = tokenizer.removeComments(content)
	test.AssertEqual(t, "foo  --> blubb", content)

	content = "foo <!-- bar -->"
	content = tokenizer.removeComments(content)
	test.AssertEqual(t, "foo ", content)

	content = "<!-- bar --> foo"
	content = tokenizer.removeComments(content)
	test.AssertEqual(t, " foo", content)

	content = "<!-- bar -->"
	content = tokenizer.removeComments(content)
	test.AssertEqual(t, "", content)

	content = "<!---->"
	content = tokenizer.removeComments(content)
	test.AssertEqual(t, "", content)

	content = "<!-- foo"
	content = tokenizer.removeComments(content)
	test.AssertEqual(t, "<!-- foo", content)

	content = "foo -->"
	content = tokenizer.removeComments(content)
	test.AssertEqual(t, "foo -->", content)

	content = `* foo
<!--
-->
* bar
`
	content = tokenizer.removeComments(content)
	test.AssertEqual(t, `* foo
* bar
`, content)

	content = `* list<!--comment-->

bar
`
	content = tokenizer.removeComments(content)
	test.AssertEqual(t, `* list

bar
`, content)
}

func TestRemoveUnwantedLinks_unwantedCategories(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})

	content := "[[Category:foo]][[Category:FOO:BAR\n$ome+µeird-string]]"
	content = tokenizer.removeUnwantedInternalLinks(content)
	test.AssertEmptyString(t, content)
}

func TestRemoveUnwantedLinks_unwantedCategoriesStayWhenNormalLink(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})

	content := "[[Category:foo]][[:Category:FOO:This will stay]]"
	content = tokenizer.removeUnwantedInternalLinks(content)
	test.AssertEqual(t, "[[:Category:FOO:This will stay]]", content)
}

func TestRemoveUnwantedLinks(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})

	config.Current.AllowedLinkPrefixes = []string{"arxiv"}

	content := `[[de:foo]][[arxiv:whatever]][[DE:FOO]][[EN:FOO:BAR
$ome+µeird-string]]
[[should_stay]]
before[[de:foo:bar]]after
before[[internal]]after
before image [[iMAge:this-should:stay.jpg]] after image`
	expected := `[[arxiv:whatever]]
[[should_stay]]
beforeafter
before[[internal]]after
before image [[iMAge:this-should:stay.jpg]] after image`
	actual := tokenizer.removeUnwantedInternalLinks(content)
	test.AssertEqual(t, expected, actual)

	content = "foo[[de:pic.jpg|mini|With [[nested]]]]bar"
	expected = "foobar"
	actual = tokenizer.removeUnwantedInternalLinks(content)
	test.AssertEqual(t, expected, actual)
}

func TestRemoveUnwantedLinks_nestedLinks(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})

	content := `foo[[file:pic.jpg|mini|Nested [[link|l]]-thingy]]bar`
	cleanedContent := tokenizer.removeUnwantedInternalLinks(content)
	test.AssertEqual(t, content, cleanedContent)

	content = "foo[[de:pic.jpg|mini|With [[nested]] link]]bar"
	expected := "foobar"
	actual := tokenizer.removeUnwantedInternalLinks(content)
	test.AssertEqual(t, expected, actual)
}

func TestRemoveUnwantedTemplates(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})

	config.Current.IgnoredTemplates = []string{"graph:chart", "siehe auch", "toc"}

	content := `{{siehe auch}}{{GRAPH:CHART
|$ome+µeird-string}}{{let this template stay}}{{
toc }}`
	content = tokenizer.handleUnwantedAndTrailingTemplates(content)
	test.AssertEqual(t, "{{let this template stay}}", content)
}

func TestRemoveUnwantedMultiLineTemplates(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})

	config.Current.IgnoredTemplates = []string{"naviblock"}

	content := `foo
{{NaviBlock
|Navigationsleiste Monde
|Navigationsleiste_Sonnensystem}}
bar`
	content = tokenizer.handleUnwantedAndTrailingTemplates(content)
	test.AssertEqual(t, "foo\n\nbar", content)
}

func TestRemoveUnwantedHtml(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})

	content := "Some <div>noice</div><div style=\"height: 123px;\"> HTML</div>"
	content = tokenizer.removeUnwantedHtml(content)
	test.AssertEqual(t, "Some noice HTML", content)
}

func TestMoveTrailingTemplatesDown(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})

	config.Current.TrailingTemplates = []string{"FOO", "bar"}

	content := `{{siehe auch}}{{foo}}{{foo}}{{let this template stay}}{{bar}}`
	content = tokenizer.handleUnwantedAndTrailingTemplates(content)
	test.AssertEqual(t, "{{siehe auch}}{{let this template stay}}\n{{foo}}\n{{foo}}\n{{bar}}", content)
}

func TestClean(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})

	config.Current.IgnoredTemplates = []string{"wikisource", "gesprochene version", "naviblock", "positionskarte+", "positionskarte~", "hauptartikel"}
	var err error

	content := "<div foo>Some</div> [[Category:weird]]wikitext{{Wikisource}}"
	content, err = tokenizer.clean(content)
	test.AssertNil(t, err)
	test.AssertEqual(t, "<div foo>Some</div> wikitext", content)

	content = " '''test'''"
	content, err = tokenizer.clean(content)
	test.AssertNil(t, err)
	test.AssertEqual(t, content, content)

	content = `
== Einzelnachweise ==
<references />

{{Gesprochene Version
|artikel    = Erde
|datum      = 2013-08-25}}

{{NaviBlock
|Navigationsleiste Sonnensystem
|Navigationsleiste Monde
}}
foo`
	content, err = tokenizer.clean(content)
	test.AssertNil(t, err)
	test.AssertEqual(t, `
== Einzelnachweise ==
<references />




foo`, content)
	content = `
== Foo ==
{{Hauptartikel}}<!-- This comment should not cause problems -->
{{Positionskarte+ |Title | foo {{should_be_removed_as_well}} |places=
{{Positionskarte~ |foobar}}
}}

bar`
	content, err = tokenizer.clean(content)
	test.AssertNil(t, err)
	test.AssertEqual(t, `
== Foo ==



bar`, content)
}

func TestIsHeading(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})

	test.AssertEqual(t, 1, tokenizer.headingDepth("= abc ="))
	test.AssertEqual(t, 2, tokenizer.headingDepth("== abc =="))
	test.AssertEqual(t, 3, tokenizer.headingDepth("=== abc ==="))
	test.AssertEqual(t, 4, tokenizer.headingDepth("==== abc ===="))
	test.AssertEqual(t, 5, tokenizer.headingDepth("===== abc ====="))
	test.AssertEqual(t, 6, tokenizer.headingDepth("====== abc ======"))
	test.AssertEqual(t, 7, tokenizer.headingDepth("======= abc ======="))

	test.AssertEqual(t, 3, tokenizer.headingDepth("=== abc==="))
	test.AssertEqual(t, 3, tokenizer.headingDepth("===abc ==="))
	test.AssertEqual(t, 3, tokenizer.headingDepth("=== äöü ==="))
	test.AssertEqual(t, 3, tokenizer.headingDepth("=== äöß ==="))
	test.AssertEqual(t, 3, tokenizer.headingDepth("=== ä→ß ==="))
	test.AssertEqual(t, semiHeadingDepth, tokenizer.headingDepth("'''test'''"))

	test.AssertEqual(t, 0, tokenizer.headingDepth("== abc "))
	test.AssertEqual(t, 0, tokenizer.headingDepth("abc =="))
	test.AssertEqual(t, 0, tokenizer.headingDepth("=== abc =="))
	test.AssertEqual(t, 0, tokenizer.headingDepth("== abc ==="))
	test.AssertEqual(t, 0, tokenizer.headingDepth("abc"))
	test.AssertEqual(t, 0, tokenizer.headingDepth(""))
	test.AssertEqual(t, 0, tokenizer.headingDepth(" == abc =="))
	test.AssertEqual(t, 0, tokenizer.headingDepth("== abc == "))
	test.AssertEqual(t, 0, tokenizer.headingDepth(" '''test'''"))
	test.AssertEqual(t, 0, tokenizer.headingDepth("'''test''' "))
}

func TestRemoveEmptyListEntries(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})

	content := `foo
*
* 
* bar
*  
blubb`
	test.AssertEqual(t, `foo
* bar
blubb`, tokenizer.removeEmptyListEntries(content))
}

func TestRemoveEmptyListEntries_nestedLists(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})

	content := `foo
*
*#
*# bar
*#  
blubb`
	test.AssertEqual(t, `foo
*# bar
blubb`, tokenizer.removeEmptyListEntries(content))
}

func TestRemoveEmptySection_normal(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})

	content := `foo

== heading ==
foo

bar
`

	test.AssertEqual(t, content, tokenizer.removeEmptySections(content))
}

func TestRemoveEmptySection_withEmptySections(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})

	content := `foo

== heading ==
should remain

== should be removed==


== should be removed as well ==`
	expectedResult := `foo

== heading ==
should remain
`

	test.AssertEqual(t, expectedResult, tokenizer.removeEmptySections(content))
}

func TestRemoveEmptySection_noSection(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})

	content := " '''test'''"
	test.AssertEqual(t, content, tokenizer.removeEmptySections(content))
}

func TestRemoveEmptySection_linesWithSpaces(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})

	content := `foo
== heading==


== heading ==
 
	
				
`
	expectedResult := "foo"

	test.AssertEqual(t, expectedResult, tokenizer.removeEmptySections(content))
}

func TestRemoveEmptySection_superSectionNotRemoved(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})

	content := `foo
== heading ==

=== sub heading to be removed ===

=== sub heading ===
bar

=== sub heading ===
 
==== sub sub heading ===
blubb

= heading to be removed =
				
`
	expected := `foo
== heading ==

=== sub heading ===
bar

=== sub heading ===
 
==== sub sub heading ===
blubb
`

	test.AssertEqual(t, expected, tokenizer.removeEmptySections(content))
}

func TestRemoveEmptySection_withSemiSection(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})

	content := `foo

= heading =
foo

''' semi heading ''''

== some other heading ==

== some other non empty heading ==
bar
`
	expectedResult := `foo

= heading =
foo

== some other non empty heading ==
bar
`

	test.AssertEqual(t, expectedResult, tokenizer.removeEmptySections(content))
}

func TestRemoveEmptySection_pureBoldTextShouldNotBeChanged(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})

	content := `'''foo'''`
	expectedResult := `'''foo'''`
	test.AssertEqual(t, expectedResult, tokenizer.removeEmptySections(content))
}
