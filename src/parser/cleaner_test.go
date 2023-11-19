package parser

import (
	"testing"
	"wiki2book/config"
	"wiki2book/test"
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

	content = "foo <!-- <!-- bar --> --> blubb"
	content = removeComments(content)
	test.AssertEqual(t, "foo  --> blubb", content)

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

	content = `* foo
<!--
-->
* bar
`
	content = removeComments(content)
	test.AssertEqual(t, `* foo
* bar
`, content)

	content = `* list<!--comment-->

bar
`
	content = removeComments(content)
	test.AssertEqual(t, `* list

bar
`, content)
}

func TestRemoveUnwantedCategories(t *testing.T) {
	content := "[[Category:foo]][[Category:FOO:BAR\n$ome+µeird-string]]"
	content = removeUnwantedInternalLinks(content)
	test.AssertEmptyString(t, content)
}

func TestRemoveUnwantedLinks(t *testing.T) {
	config.Current.AllowedLinkPrefixes = []string{"arxiv"}

	content := `[[:de:foo]][[arxiv:whatever]][[:DE:FOO]][[:EN:FOO:BAR
$ome+µeird-string]]
[[:should_stay]]
before[[:de:foo:bar]]after
before[[internal]]after
before image [[iMAge:this-should:stay.jpg]] after image`
	expected := `[[arxiv:whatever]]
[[:should_stay]]
beforeafter
before[[internal]]after
before image [[iMAge:this-should:stay.jpg]] after image`
	actual := removeUnwantedInternalLinks(content)
	test.AssertEqual(t, expected, actual)

	content = "foo[[:de:pic.jpg|mini|With [[nested]]]]bar"
	expected = "foobar"
	actual = removeUnwantedInternalLinks(content)
	test.AssertEqual(t, expected, actual)
}

func TestRemoveUnwantedLinks_nestedLinks(t *testing.T) {
	content := `foo[[file:pic.jpg|mini|Nested [[link|l]]-thingy]]bar`
	cleanedContent := removeUnwantedInternalLinks(content)
	test.AssertEqual(t, content, cleanedContent)

	content = "foo[[:de:pic.jpg|mini|With [[nested]] link]]bar"
	expected := "foobar"
	actual := removeUnwantedInternalLinks(content)
	test.AssertEqual(t, expected, actual)
}

func TestRemoveUnwantedTemplates(t *testing.T) {
	config.Current.IgnoredTemplates = []string{"graph:chart", "siehe auch", "toc"}

	content := `{{siehe auch}}{{GRAPH:CHART
|$ome+µeird-string}}{{let this template stay}}{{
toc }}`
	content = removeUnwantedTemplates(content)
	test.AssertEqual(t, "{{let this template stay}}", content)
}

func TestRemoveUnwantedMultiLineTemplates(t *testing.T) {
	config.Current.IgnoredTemplates = []string{"naviblock"}

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
	config.Current.IgnoredTemplates = []string{"wikisource", "gesprochene version", "naviblock", "positionskarte+", "positionskarte~", "hauptartikel"}
	var err error

	content := "<div foo>Some</div> [[Category:weird]]wikitext{{Wikisource}}"
	content, err = clean(content)
	test.AssertNil(t, err)
	test.AssertEqual(t, "Some wikitext", content)

	content = " '''test'''"
	content, err = clean(content)
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
	content, err = clean(content)
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
	content, err = clean(content)
	test.AssertNil(t, err)
	test.AssertEqual(t, `
== Foo ==



bar`, content)
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

	test.AssertEqual(t, 3, headingDepth("=== abc==="))
	test.AssertEqual(t, 3, headingDepth("===abc ==="))
	test.AssertEqual(t, 3, headingDepth("=== äöü ==="))
	test.AssertEqual(t, 3, headingDepth("=== äöß ==="))
	test.AssertEqual(t, 3, headingDepth("=== ä→ß ==="))
	test.AssertEqual(t, semiHeadingDepth, headingDepth("'''test'''"))

	test.AssertEqual(t, 0, headingDepth("== abc "))
	test.AssertEqual(t, 0, headingDepth("abc =="))
	test.AssertEqual(t, 0, headingDepth("=== abc =="))
	test.AssertEqual(t, 0, headingDepth("== abc ==="))
	test.AssertEqual(t, 0, headingDepth("abc"))
	test.AssertEqual(t, 0, headingDepth(""))
	test.AssertEqual(t, 0, headingDepth(" == abc =="))
	test.AssertEqual(t, 0, headingDepth("== abc == "))
	test.AssertEqual(t, 0, headingDepth(" '''test'''"))
	test.AssertEqual(t, 0, headingDepth("'''test''' "))
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

func TestRemoveEmptyListEntries_nestedLists(t *testing.T) {
	content := `foo
*
*#
*# bar
*#  
blubb`
	test.AssertEqual(t, `foo
*# bar
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

func TestRemoveEmptySection_pureBoldTextShouldNotBeChanged(t *testing.T) {
	content := `'''foo'''`
	expectedResult := `'''foo'''`
	test.AssertEqual(t, expectedResult, removeEmptySections(content))
}
