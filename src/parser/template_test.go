package parser

import (
	"github.com/hauke96/wiki2book/src/test"
	"os"
	"testing"
)

const templateFolder = "../test/templates"

// Cleanup from previous runs
func cleanup(t *testing.T, key string) {
	err := os.Remove(templateFolder + "/" + key)
	test.AssertTrue(t, err == nil || os.IsNotExist(err))
}

func TestHasLocalTemplate(t *testing.T) {
	hasTemplate := hasLocalTemplate("template1", templateFolder)
	test.AssertEqual(t, true, hasTemplate)
}

func TestHasLocalTemplate_notExisting(t *testing.T) {
	hasTemplate := hasLocalTemplate("template-not-existing", templateFolder)
	test.AssertEqual(t, false, hasTemplate)
}

func TestGetTemplate(t *testing.T) {
	content, err := getTemplate("template1", templateFolder)
	test.AssertNil(t, err)
	test.AssertEqual(t, "foobar", content)
}

func TestGetTemplate_notExisting(t *testing.T) {
	content, err := getTemplate("template-not-existing", templateFolder)
	test.AssertEqual(t, "Error reading template template-not-existing from ../test/templates/template-not-existing: open ../test/templates/template-not-existing: no such file or directory", err.Error())
	test.AssertEqual(t, "", content)
}

func TestEvaluateTemplate_existingFile(t *testing.T) {
	content := evaluateTemplates("Wikitext with {{my-template}}.", templateFolder)
	test.AssertEqual(t, "Wikitext with blubb.", content)
}

func TestEvaluateTemplate_newTemplate(t *testing.T) {
	key := "7499ae1f1f8e45a9a95bdeb610ebf13cc4157667"
	expectedTemplateContent := "<div class=\"hauptartikel\" role=\"navigation\"><span class=\"hauptartikel-pfeil\" title=\"siehe\" aria-hidden=\"true\" role=\"presentation\">→ </span>''<span class=\"hauptartikel-text\">Hauptartikel</span>: [[Sternentstehung]]''</div>"

	cleanup(t, key)

	// Cleanup from previous runs
	err := os.Remove(templateFolder + "/" + key)
	test.AssertTrue(t, err == nil || os.IsNotExist(err))

	// Actually evaluate content
	content := evaluateTemplates("Siehe {{Hauptartikel|Sternentstehung}}.", templateFolder)
	test.AssertEqual(t, "Siehe "+expectedTemplateContent+".", content)

	// Check if file exists
	hasTemplate := hasLocalTemplate(key, templateFolder)
	test.AssertEqual(t, true, hasTemplate)

	actualContent, err := getTemplate(key, templateFolder)
	test.AssertNil(t, err)
	test.AssertEqual(t, expectedTemplateContent, actualContent)
}
