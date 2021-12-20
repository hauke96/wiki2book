package parser

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/wiki"
	"github.com/hauke96/wiki2book/src/api"
	"regexp"
	"strings"
)

const TOKEN_TEMPLATE = "$$TOKEN_%s_%d$$"
const TOKEN_INTERNAL_LINK = "TOKEN_INTERNAL_LINK"
const TOKEN_EXTERNAL_LINK = "TOKEN_EXTERNAL_LINK"

// Marker do not appear in the token map
const MARKER_BOLD_OPEN = "$$MARKER_BOLD_OPEN$$"
const MARKER_BOLD_CLOSE = "$$MARKER_BOLD_CLOSE$$"
const MARKER_ITALIC_OPEN = "$$MARKER_ITALIC_OPEN$$"
const MARKER_ITALIC_CLOSE = "$$MARKER_ITALIC_CLOSE$$"

var tokenCounter = 0

// https://www.mediawiki.org/wiki/Markup_spec
func parse(wikiPageDto *api.WikiPageDto) wiki.Article {
	content := moveCitationsToEnd(wikiPageDto.Parse.Wikitext.Content)
	content = removeUnwantedTags(content)
	content = evaluateTemplates(content)

	//content, images := processImages(content)
	return wiki.Article{
		Title: wikiPageDto.Parse.Title,
		//Images:  images,
		Content: content,
	}
}

func getToken(tokenType string) string {
	token := fmt.Sprintf(TOKEN_TEMPLATE, tokenType, tokenCounter)
	tokenCounter++
	return token
}

func tokenize(content string, tokenMap map[string]string) string {
	tokenizationHappened := false
	for {
		content, tokenizationHappened = parseBoldAndItalic(content, tokenMap)
		if tokenizationHappened {
			continue
		}

		content, tokenizationHappened = parseInternalLinks(content, tokenMap)
		if tokenizationHappened {
			continue
		}

		content, tokenizationHappened = parseExternalLinks(content, tokenMap)
		if tokenizationHappened {
			continue
		}

		break
	}

	return content
}

func parseBoldAndItalic(content string, tokenMap map[string]string) (string, bool) {
	index := strings.Index(content, "''")
	if index != -1 {
		content, _, _, _ = tokenizeBoldAndItalic(content, index, tokenMap, false, false)
		return content, true
	}
	return content, false
}

// tokenizeByRegex applies the regex which must have exactly one group. The tokenized content is returned and a flag
// saying if something changed (i.e. is a tokenization happened).
//func tokenizeByRegex(content string, tokenMap map[string]string, regexString string, tokenType string) (string, bool) {
//	regex := regexp.MustCompile(regexString)
//	matches := regex.FindStringSubmatch(content)
//	if len(matches) != 0 {
//		content = processMatch(content, tokenMap, matches[0], matches[1], tokenType)
//		return content, true
//	}
//	return content, false
//}
//
//func processMatch(content string, tokenMap map[string]string, wholeMatch string, untokenizedMatch string, tokenType string) string {
//	token := getToken(tokenType)
//	tokenizedString := tokenize(untokenizedMatch, tokenMap)
//	tokenMap[token] = tokenizedString
//	return strings.Replace(content, wholeMatch, token, 1)
//}

func tokenizeBoldAndItalic(content string, index int, tokenMap map[string]string, isBoldOpen bool, isItalicOpen bool) (string, int, bool, bool) {
	for {
		// iIn case of last opened italic marker
		sigolo.Info("Check index %d of %d long content: %s", index, len(content), content[index:index+3])
		if content[index:index+3] == "'''" {
			if !isBoldOpen {
				// -3 +3 to replace the ''' as well
				content = strings.Replace(content, content[index:index+3], MARKER_BOLD_OPEN, 1)
				index = index + len(MARKER_BOLD_OPEN)

				// Check for further nested italic tags
				content, index, isBoldOpen, isItalicOpen = tokenizeBoldAndItalic(content, index, tokenMap, true, isItalicOpen)
			} else {
				// +3 to replace the '''
				content = strings.Replace(content, content[index:index+3], MARKER_BOLD_CLOSE, 1)

				// -3 because of the ''' we replaced above
				return content, index + len(MARKER_BOLD_CLOSE), false, isItalicOpen
			}
		} else if content[index:index+2] == "''" {
			if !isItalicOpen {
				// +2 to replace the ''
				content = strings.Replace(content, content[index:index+2], MARKER_ITALIC_OPEN, 1)
				index = index + len(MARKER_ITALIC_OPEN)

				// Check for further nested italic tags
				content, index, isBoldOpen, isItalicOpen = tokenizeBoldAndItalic(content, index, tokenMap, isBoldOpen, true)
			} else {
				// +2 to replace the ''
				content = strings.Replace(content, content[index:index+2], MARKER_ITALIC_CLOSE, 1)

				// -2 because of the '' we replaced above
				return content, index + len(MARKER_ITALIC_CLOSE), isBoldOpen, false
			}
		} else {
			index++
		}

		if !isBoldOpen && !isItalicOpen {
			break
		}
	}

	return content, index, false, false
}

