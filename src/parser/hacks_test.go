package parser

import (
	"testing"
	"wiki2book/test"
)

func TestHackGermanRailwayTemplates_noTable(t *testing.T) {
	content := `foo
something
bar`
	expectedContent := `foo
something
bar`
	actualContent, err := hackGermanRailwayTemplates(content, 0)
	test.AssertNil(t, err)
	test.AssertEqual(t, expectedContent, actualContent)
}

func TestHackGermanRailwayTemplates_simple(t *testing.T) {
	content := `foo
{{BS-table}}
something
|}
bar`
	expectedContent := `foo

something

bar`
	actualContent, err := hackGermanRailwayTemplates(content, 0)
	test.AssertNil(t, err)
	test.AssertEqual(t, expectedContent, actualContent)
}

func TestHackGermanRailwayTemplates_specialCharacter(t *testing.T) {
	content := `föö
{{BS-table}}
sömethöng
|}
bär`
	expectedContent := `föö

sömethöng

bär`
	actualContent, err := hackGermanRailwayTemplates(content, 0)
	test.AssertNil(t, err)
	test.AssertEqual(t, expectedContent, actualContent)
}

func TestHackGermanRailwayTemplates_nested(t *testing.T) {
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
	actualContent, err := hackGermanRailwayTemplates(content, 0)
	test.AssertNil(t, err)
	test.AssertEqual(t, expectedContent, actualContent)
}
