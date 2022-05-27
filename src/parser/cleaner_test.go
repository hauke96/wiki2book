package parser

import (
	"github.com/hauke96/wiki2book/src/test"
	"testing"
)

func TestRemoveUnwantedCategories(t *testing.T) {
	content := "[[Kategorie:foo]][[Category:FOO:BAR\n$ome+µeird-string]]"
	content = removeUnwantedCategories(content)
	test.AssertEmptyString(t, content)
}

func TestRemoveUnwantedTemplates(t *testing.T) {
	content := "{{siehe auch}}{{GRAPH:CHART\n$ome+µeird-string}}{{let this template stay}}"
	content = removeUnwantedTemplates(content)
	test.AssertEqual(t, "{{let this template stay}}", content)
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
	test.AssertTrue(t, isHeading("= abc ="))
	test.AssertTrue(t, isHeading("== abc =="))
	test.AssertTrue(t, isHeading("=== abc ==="))
	test.AssertTrue(t, isHeading("==== abc ===="))
	test.AssertTrue(t, isHeading("===== abc ====="))
	test.AssertTrue(t, isHeading("====== abc ======"))
	test.AssertTrue(t, isHeading("======= abc ======="))

	test.AssertFalse(t, isHeading("== abc "))
	test.AssertFalse(t, isHeading("abc =="))
	test.AssertFalse(t, isHeading("abc"))
	test.AssertFalse(t, isHeading(""))
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
foo

bar

== heading==


== heading ==`
	expectedResult := `foo

== heading ==
foo

bar
`

	test.AssertEqual(t, expectedResult, removeEmptySections(content))
}

func TestRemoveEmptySection_linesWithSpaces(t *testing.T) {
	content := `foo
== heading==


== heading ==
 
	
				
`
	expectedResult := "foo"

	test.AssertEqual(t, expectedResult, removeEmptySections(content))
}

// TODO
func TestRemoveEmptySection_superSectionNotRemoved(t *testing.T) {
	content := `foo
== heading ==

=== sub heading ===
 
==== sub sub heading ===
	
				
`

	test.AssertEqual(t, content, removeEmptySections(content))
}