func parseInternalLinks(content string, tokenMap map[string]string) (string,bool) {
	tokenizationHappened := false
	regex := regexp.MustCompile(`\[\[(.*?)]]`)
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		token := getToken(TOKEN_INTERNAL_LINK)
		tokenMap[token] = submatch[1]
		content = strings.Replace(content, submatch[0], token, 1)
		tokenizationHappened = true
	}

	return content, tokenizationHappened
}

func parseExternalLinks(content string, tokenMap map[string]string) (string,bool) {
	tokenizationHappened := false
	regex := regexp.MustCompile(`([^\[])\[([^\[].*?)\]`)
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		token := getToken(TOKEN_EXTERNAL_LINK)
		tokenMap[token] = submatch[2]
		content = strings.Replace(content, submatch[0], submatch[1] + token, 1)
		tokenizationHappened = true
	}

	return content, tokenizationHappened
}















func removeUnwantedTags(content string) string {
	regex := regexp.MustCompile("<references.*?\\/>\n?")
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile("\\[\\[Kategorie:.*?]]\n?")
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile("\\{\\{Gesprochener Artikel(.|\\n|\\r)*?}}\n?")
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile("\\{\\{Exzellent(.|\\n|\\r)*?}}\n?")
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile("\\{\\{Normdaten(.|\\n|\\r)*?}}\n?")
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile("\\{\\{Hauptartikel(.|\\n|\\r)*?}}\n?")
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile("\\{\\{Begriffsklärungshinweis(.|\\n|\\r)*?}}\n?")
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile("\\{\\{Weiterleitungshinweis(.|\\n|\\r)*?}}\n?")
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile("\\{\\{Dieser Artikel(.|\\n|\\r)*?}}\n?")
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile("\\{\\{.*(box|Box).*(.|\\n|\\r)*?}}\n?")
	content = regex.ReplaceAllString(content, "")

	return content
}

func moveCitationsToEnd(content string) string {
	counter := 0
	citations := ""

	regex := regexp.MustCompile("<ref.*?>(.*?)</ref>")
	content = regex.ReplaceAllStringFunc(content, func(match string) string {
		counter++
		if counter > 1 {
			citations += "<br>"
		}
		citations += fmt.Sprintf("\n[%d] %s", counter, match)
		return fmt.Sprintf("[%d]", counter)
	})

	return content + citations
}

func evaluateTemplates(content string) string {
	regex := regexp.MustCompile("\\{\\{(.*?)}}")
	content = regex.ReplaceAllStringFunc(content, func(match string) string {
		evaluatedTemplate, err := api.EvaluateTemplate(match)
		if err != nil {
			sigolo.Stack(err)
			return ""
		}
		return evaluatedTemplate
	})
	return content
}

// processImages returns the list of all images and also escapes the image names in the content
func processImages(content string) (string, []string) {
	var result []string

	// Remove videos and gifs
	regex := regexp.MustCompile("\\[\\[((Datei|File):.*?\\.(webm|gif|ogv|mp3|mp4)).*(]]|\\|)")
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile("\\[\\[((Datei|File):(.*?))(]]|\\|)(.*?)]]")
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		filename := strings.ReplaceAll(submatch[3], " ", "_")
		filename = submatch[2] + ":" + strings.ToUpper(string(filename[0])) + filename[1:]

		content = strings.ReplaceAll(content, submatch[1], filename)

		result = append(result, filename)

		sigolo.Debug("Found image: %s", filename)
	}

	sigolo.Info("Found and embedded %d images", len(submatches))
	return content, result
}
