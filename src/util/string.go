package util

import (
	"io"
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

func HasAnyPrefix(element string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(element, prefix) {
			return true
		}
	}
	return false
}

func AllToLower(items []string) []string {
	result := make([]string, len(items))

	for i, s := range items {
		result[i] = strings.ToLower(s)
	}

	return result
}

func ReaderToString(reader io.Reader) string {
	if reader != nil {
		buf := new(strings.Builder)
		_, err := io.Copy(buf, reader)
		if err == nil {
			return buf.String()
		}
	}

	return ""
}
