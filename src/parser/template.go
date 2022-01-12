package parser

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/api"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func evaluateTemplates(content string, templateFolder string) string {
	regex := regexp.MustCompile(`\{\{((.|\n|\r)*?)}}`)
	content = regex.ReplaceAllStringFunc(content, func(match string) string {
		evaluatedTemplate := ""
		var err error = nil

		hash := sha1.New()
		hash.Write([]byte(match))
		key := hex.EncodeToString(hash.Sum(nil))

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
			evaluatedTemplate, err = api.EvaluateTemplate(match)
			if err != nil {
				sigolo.Stack(err)
				return ""
			}

			err = saveTemplate(key, evaluatedTemplate, templateFolder)
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

func saveTemplate(key string, evaluatedTemplate string, templateFolder string) error {
	// Create the output folder
	info, err := os.Stat(templateFolder)
	if err == nil && !info.IsDir() {
		return errors.New(fmt.Sprintf("Given path exists but is not a folder: %s", templateFolder))
	} else if os.IsNotExist(err) {
		err := os.Mkdir(templateFolder, os.ModePerm)
		if err != nil && !os.IsExist(err) {
			return errors.Wrap(err, fmt.Sprintf("Unable to create output folder %s", templateFolder))
		}
	}

	// Create the output file
	outputFilepath := filepath.Join(templateFolder, key)
	outputFile, err := os.Create(outputFilepath)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to create output file for template %s", key))
	}
	defer outputFile.Close()

	// Write the body to file
	_, err = io.Copy(outputFile, strings.NewReader(evaluatedTemplate))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable copy evaluated template of %s to file %s", key, outputFilepath))
	}

	sigolo.Debug("Template result saved to %s", outputFilepath)
	return nil
}
