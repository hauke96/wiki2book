package parser

import (
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/api"
	"github.com/hauke96/wiki2book/src/util"
	"regexp"
)

func evaluateTemplates(content string, templateFolder string) string {
	regex := regexp.MustCompile(`\{\{((.|\n|\r)*?)}}`)
	content = regex.ReplaceAllStringFunc(content, func(match string) string {
		key := util.Hash(match)

		evaluatedTemplate, err := api.EvaluateTemplate(match, templateFolder, key)
		if err != nil {
			sigolo.Stack(err)
			return ""
		}

		return evaluatedTemplate
	})
	return content
}
