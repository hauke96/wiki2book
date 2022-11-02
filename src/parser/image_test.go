package parser

import (
	"github.com/hauke96/wiki2book/src/test"
	"testing"
)

func TestEscapeImages_removeVideos(t *testing.T) {
	unwantedMedia := []string{"webm", "gif", "ogv", "mp3", "mp4", "ogg", "wav"}

	for _, extension := range unwantedMedia {
		content := "[[Datei:foo." + extension + "]][[File:bar." + extension + "|some|further|settings]]"
		content = escapeImages(content)
		test.AssertEmptyString(t, content)
		test.AssertEqual(t, 0, len(images))
	}
}

func TestEscapeImages_escapeFileNames(t *testing.T) {
	content := "[[Datei:some photo.png|with|properties]]"
	content = escapeImages(content)
	test.AssertEqual(t, "[[Datei:Some_photo.png|with|properties]]", content)
	test.AssertEqual(t, []string{"Datei:Some_photo.png"}, images)
}
