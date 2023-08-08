package parser

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"strings"
	"wiki2book/api"
	"wiki2book/util"
)

const templatePlaceholderTemplate = "$$TEMPLATE_PLACEHOLDER_%s$$"

// evaluateTemplates evaluates all templates including nested ones.
func (t *Tokenizer) evaluateTemplates(content string) string {
	// The lastOpeningTemplateIndex is set whenever a new opening template is found. When closing template brackets
	// are discovered, this variable contains the index of the corresponding opening brackets.
	lastOpeningTemplateIndex := -1

	// All evaluated templates are stored in this map. Replacing evaluated templates by placeholders reduces the length
	// of request URLs significantly and prevents errors due to too long URLs.
	placeholderToContent := map[string]string{}

	// Go through the content until the end. Whenever a template has been found, the variable "i" is reset to the
	// beginning so that this loop actually iterates the content over and over again until no unevaluated templates
	// are left.
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
			// evaluated form because nested templates might lead to too long URLs.
			placeholderToContent[key] = evaluatedTemplate
			placeholder := fmt.Sprintf(templatePlaceholderTemplate, key)
			content = strings.Replace(content, templateText, placeholder, 1)

			// Resetting the cursor index to 0 (= -1 due to the i++ of the loop) works fine and is simple, even though
			// it's not the most efficient or fastest approach.
			i = -1
			lastOpeningTemplateIndex = -1
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

	return content
}
