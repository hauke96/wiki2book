package parser

import (
	"fmt"
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

	// cleanup
	images = make([]string, 0)
}

func TestEscapeImages_leadingNonAscii(t *testing.T) {
	content := "[[Datei:öäü.png|with|properties]]"
	content = escapeImages(content)
	test.AssertEqual(t, "[[Datei:Öäü.png|with|properties]]", content)
	test.AssertEqual(t, []string{"Datei:Öäü.png"}, images)

	// cleanup
	images = make([]string, 0)
}

func TestEscapeImages_leadingSpecialChar(t *testing.T) {
	content := "[[Datei:\"öäü\".png|with|properties]]"
	content = escapeImages(content)
	test.AssertEqual(t, "[[Datei:\"öäü\".png|with|properties]]", content)
	test.AssertEqual(t, []string{"Datei:\"öäü\".png"}, images)

	// cleanup
	images = make([]string, 0)
}

func TestParseGalleries(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseGalleries(`foo
<gallery>file0.jpg
file1.jpg|captiion
</gallery>
bar
 <gallery some="parameter">
File:file2.jpg|test123
  File:file 3.jpg
</gallery>blubb`)

	test.AssertEqual(t, `foo
[[File:File0.jpg|mini]]
[[File:File1.jpg|mini|captiion]]
bar
[[File:File2.jpg|mini|test123]]
[[File:File_3.jpg|mini]]
blubb`, content)

	test.AssertEqual(t, map[string]string{}, tokenizer.getTokenMap())
}

func TestParseGalleries_emptyGallery(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseGalleries(`foo
<gallery>
</gallery>
bar`)

	test.AssertEqual(t, `foo
bar`, content)

	test.AssertEqual(t, map[string]string{}, tokenizer.getTokenMap())
}

func TestParseImagemaps(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseImageMaps(`foo
<imagemap>File:picture.jpg
some
stuff
</imagemap>
bar
<imagemap some="parameter">
Image:picture.jpg
some stuff
</imagemap>
blubb`)

	test.AssertEqual(t, `foo
[[File:Picture.jpg]]
bar
[[Image:Picture.jpg]]
blubb`, content)

	test.AssertEqual(t, map[string]string{}, tokenizer.getTokenMap())
}

func TestParseImages(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseImages("foo [[Datei:image.jpg]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE_INLINE+"_1$$ bar", content)

	for _, param := range imageNonInlineParameters {
		tokenizer = NewTokenizer("foo", "bar")
		content = tokenizer.parseImages(fmt.Sprintf("foo [[Datei:image.jpg|%s]] bar", param))
		test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_1$$ bar", content)
	}

	for _, param := range imageIgnoreParameters {
		tokenizer := NewTokenizer("foo", "bar")
		content = tokenizer.parseImages(fmt.Sprintf("foo [[Datei:image.jpg|%s]] bar", param))
		test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE_INLINE+"_1$$ bar", content)
	}

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("foo [[Datei:image.jpg|100x50px]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE_INLINE+"_2$$ bar", content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("foo [[Datei:image.jpg|101x51px]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_2$$ bar", content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("foo [[Datei:image.jpg|10x20px|mini|some caption]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_3$$ bar", content)
	test.AssertEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_IMAGE + "_3$$":          fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_IMAGE_FILENAME, 0, TOKEN_IMAGE_CAPTION, 2, TOKEN_IMAGE_SIZE, 1),
		"$$TOKEN_" + TOKEN_IMAGE_FILENAME + "_0$$": "foo/image.jpg",
		"$$TOKEN_" + TOKEN_IMAGE_CAPTION + "_2$$":  "some caption",
		"$$TOKEN_" + TOKEN_IMAGE_SIZE + "_1$$":     "10x20",
	}, tokenizer.getTokenMap())
}
