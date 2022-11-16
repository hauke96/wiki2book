package parser

import (
	"encoding/json"
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/api"
	"github.com/hauke96/wiki2book/src/test"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const cacheSubFolder = "templates"

var templateFolder = test.GetCacheFolder(cacheSubFolder)

func TestMain(m *testing.M) {
	test.CleanRun(m, cacheSubFolder)
}

func TestEvaluateTemplate_existingFile(t *testing.T) {
	mockHttpClient := api.MockHttp("", 200)

	templateFile, err := os.Create(templateFolder + "/c740539f1a69d048c70ac185407dd5244b56632d")
	sigolo.FatalCheck(err)
	_, err = templateFile.WriteString("{\"expandtemplates\":{\"wikitext\":\"blubb\"}}")
	sigolo.FatalCheck(err)
	templateFile.Close()

	content := evaluateTemplates("Wikitext with {{my-template}}.", templateFolder)
	test.AssertEqual(t, 0, mockHttpClient.GetCalls)
	test.AssertEqual(t, 0, mockHttpClient.PostCalls)
	test.AssertEqual(t, "Wikitext with blubb.", content)
}

func TestEvaluateTemplate_newTemplate(t *testing.T) {
	key := "7499ae1f1f8e45a9a95bdeb610ebf13cc4157667"
	expectedTemplateContent := "<div class=\"hauptartikel\" role=\"navigation\"><span class=\"hauptartikel-pfeil\" title=\"siehe\" aria-hidden=\"true\" role=\"presentation\">→ </span>''<span class=\"hauptartikel-text\">Hauptartikel</span>: [[Sternentstehung]]''</div>"
	jsonBytes, _ := json.Marshal(&api.WikiExpandedTemplateDto{ExpandTemplate: api.WikitextDto{Content: expectedTemplateContent}})
	expectedTemplateFileContent := string(jsonBytes)

	mockHttpClient := api.MockHttp(expectedTemplateFileContent, 200)

	// Evaluate content
	content := evaluateTemplates("Siehe {{Hauptartikel|Sternentstehung}}.", templateFolder)
	test.AssertEqual(t, 1, mockHttpClient.GetCalls)
	test.AssertEqual(t, 0, mockHttpClient.PostCalls)
	test.AssertEqual(t, "Siehe "+expectedTemplateContent+".", content)
	test.AssertTrue(t, hasLocalTemplate(key, templateFolder))

	// Read template content from disk
	actualContent, err := getLocalTemplate(key, templateFolder)
	test.AssertNil(t, err)
	test.AssertEqual(t, expectedTemplateFileContent, actualContent)
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

	content, err := ioutil.ReadFile(templateFilepath)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Error reading template %s from %s", key, templateFilepath))
	}

	return string(content), nil
}
