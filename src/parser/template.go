package parser

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/api"
	"github.com/pkg/errors"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"regexp"
)

func evaluateTemplates(content string, templateFolder string) string {
	regex := regexp.MustCompile(`\{\{((.|\n|\r)*?)}}`)
	content = regex.ReplaceAllStringFunc(content, func(match string) string {
		evaluatedTemplate := ""
		var err error = nil

		hash := sha1.New()
		hash.Write([]byte(match))
		key := hex.EncodeToString(hash.Sum(nil))

		// TODO remove if, the api already takes care of existing files
		if hasLocalTemplate(key, templateFolder) {
			matchSubString := match[:int(math.Min(float64(len(match)), 30))]
			if len(match) > 30 {
				matchSubString += "..."
			}
			sigolo.Info("Template \"%s\" already evaluated, use cached version", matchSubString)

			evaluatedTemplate, err = getTemplate(key, templateFolder)
			if err != nil {
				sigolo.Stack(err)
				return ""
			}
		} else {
			evaluatedTemplate, err = api.EvaluateTemplate(match, key)
			if err != nil {
				sigolo.Stack(err)
				return ""
			}
		}

		return evaluatedTemplate
	})
	return content
}

func hasLocalTemplate(key string, templateFolder string) bool {
	templateFilepath := filepath.Join(templateFolder, key)

	file, err := os.Open(templateFilepath)
	if file == nil || errors.Is(err, os.ErrNotExist) {
		return false
	}
	defer file.Close()

	return true
}

func getTemplate(key string, templateFolder string) (string, error) {
	templateFilepath := filepath.Join(templateFolder, key)

	content, err := ioutil.ReadFile(templateFilepath)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Error reading template %s from %s", key, templateFilepath))
	}

	return string(content), nil
}
