package parser

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/api"
	"github.com/hauke96/wiki2book/src/util"
	"strings"
)

func evaluateTemplates(content string, templateFolder string) string {
	containsUnevaluatedTemplated := strings.Contains(content, "{{")

	for containsUnevaluatedTemplated {
		possibleTemplateBlocks := strings.Split(content, "{{")

		// Let's assume for a second, that we'll evaluate all templates
		containsUnevaluatedTemplated = false

		// The index 0 can be skipped since it's the first block above the first template, where no evaluation is needed.
		for i := 1; i < len(possibleTemplateBlocks); i++ {
			possibleTemplateBlock := possibleTemplateBlocks[i]

			if strings.Contains(possibleTemplateBlock, "}}") {
				// If this block contains a closing template marker, then we can evaluate the space between.
				templateContent := strings.Split(possibleTemplateBlock, "}}")[0]
				templateContent = fmt.Sprintf("{{%s}}", templateContent)

				key := util.Hash(templateContent)

				evaluatedTemplate, err := api.EvaluateTemplate(templateContent, templateFolder, key)
				if err != nil {
					sigolo.Stack(err)
					return ""
				}

				// Replace the template by its evaluated form. Only do this for the first occurrence to not alter templates
				// that will be handled by this loop anyway.
				content = strings.Replace(content, templateContent, evaluatedTemplate, 1)
			} else {
				// The current block is a template but doesn't have an end-tag -> it contains an inner template and will
				// not be evaluated yet -> we need another round of the outer loop to also evaluate this one.
				containsUnevaluatedTemplated = true
			}
		}
	}

	return content
}
