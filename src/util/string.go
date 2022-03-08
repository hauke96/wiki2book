package util

import "math"

func TruncString(content string) string {
	firstPartOfContent := content[:int(math.Min(float64(len(content)), 50))]
	if len(content) > 50 {
		firstPartOfContent += "..."
	}
	return firstPartOfContent
}
