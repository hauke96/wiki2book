package parser

import (
	"fmt"
	"github.com/hauke96/wiki2book/src/test"
	"testing"
)

func TestImageRegex(t *testing.T) {
	valid := []string{
		"[[Datei:foo]]",
		"[[Datei:foo.png]]",
		"[[Datei:foo.png|mini]]",
		"[[Datei:foo|mini]]",
		"[[Datei:foo.jpg|mini|16px]]",
		"[[File:foo.png]]",
		"[[datei:foo]]",
		"[[dATEI:foo.png]]",
		"[[datei:foo.png|mini]]",
		"[[dATEi:foo|mini]]",
		"[[DateI:foo.jpg|mini|16px]]",
		"[[file:foo.png]]",
	}

	for _, s := range valid {
		test.AssertTrue(t, imageRegex.MatchString(s))
	}

	invalid := []string{
		"",
		"Datei.foo.png",
		"[Datei:foo.png]",
		"[[Fiel:foo.png]]",
		"[[foo.png]]",
	}

	for _, s := range invalid {
		test.AssertFalse(t, imageRegex.MatchString(s))
	}
}

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
Datei:file1.jpg|captiion
</gallery>
bar
 <gallery some="parameter">
File:file2.jpg|test123
  file 3.jpg
</gallery>blubb`)

	test.AssertEqual(t, `foo
[[File:File0.jpg|mini]]
[[Datei:File1.jpg|mini|captiion]]
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

func TestParseImages_inlineHappyPath(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseImages("foo [[Datei:image.jpg]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE_INLINE+"_1$$ bar", content)
}

func TestParseImages_ignoreParameters(t *testing.T) {
	for _, param := range imageNonInlineParameters {
		tokenizer := NewTokenizer("foo", "bar")
		content := tokenizer.parseImages(fmt.Sprintf("foo [[Datei:image.jpg|%s]] bar", param))
		test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_1$$ bar", content)
	}
}

func TestParseImages_ignoreParametersOnInlineImage(t *testing.T) {
	for _, param := range imageIgnoreParameters {
		tokenizer := NewTokenizer("foo", "bar")
		content := tokenizer.parseImages(fmt.Sprintf("foo [[Datei:image.jpg|%s]] bar", param))
		test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE_INLINE+"_1$$ bar", content)
	}
}

func TestParseImages_smallSizesProduceInlineImage(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseImages("foo [[Datei:image.jpg|99x49px]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE_INLINE+"_2$$ bar", content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("foo [[Datei:image.jpg|101x51px]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_2$$ bar", content)
}

func TestParseImages_withSizes(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseImages("foo [[Datei:image.jpg|100x200px]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_2$$ bar", content)
	test.AssertEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_IMAGE + "_2$$":          fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_IMAGE_FILENAME, 0, TOKEN_IMAGE_SIZE, 1),
		"$$TOKEN_" + TOKEN_IMAGE_FILENAME + "_0$$": "foo/image.jpg",
		"$$TOKEN_" + TOKEN_IMAGE_SIZE + "_1$$":     "100x200",
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("foo [[Datei:image.jpg|x200px]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_2$$ bar", content)
	test.AssertEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_IMAGE + "_2$$":          fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_IMAGE_FILENAME, 0, TOKEN_IMAGE_SIZE, 1),
		"$$TOKEN_" + TOKEN_IMAGE_FILENAME + "_0$$": "foo/image.jpg",
		"$$TOKEN_" + TOKEN_IMAGE_SIZE + "_1$$":     "x200",
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("foo [[Datei:image.jpg|200px]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_2$$ bar", content)
	test.AssertEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_IMAGE + "_2$$":          fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_IMAGE_FILENAME, 0, TOKEN_IMAGE_SIZE, 1),
		"$$TOKEN_" + TOKEN_IMAGE_FILENAME + "_0$$": "foo/image.jpg",
		"$$TOKEN_" + TOKEN_IMAGE_SIZE + "_1$$":     "200x",
	}, tokenizer.getTokenMap())
}

func TestParseImages_withCaption(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")

	content := tokenizer.parseImages("foo [[Datei:image.jpg|10x20px|mini|some caption]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_3$$ bar", content)
	test.AssertEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_IMAGE + "_3$$":          fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_IMAGE_FILENAME, 0, TOKEN_IMAGE_CAPTION, 2, TOKEN_IMAGE_SIZE, 1),
		"$$TOKEN_" + TOKEN_IMAGE_FILENAME + "_0$$": "foo/image.jpg",
		"$$TOKEN_" + TOKEN_IMAGE_CAPTION + "_2$$":  "some caption",
		"$$TOKEN_" + TOKEN_IMAGE_SIZE + "_1$$":     "10x20",
	}, tokenizer.getTokenMap())
}

func TestParseImages_withCaptionAndTrailingParameter(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")

	content := tokenizer.parseImages("foo [[Datei:image.jpg|10x20px|mini|some caption|verweis=]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_3$$ bar", content)
	test.AssertEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_IMAGE + "_3$$":          fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_IMAGE_FILENAME, 0, TOKEN_IMAGE_CAPTION, 2, TOKEN_IMAGE_SIZE, 1),
		"$$TOKEN_" + TOKEN_IMAGE_FILENAME + "_0$$": "foo/image.jpg",
		"$$TOKEN_" + TOKEN_IMAGE_CAPTION + "_2$$":  "some caption",
		"$$TOKEN_" + TOKEN_IMAGE_SIZE + "_1$$":     "10x20",
	}, tokenizer.getTokenMap())
}
