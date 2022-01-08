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
	test.AssertEqualString(t, "{{let this template stay}}", content)
}

func TestRemoveUnwantedHtml(t *testing.T) {
	content := "Some <div>noice</div><div style=\"height: 123px;\"> HTML</div>"
	content = removeUnwantedHtml(content)
	test.AssertEqualString(t, "Some noice HTML", content)
}

func TestClean(t *testing.T) {
	content := "<div foo>Some</div> [[Category:weird]]wikitext{{Wikisource}}"
	content = clean(content)
	test.AssertEqualString(t, "Some wikitext", content)
}
