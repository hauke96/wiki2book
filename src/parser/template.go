package parser

import (
	"errors"
	"fmt"
	"github.com/hauke96/sigolo"
	"strings"
	"wiki2book/api"
	"wiki2book/util"
)

const templatePlaceholderTemplate = "$$TEMPLATE_PLACEHOLDER_%s$$"

// evaluateTemplates evaluates all templates including nested ones.
func (t *Tokenizer) evaluateTemplates(content string) (string, error) {
	// All evaluated templates are stored in this map. Replacing evaluated templates by placeholders reduces the length
	// of request URLs significantly and prevents errors due to too long URLs.
	placeholderToContent := map[string]string{}
	startToken := "{{"
	endToken := "}}"

	sigolo.Debug("Start evaluating templates and replacing them by placeholders")
	for i := 0; i < len(content)-2; i++ {
		cursor := content[i : i+2]

		if cursor == startToken {
			endIndex := findCorrespondingCloseToken(content, i+2, startToken, endToken)
			if endIndex == -1 {
				return "", errors.New(fmt.Sprintf("Found %s but no corresponding %s. I'll ignore this but something's wrong with the input wikitext!", startToken, endToken))
			}

			templateText := content[i : endIndex+2]

			if strings.Contains(templateText[2:], startToken) {
				// If the template itself contains a template, then proceed to first evaluate the inner template and
				// to evaluate the outer template in a later run
				continue
			}

			key := util.Hash(templateText)

			evaluatedTemplate, err := api.EvaluateTemplate(templateText, t.templateFolder, key)
			if err != nil {
				return "", err
			}

			// Replace the template by a placeholder. We do not directly replace the wikitext of the template with the
			// evaluated form because nested templates might lead to too long URLs.
			placeholderToContent[key] = evaluatedTemplate
			placeholder := fmt.Sprintf(templatePlaceholderTemplate, key)
			content = strings.Replace(content, templateText, placeholder, 1)
		}
	}
	sigolo.Debug("Finished finding and evaluating templates")

	// Replace all template placeholders with the actual content until no placeholders are unresolved. This is not very
	// elegant or fast but due to the nesting a simple and working approach.
	sigolo.Debug("Replace %d template placeholder with evaluated content", len(placeholderToContent))
	for len(placeholderToContent) != 0 {
		for key, template := range placeholderToContent {
			placeholder := fmt.Sprintf(templatePlaceholderTemplate, key)
			if strings.Contains(content, placeholder) {
				content = strings.ReplaceAll(content, placeholder, template)
				delete(placeholderToContent, key)
			}
		}
	}
	sigolo.Debug("Finished replacing template placeholders. Template handling done.")

	return content, nil
}
