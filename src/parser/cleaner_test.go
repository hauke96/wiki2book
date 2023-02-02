package parser

import (
	"github.com/hauke96/wiki2book/src/test"
	"testing"
)

func TestRemoveComments(t *testing.T) {
	content := "foo bar\nblubb hi"
	content = removeComments(content)
	test.AssertEqual(t, "foo bar\nblubb hi", content)

	content = "foo"
	content = removeComments(content)
	test.AssertEqual(t, "foo", content)

	content = ""
	content = removeComments(content)
	test.AssertEqual(t, "", content)

	content = "foo <!-- bar --> blubb"
	content = removeComments(content)
	test.AssertEqual(t, "foo  blubb", content)

	content = "foo <!-- bar -->"
	content = removeComments(content)
	test.AssertEqual(t, "foo ", content)

	content = "<!-- bar --> foo"
	content = removeComments(content)
	test.AssertEqual(t, " foo", content)

	content = "<!-- bar -->"
	content = removeComments(content)
	test.AssertEqual(t, "", content)

	content = "<!---->"
	content = removeComments(content)
	test.AssertEqual(t, "", content)

	content = "<!-- foo"
	content = removeComments(content)
	test.AssertEqual(t, "<!-- foo", content)

	content = "foo -->"
	content = removeComments(content)
	test.AssertEqual(t, "foo -->", content)
}

func TestRemoveUnwantedCategories(t *testing.T) {
	content := "[[Kategorie:foo]][[Category:FOO:BAR\n$ome+µeird-string]]"
	content = removeUnwantedCategories(content)
	test.AssertEmptyString(t, content)
}

func TestRemoveUnwantedTemplates(t *testing.T) {
	content := `{{siehe auch}}{{GRAPH:CHART
|$ome+µeird-string}}{{let this template stay}}{{
toc }}`
	content = removeUnwantedTemplates(content)
	test.AssertEqual(t, "{{let this template stay}}", content)
}

func TestRemoveUnwantedMultiLineTemplates(t *testing.T) {
	content := "foo\n{{NaviBlock\n|Navigationsleiste Monde\n|Navigationsleiste_Sonnensystem}}\nbar"
	content = removeUnwantedTemplates(content)
	test.AssertEqual(t, "foo\n\nbar", content)
}

func TestRemoveUnwantedHtml(t *testing.T) {
	content := "Some <div>noice</div><div style=\"height: 123px;\"> HTML</div>"
	content = removeUnwantedHtml(content)
	test.AssertEqual(t, "Some noice HTML", content)
}

func TestClean(t *testing.T) {
	content := "<div foo>Some</div> [[Category:weird]]wikitext{{Wikisource}}"
	content = clean(content)
	test.AssertEqual(t, "Some wikitext", content)

	content = " '''test'''"
	content = clean(content)
	test.AssertEqual(t, content, content)

	content = `
== Einzelnachweise ==
<references />

{{Gesprochene Version
|artikel    = Erde
|datei      = De-Erde-article.ogg
|länge      = 41:43 min
|größe      = 20,4 MB
|sprecher   = Ahoek
|geschlecht = Männlich
|dialekt    = Hochdeutsch
|version    = 121631898
|datum      = 2013-08-25}}

{{NaviBlock
|Navigationsleiste Sonnensystem
|Navigationsleiste Monde
}}
foo`
	content = clean(content)
	test.AssertEqual(t, `
== Einzelnachweise ==
<references />




foo`, content)
}

func TestGetTrimmedLine(t *testing.T) {
	lines := make([]string, 10)
	lines[0] = "abc"
	lines[1] = " abc"
	lines[2] = "abc "
	lines[3] = "	abc "
	lines[4] = "	abc\n "
	lines[5] = " "
	lines[6] = "	"
	lines[7] = "\n"

	test.AssertEqual(t, "abc", getTrimmedLine(lines, 0))
	test.AssertEqual(t, "abc", getTrimmedLine(lines, 1))
	test.AssertEqual(t, "abc", getTrimmedLine(lines, 2))
	test.AssertEqual(t, "abc", getTrimmedLine(lines, 3))
	test.AssertEqual(t, "abc", getTrimmedLine(lines, 4))
	test.AssertEqual(t, "", getTrimmedLine(lines, 5))
	test.AssertEqual(t, "", getTrimmedLine(lines, 6))
	test.AssertEqual(t, "", getTrimmedLine(lines, 7))
}

func TestIsHeading(t *testing.T) {
	test.AssertEqual(t, 1, headingDepth("= abc ="))
	test.AssertEqual(t, 2, headingDepth("== abc =="))
	test.AssertEqual(t, 3, headingDepth("=== abc ==="))
	test.AssertEqual(t, 4, headingDepth("==== abc ===="))
	test.AssertEqual(t, 5, headingDepth("===== abc ====="))
	test.AssertEqual(t, 6, headingDepth("====== abc ======"))
	test.AssertEqual(t, 7, headingDepth("======= abc ======="))

	test.AssertEqual(t, 0, headingDepth("== abc "))
	test.AssertEqual(t, 0, headingDepth("=== abc =="))
	test.AssertEqual(t, 0, headingDepth("== abc ==="))
	test.AssertEqual(t, 0, headingDepth("abc =="))
	test.AssertEqual(t, 0, headingDepth("abc"))
	test.AssertEqual(t, 0, headingDepth(""))
	test.AssertEqual(t, 0, headingDepth(" '''test'''"))
}

func TestRemoveEmptyListEntries(t *testing.T) {
	content := `foo
*
* 
* bar
*  
blubb`
	test.AssertEqual(t, `foo
* bar
blubb`, removeEmptyListEntries(content))
}

func TestRemoveEmptySection_normal(t *testing.T) {
	content := `foo

== heading ==
foo

bar
`

	test.AssertEqual(t, content, removeEmptySections(content))
}

func TestRemoveEmptySection_withEmptySections(t *testing.T) {
	content := `foo

== heading ==
should remain

== should be removed==


== should be removed as well ==`
	expectedResult := `foo

== heading ==
should remain
`

	test.AssertEqual(t, expectedResult, removeEmptySections(content))
}

func TestRemoveEmptySection_noSection(t *testing.T) {
	content := " '''test'''"
	test.AssertEqual(t, content, removeEmptySections(content))
}

func TestRemoveEmptySection_linesWithSpaces(t *testing.T) {
	content := `foo
== heading==


== heading ==
 
	
				
`
	expectedResult := "foo"

	test.AssertEqual(t, expectedResult, removeEmptySections(content))
}

func TestRemoveEmptySection_superSectionNotRemoved(t *testing.T) {
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

	test.AssertEqual(t, expected, removeEmptySections(content))
}

func TestRemoveEmptySection_withSemiSection(t *testing.T) {
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

	test.AssertEqual(t, expectedResult, removeEmptySections(content))
}
