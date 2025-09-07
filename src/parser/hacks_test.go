package parser

import (
	"testing"
	"wiki2book/test"
)

func TestHackGermanRailwayTemplates_noTable(t *testing.T) {
	tokenizer := NewTokenizerWithMockWikipediaService()

	content := `foo
something
bar`
	expectedContent := `foo
something
bar`
	actualContent, err := tokenizer.hackGermanRailwayTemplates(content, 0)
	test.AssertNil(t, err)
	test.AssertEqual(t, expectedContent, actualContent)
}

func TestHackGermanRailwayTemplates_simple(t *testing.T) {
	tokenizer := NewTokenizerWithMockWikipediaService()

	content := `foo
{{BS-table}}
something
|}
bar`
	expectedContent := `foo

something

bar`
	actualContent, err := tokenizer.hackGermanRailwayTemplates(content, 0)
	test.AssertNil(t, err)
	test.AssertEqual(t, expectedContent, actualContent)
}

func TestHackGermanRailwayTemplates_specialCharacter(t *testing.T) {
	tokenizer := NewTokenizerWithMockWikipediaService()

	content := `föö
{{BS-table}}
sömethöng
|}
bär`
	expectedContent := `föö

sömethöng

bär`
	actualContent, err := tokenizer.hackGermanRailwayTemplates(content, 0)
	test.AssertNil(t, err)
	test.AssertEqual(t, expectedContent, actualContent)
}

func TestHackGermanRailwayTemplates_nested(t *testing.T) {
	tokenizer := NewTokenizerWithMockWikipediaService()

	content := `foo
{{BS-table}}
something
{{BS-table}}
some inner stuff
|}
some outer stuff
|}
bar`
	expectedContent := `foo

something

some inner stuff

some outer stuff

bar`
	actualContent, err := tokenizer.hackGermanRailwayTemplates(content, 0)
	test.AssertNil(t, err)
	test.AssertEqual(t, expectedContent, actualContent)
}
