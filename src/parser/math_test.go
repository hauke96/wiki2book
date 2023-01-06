package parser

import "testing"

func TestParseMath(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `abc<math>x \cdot y</math>def
some
<math>
\multiline{math}
</math>
end`
	tokenizedContent := tokenizer.tokenizeContent(&tokenizer, content)

	test.AssertEqual(t, "abc"+fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_MATH, 0)+"def\nsome\n"+fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_MATH, 1)+"\nend", tokenizedContent)
}
