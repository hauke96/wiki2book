package parser

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/api"
	"regexp"
)

func evaluateTemplates(content string, templateFolder string) string {
	regex := regexp.MustCompile(`\{\{((.|\n|\r)*?)}}`)
	content = regex.ReplaceAllStringFunc(content, func(match string) string {
		hash := sha1.New()
		hash.Write([]byte(match))
		key := hex.EncodeToString(hash.Sum(nil))

		evaluatedTemplate, err := api.EvaluateTemplate(match, templateFolder, key)
		if err != nil {
			sigolo.Stack(err)
			return ""
		}

		return evaluatedTemplate
	})
	return content
}
