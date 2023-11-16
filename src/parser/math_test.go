package parser

import (
	"fmt"
	"testing"
	"wiki2book/test"
)

func TestParseMath(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `abc<math>x \cdot y</math>def
some
<math>
\multiline{math}
</math>
end`
	tokenizedContent := tokenizer.tokenizeContent(&tokenizer, content)

	expectedTokenKey0 := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_MATH, 0)
	expectedTokenKey1 := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_MATH, 1)
	test.AssertEqual(t, "abc"+expectedTokenKey0+"def\nsome\n"+expectedTokenKey1+"\nend", tokenizedContent)
	test.AssertMapEqual(t, map[string]interface{}{
		expectedTokenKey0: MathToken{Content: `x \cdot y`},
		expectedTokenKey1: MathToken{Content: "\n\\multiline{math}\n"},
	}, tokenizer.getTokenMap())
}
