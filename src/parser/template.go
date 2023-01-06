package parser

import (
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/api"
	"github.com/hauke96/wiki2book/src/util"
	"strings"
)

func (t *Tokenizer) evaluateTemplates(content string) string {
	lastOpeningTemplateIndex := -1

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

			// Replace the template by its evaluated form. Only do this for the first occurrence to not alter templates
			// that will be handled by this loop anyway.
			content = strings.Replace(content, templateText, evaluatedTemplate, 1)

			// Reset loop to start from scratch
			i = 0
			lastOpeningTemplateIndex = -1
		}
	}

	return content
}
