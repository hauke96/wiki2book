package parser

import (
	"encoding/json"
	"fmt"
	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"testing"
	"wiki2book/api"
	"wiki2book/test"
)

var templateFolder = test.TestCacheFolder

func TestEvaluateTemplate_existingFile(t *testing.T) {
	tokenizer := NewTokenizer("foo", templateFolder)
	mockHttpClient := api.NewMockHttp("", 200)

	templateFile, err := os.Create(templateFolder + "/c740539f1a69d048c70ac185407dd5244b56632d")
	sigolo.FatalCheck(err)
	_, err = templateFile.WriteString("{\"expandtemplates\":{\"wikitext\":\"blubb\"}}")
	sigolo.FatalCheck(err)
	templateFile.Close()

	content, err := tokenizer.evaluateTemplates("Wikitext with {{my-template}}.")
	test.AssertNil(t, err)
	test.AssertEqual(t, 0, mockHttpClient.GetCalls)
	test.AssertEqual(t, 0, mockHttpClient.PostCalls)
	test.AssertEqual(t, "Wikitext with blubb.", content)
}

func TestEvaluateTemplate_newTemplate(t *testing.T) {
	tokenizer := NewTokenizer("foo", templateFolder)
	key := "7499ae1f1f8e45a9a95bdeb610ebf13cc4157667"
	expectedTemplateContent := "<div class=\"hauptartikel\" role=\"navigation\"><span class=\"hauptartikel-pfeil\" title=\"siehe\" aria-hidden=\"true\" role=\"presentation\">→ </span>''<span class=\"hauptartikel-text\">Hauptartikel</span>: [[Sternentstehung]]''</div>"
	jsonBytes, _ := json.Marshal(&api.WikiExpandedTemplateDto{ExpandTemplate: api.WikitextDto{Content: expectedTemplateContent}})
	expectedTemplateFileContent := string(jsonBytes)

	mockHttpClient := api.NewMockHttp(expectedTemplateFileContent, 200)

	// Evaluate content
	content, err := tokenizer.evaluateTemplates("Siehe {{Hauptartikel|Sternentstehung}}.")
	test.AssertNil(t, err)
	test.AssertEqual(t, 1, mockHttpClient.GetCalls)
	test.AssertEqual(t, 0, mockHttpClient.PostCalls)
	test.AssertEqual(t, "Siehe "+expectedTemplateContent+".", content)
	test.AssertTrue(t, hasLocalTemplate(key, templateFolder))

	// Read template content from disk
	expectedContent, err := getLocalTemplate(key, templateFolder)
	test.AssertNil(t, err)
	test.AssertEqual(t, expectedTemplateFileContent, expectedContent)
}

func TestEvaluateTemplate_nestedTemplates(t *testing.T) {
	// Simplified setup: There are two templates (outer and inner) evaluating to the same static string. This means in
	// the end, only one of these static strings remains, which is checked below.

	test.Prepare()

	tokenizer := NewTokenizer("foo", templateFolder)
	expectedTemplateContent := "<div>foo</div>"
	jsonBytes, _ := json.Marshal(&api.WikiExpandedTemplateDto{ExpandTemplate: api.WikitextDto{Content: expectedTemplateContent}})
	expectedTemplateFileContent := string(jsonBytes)

	mockHttpClient := api.NewMockHttp(expectedTemplateFileContent, 200)

	// Evaluate content
	content, err := tokenizer.evaluateTemplates("Siehe {{FOO|{{FOO}} bar}}")
	test.AssertNil(t, err)
	test.AssertEqual(t, 2, mockHttpClient.GetCalls)
	test.AssertEqual(t, 0, mockHttpClient.PostCalls)

	expectedContent := "Siehe " + expectedTemplateContent
	test.AssertNil(t, err)
	test.AssertEqual(t, expectedContent, content)
}

func TestEvaluateTemplate_nestedTemplatesWithTouchingEnds(t *testing.T) {
	test.Prepare()

	tokenizer := NewTokenizer("foo", templateFolder)
	expectedTemplateContent := "<div>foo</div>"
	jsonBytes, _ := json.Marshal(&api.WikiExpandedTemplateDto{ExpandTemplate: api.WikitextDto{Content: expectedTemplateContent}})
	expectedTemplateFileContent := string(jsonBytes)

	mockHttpClient := api.NewMockHttp(expectedTemplateFileContent, 200)

	// Evaluate content -> no space/separator between first }} and second }}
	content, err := tokenizer.evaluateTemplates("Siehe {{FOO|{{FOO}}}}")
	test.AssertNil(t, err)
	test.AssertEqual(t, 2, mockHttpClient.GetCalls)
	test.AssertEqual(t, 0, mockHttpClient.PostCalls)

	expectedContent := "Siehe " + expectedTemplateContent
	test.AssertNil(t, err)
	test.AssertEqual(t, expectedContent, content)
}

func hasLocalTemplate(key string, templateFolder string) bool {
	templateFilepath := filepath.Join(templateFolder, key)

	file, err := os.Open(templateFilepath)
	if file == nil || errors.Is(err, os.ErrNotExist) {
		return false
	}
	defer file.Close()

	return true
}

func getLocalTemplate(key string, templateFolder string) (string, error) {
	templateFilepath := filepath.Join(templateFolder, key)

	content, err := os.ReadFile(templateFilepath)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Error reading template %s from %s", key, templateFilepath))
	}

	return string(content), nil
}
