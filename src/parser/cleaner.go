package parser

import (
	"fmt"
	"regexp"
)

func clean(content string) string {
	content = removeUnwantedTags(content)
	content = moveCitationsToEnd(content)
	return content
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
