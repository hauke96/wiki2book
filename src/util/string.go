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

// GetTextAround returns the text around the given atIndex with the given size. Example: Is the size 3, then the char
// at the given location and 3 chars before and after are returned.
func GetTextAround(text string, atIndex int, areaSizeAround int) string {
	startIndex := int(math.Max(0, float64(atIndex-areaSizeAround)))
	endIndex := int(math.Min(float64(len(text)-1), float64(atIndex+areaSizeAround))) + 1
	return text[startIndex:endIndex]
}
