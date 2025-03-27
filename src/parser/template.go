package parser

import (
	"fmt"
	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
	"strings"
	"wiki2book/util"
	"wiki2book/wikipedia"
)

const templatePlaceholderPrefix = "$$TEMPLATE_PLACEHOLDER_"
const templatePlaceholderTemplate = templatePlaceholderPrefix + "%s$$"

var (
	templateStartToken    = "{{"
	templateEndToken      = "}}"
	templateStartTokenLen = len(templateStartToken)
	templateEndTokenLen   = len(templateEndToken)
)

// evaluateTemplates evaluates all templates including nested ones.
func (t *Tokenizer) evaluateTemplates(content string) (string, error) {
	// All evaluated templates are stored in this map. Replacing evaluated templates by placeholders reduces the length
	// of request URLs significantly and prevents errors due to too long URLs.
	placeholderToContent := map[string]string{}

	sigolo.Debug("Start evaluating templates and replacing them by placeholders")
	content, err := t.replaceTemplateByPlaceholders(content, placeholderToContent)
	if err != nil {
		return "", err
	}
	sigolo.Debug("Finished finding and evaluating templates")

	// Replace all template placeholders with the actual content until no placeholders are unresolved. This is not very
	// elegant or fast but due to the nesting a simple and working approach.
	sigolo.Debugf("Replace %d template placeholder with evaluated content", len(placeholderToContent))
	for strings.Contains(content, templatePlaceholderPrefix) {
		sigolo.Tracef("Check content for template placeholders (%d remain)", len(placeholderToContent))

		for key, template := range placeholderToContent {
			placeholder := fmt.Sprintf(templatePlaceholderTemplate, key)
			containsPlaceholder := strings.Contains(content, placeholder)
			sigolo.Tracef("Check template placeholder %s -> content contains placeholder? %v", key, containsPlaceholder)

			if containsPlaceholder {
				content = strings.ReplaceAll(content, placeholder, template)
				sigolo.Tracef("Replaced template placeholder %s in content", key)
			}
		}
	}
	sigolo.Debug("Finished replacing template placeholders. Template handling done.")

	return content, nil
}

func (t *Tokenizer) replaceTemplateByPlaceholders(content string, placeholderToContent map[string]string) (string, error) {
	sigolo.Tracef("Replace template tokens in content '%s'", util.TruncString(content))
	for i := 0; i < len(content)-templateEndTokenLen; i++ {
		cursor := content[i : i+templateStartTokenLen]

		if cursor == templateStartToken {
			endIndex := findCorrespondingCloseToken(content, i+templateStartTokenLen, templateStartToken, templateEndToken)
			if endIndex == -1 {
				return "", errors.Errorf("Found %s but no corresponding %s. I'll ignore this but something's wrong with the input wikitext!", templateStartToken, templateEndToken)
			}

			originalTemplateText := content[i : endIndex+templateEndTokenLen]
			templateText := originalTemplateText
			sigolo.Tracef("Found template: %s", util.TruncString(templateText))

			if strings.Contains(templateText[templateStartTokenLen:], templateStartToken) {
				// If the template itself contains a template, then proceed to first evaluate the inner template and
				// to evaluate the outer template in a later run
				sigolo.Trace("Template contains templates, inner templates are replaced first")
				newContent, err := t.replaceTemplateByPlaceholders(templateText[templateStartTokenLen:], placeholderToContent)
				if err != nil {
					return "", err
				}
				templateText = templateStartToken + newContent
			}

			key := util.Hash(templateText)

			sigolo.Tracef("Evaluate template: %s", util.TruncString(templateText))
			evaluatedTemplate, err := wikipedia.EvaluateTemplate(templateText, t.templateFolder, key)
			if err != nil {
				return "", err
			}

			// Replace the template by a placeholder. We do not directly replace the wikitext of the template with the
			// evaluated form because nested templates might lead to too long URLs.
			placeholderToContent[key] = evaluatedTemplate
			placeholder := fmt.Sprintf(templatePlaceholderTemplate, key)
			content = strings.Replace(content, originalTemplateText, placeholder, 1)
		}
	}
	sigolo.Tracef("Finished replacing templates in: %s", util.TruncString(content))
	return content, nil
}
