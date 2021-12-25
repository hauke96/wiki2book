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
	"regexp"
	"sort"
	"strings"
)

const TEMPLATE_FOLDER = "./templates/"

const IMAGE_REGEX = `\[\[((Datei|File):([^|^\]]*))(\|([^\]]*))?]]`

var images = []string{}

func Parse(content string, title string) Article {
	tokenMap := map[string]string{}

	content = tokenize(content, tokenMap)

	sigolo.Info("Token map length: %d", len(tokenMap))

	// print some debug information if wanted
	if sigolo.LogLevel >= sigolo.LOG_DEBUG {
		sigolo.Debug(content)

		keys := make([]string, 0, len(tokenMap))
		for k := range tokenMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			sigolo.Debug("%s : %s\n", k, tokenMap[k])
		}
	}

	return Article{
		Title:    title,
		TokenMap: tokenMap,
		Images:   images,
		Content:  content,
	}
}

func evaluateTemplates(content string, tokenMap map[string]string) string {
	regex := regexp.MustCompile(`\{\{([a-zA-Z](.|\n|\r)*?)}}`)
	content = regex.ReplaceAllStringFunc(content, func(match string) string {
		evaluatedTemplate := ""
		var err error = nil

		hash := sha1.New()
		hash.Write([]byte(match))
		key := hex.EncodeToString(hash.Sum(nil))

		if hasLocalTemplate(key) {
			matchSubString := match[:int(math.Min(float64(len(match)), 30))]
			if len(match) > 30 {
				matchSubString += "..."
			}
			sigolo.Info("Template \"%s\" already evaluated, use cached version", matchSubString)

			evaluatedTemplate, err = getTemplate(key)
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

			err = saveTemplate(key, evaluatedTemplate)
			if err != nil {
				sigolo.Stack(err)
				return ""
			}
		}

		evaluatedTemplate = escapeImages(evaluatedTemplate)

		return tokenize(evaluatedTemplate, tokenMap)
	})
	return content
}

func hasLocalTemplate(key string) bool {
	templateFilepath := TEMPLATE_FOLDER + key

	file, err := os.Open(templateFilepath)
	if file == nil || errors.Is(err, os.ErrNotExist) {
		return false
	}
	defer file.Close()

	return true
}

func getTemplate(key string) (string, error) {
	templateFilepath := TEMPLATE_FOLDER + key

	content, err := ioutil.ReadFile(templateFilepath)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Error reading template %s from %s", key, templateFilepath))
	}

	return string(content), nil
}

func saveTemplate(key string, evaluatedTemplate string) error {
	outputFilepath := TEMPLATE_FOLDER + key

	// Create the output folder
	err := os.Mkdir(TEMPLATE_FOLDER, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, fmt.Sprintf("Unable to create output folder %s", TEMPLATE_FOLDER))
	}

	// Create the output file
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

// escapeImages returns the list of all images and also escapes the image names in the content
func escapeImages(content string) string {
	var result []string

	// Remove videos and gifs
	regex := regexp.MustCompile(`\[\[((Datei|File):.*?\.(webm|gif|ogv|mp3|mp4|ogg|wav)).*(]]|\|)`)
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile(IMAGE_REGEX)
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		filePrefix := submatch[2]
		filename := submatch[3]
		filename = strings.ReplaceAll(filename, " ", "_")
		filename = strings.ReplaceAll(filename, "%20", "_")
		filename = filePrefix + ":" + strings.ToUpper(string(filename[0])) + filename[1:]

		content = strings.ReplaceAll(content, submatch[1], filename)

		result = append(result, filename)

		sigolo.Debug("Found image: %s", filename)
	}

	images = append(images, result...)

	sigolo.Info("Found and embedded %d images", len(submatches))
	return content
}
