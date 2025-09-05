package parser

import (
	"testing"
	"wiki2book/test"
	"wiki2book/wikipedia"
)

var templateFolder = test.TestCacheFolder

func TestEvaluateTemplate_existingFile(t *testing.T) {
	wikipediaService := wikipedia.DummyWikipediaService{EvaluateTemplateResponse: "blubb"}
	tokenizer := NewTokenizer(&wikipediaService)

	content, err := tokenizer.evaluateTemplates("Wikitext with {{my-template}}.")
	test.AssertNil(t, err)
	test.AssertEqual(t, "Wikitext with blubb.", content)
}

func TestEvaluateTemplate_newTemplate(t *testing.T) {
	expectedTemplateContent := "<div class=\"hauptartikel\" role=\"navigation\"><span class=\"hauptartikel-pfeil\" title=\"siehe\" aria-hidden=\"true\" role=\"presentation\">→ </span>''<span class=\"hauptartikel-text\">Hauptartikel</span>: [[Sternentstehung]]''</div>"
	wikipediaService := wikipedia.DummyWikipediaService{EvaluateTemplateResponse: expectedTemplateContent}
	tokenizer := NewTokenizer(&wikipediaService)

	// Evaluate content
	content, err := tokenizer.evaluateTemplates("Siehe {{Hauptartikel|Sternentstehung}}.")
	test.AssertNil(t, err)
	test.AssertEqual(t, "Siehe "+expectedTemplateContent+".", content)
}

func TestEvaluateTemplate_nestedTemplates(t *testing.T) {
	// Simplified setup: There are two templates (outer and inner) evaluating to the same static string. This means in
	// the end, only one of these static strings remains, which is checked below.

	test.Prepare()

	expectedTemplateContent := "<div>foo</div>"
	wikipediaService := wikipedia.DummyWikipediaService{EvaluateTemplateResponse: expectedTemplateContent}
	tokenizer := NewTokenizer(&wikipediaService)

	// Evaluate content
	content, err := tokenizer.evaluateTemplates("Siehe {{FOO|{{FOO}} bar}}")
	test.AssertNil(t, err)

	expectedContent := "Siehe " + expectedTemplateContent
	test.AssertNil(t, err)
	test.AssertEqual(t, expectedContent, content)
}

func TestEvaluateTemplate_nestedTemplatesWithTouchingEnds(t *testing.T) {
	test.Prepare()

	expectedTemplateContent := "<div>foo</div>"
	wikipediaService := wikipedia.DummyWikipediaService{EvaluateTemplateResponse: expectedTemplateContent}
	tokenizer := NewTokenizer(&wikipediaService)

	// Evaluate content -> no space/separator between first }} and second }}
	content, err := tokenizer.evaluateTemplates("Siehe {{FOO|{{FOO}}}}")
	test.AssertNil(t, err)

	expectedContent := "Siehe " + expectedTemplateContent
	test.AssertNil(t, err)
	test.AssertEqual(t, expectedContent, content)
}
