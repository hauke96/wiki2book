package parser

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/api"
	"github.com/hauke96/wiki2book/src/util"
	"strings"
)

const templatePlaceholderTemplate = "$$TEMPLATE_PLACEHOLDER_%s$$"

func (t *Tokenizer) evaluateTemplates(content string) string {
	lastOpeningTemplateIndex := -1
	placeholderToContent := map[string]string{}

	for i := 0; i < len(content)-1; i++ {
		cursor := content[i : i+2]

		if cursor == "{{" {
			lastOpeningTemplateIndex = i
		} else if lastOpeningTemplateIndex != -1 && cursor == "}}" {
			templateText := content[lastOpeningTemplateIndex : i+2]
			key := util.Hash(templateText)

			evaluatedTemplate, err := api.EvaluateTemplate(templateText, t.templateFolder, key)
			if err != nil {
				sigolo.Stack(err)
				return ""
			}

			// Replace the template by a placeholder. We do not directly replace the wikitext of the template with the
			//evaluated form because nested templates might lead to too long URLs.
			placeholderToContent[key] = evaluatedTemplate
			placeholder := fmt.Sprintf(templatePlaceholderTemplate, key)
			content = strings.Replace(content, templateText, placeholder, 1)

			// Resetting the cursor index to 0 (= -1 due to the i++ of the loop) works fine and is simple, even though
			// it's not the most efficient or fastest approach.
			i = -1
			lastOpeningTemplateIndex = -1
		}
	}

	// Replace all template placeholders with the actual content until no placeholders are unresolved. This is not very
	// elegant or fast but due to the nesting a simple and working approach.
	for len(placeholderToContent) != 0 {
		for key, template := range placeholderToContent {
			placeholder := fmt.Sprintf(templatePlaceholderTemplate, key)
			if strings.Contains(content, placeholder) {
				content = strings.ReplaceAll(content, placeholder, template)
				delete(placeholderToContent, key)
			}
		}
	}

	return content
}
