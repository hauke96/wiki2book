package parser

import (
	"strings"
	"wiki2book/util"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

var (
	templateStartToken    = "{{"
	templateEndToken      = "}}"
	templateStartTokenLen = len(templateStartToken)
	templateEndTokenLen   = len(templateEndToken)
)

// evaluateTemplates evaluates all templates and replaces their occurrence with the expanded (i.e. evaluated) template
// from the Wikipedia API. After calling this, there are no templates left.
func (t *Tokenizer) evaluateTemplates(content string) (string, error) {
	sigolo.Debug("Expand templates and replace them by placeholders")

	for i := 0; i < len(content)-templateEndTokenLen; i++ {
		cursor := content[i : i+templateStartTokenLen]

		if cursor == templateStartToken {
			endIndex := FindCorrespondingCloseToken(content, i+templateStartTokenLen, templateStartToken, templateEndToken)
			if endIndex == -1 {
				return "", errors.Errorf("Found %s but no corresponding %s. I'll ignore this but something's wrong with the input wikitext!", templateStartToken, templateEndToken)
			}

			templateText := content[i : endIndex+templateEndTokenLen]
			key := util.Hash(templateText)

			sigolo.Tracef("Evaluate template (key=%s): %s", key, util.TruncString(templateText))
			evaluatedTemplate, err := t.wikipediaService.EvaluateTemplate(templateText, key)
			if err != nil {
				return "", err
			}

			content = strings.Replace(content, templateText, evaluatedTemplate, 1)
		}
	}

	sigolo.Debug("Finished expanding templates")
	return content, nil
}
