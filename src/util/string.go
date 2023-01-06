package util

import (
	"math"
	"strings"
	"unicode/utf8"
)

func TruncString(content string) string {
	firstPartOfContent := content[:int(math.Min(float64(len(content)), 50))]
	if len(content) > 50 {
		firstPartOfContent += "..."
	}
	return firstPartOfContent
}

func RemoveLastChar(s string) string {
	_, sizeOfLastChar := utf8.DecodeLastRuneInString(s)
	return s[:len(s)-sizeOfLastChar]
}

func ElementHasPrefix(element string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(element, prefix) {
			return true
		}
	}
	return false
}
