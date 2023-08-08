package parser

import (
	"fmt"
	"testing"
	"wiki2book/config"
	"wiki2book/test"
)

func TestEscapeImages_removeVideos(t *testing.T) {
	var content string

	for _, extension := range config.Current.IgnoredMediaTypes {
		content = "file:foo." + extension
		content = escapeImages(content)
		test.AssertEmptyString(t, content)
		test.AssertEqual(t, 0, len(images))

		content = "File:bar." + extension + "|some|further|settings"
		content = escapeImages(content)
		test.AssertEmptyString(t, content)
		test.AssertEqual(t, 0, len(images))
	}
}

func TestEscapeImages_removeVideoWithMultilineCaption(t *testing.T) {
	content := `file:foo.webm|this caption<br>
is<br>
important!`
	content = escapeImages(content)
	test.AssertEqual(t, "", content)
	test.AssertEqual(t, 0, len(images))

	content = `file:foo.jpg|this caption<br>
is<br>
important!`
	content = escapeImages(content)
	test.AssertEqual(t, `file:Foo.jpg|this caption<br>
is<br>
important!`, content)
	test.AssertEqual(t, []string{"file:Foo.jpg"}, images)

	// cleanup
	images = make([]string, 0)
}

func TestEscapeImages_escapeFileNames(t *testing.T) {
	content := "file:some photo.png|with|properties"
	content = escapeImages(content)
	test.AssertEqual(t, "file:Some_photo.png|with|properties", content)
	test.AssertEqual(t, []string{"file:Some_photo.png"}, images)

	// cleanup
	images = make([]string, 0)
}

func TestEscapeImages_leadingNonAscii(t *testing.T) {
	content := "file:öäü.png|with|properties"
	content = escapeImages(content)
	test.AssertEqual(t, "file:Öäü.png|with|properties", content)
	test.AssertEqual(t, []string{"file:Öäü.png"}, images)

	// cleanup
	images = make([]string, 0)
}

func TestEscapeImages_leadingSpecialChar(t *testing.T) {
	content := "file:\"öäü\".png|with|properties"
	content = escapeImages(content)
	test.AssertEqual(t, "file:\"öäü\".png|with|properties", content)
	test.AssertEqual(t, []string{"file:\"öäü\".png"}, images)

	// cleanup
	images = make([]string, 0)
}

func TestParseGalleries(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseGalleries(`foo
<gallery>file0.jpg
file:file1.jpg|captiion
</gallery>
bar
 <gallery some="parameter">
File:file2.jpg|test123
  file 3.jpg
</gallery>blubb`)

	test.AssertEqual(t, `foo
[[File:File0.jpg|mini]]
[[file:File1.jpg|mini|captiion]]
bar
[[File:File2.jpg|mini|test123]]
[[File:File_3.jpg|mini]]
blubb`, content)

	test.AssertMapEqual(t, map[string]string{}, tokenizer.getTokenMap())
}

func TestParseGalleries_emptyGallery(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseGalleries(`foo
<gallery>
</gallery>
bar`)

	test.AssertEqual(t, `foo
bar`, content)

	test.AssertMapEqual(t, map[string]string{}, tokenizer.getTokenMap())
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

	test.AssertMapEqual(t, map[string]string{}, tokenizer.getTokenMap())
}

func TestParseImages_inlineHappyPath(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseImages("foo [[file:image.jpg]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE_INLINE+"_1$$ bar", content)
}

func TestParseImages_withEscaping(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseImages("foo [[file:nice image.jpg]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE_INLINE+"_1$$ bar", content)
	test.AssertMapEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_IMAGE_INLINE + "_1$$":   fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_IMAGE_FILENAME, 0),
		"$$TOKEN_" + TOKEN_IMAGE_FILENAME + "_0$$": "foo/Nice_image.jpg",
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("foo [[file:nice image.gif]] bar")
	test.AssertEqual(t, "foo  bar", content)
	test.AssertMapEqual(t, map[string]string{}, tokenizer.getTokenMap())
}

func TestParseImages_ignoreParameters(t *testing.T) {
	for _, param := range imageNonInlineParameters {
		tokenizer := NewTokenizer("foo", "bar")
		content := tokenizer.parseImages(fmt.Sprintf("foo [[file:image.jpg|%s]] bar", param))
		test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_1$$ bar", content)
	}
}

func TestParseImages_ignoreParametersOnInlineImage(t *testing.T) {
	config.Current.IgnoredImageParams = []string{
		"param",
		"otherParam",
	}

	for _, param := range config.Current.IgnoredImageParams {
		tokenizer := NewTokenizer("foo", "bar")
		content := tokenizer.parseImages(fmt.Sprintf("foo [[file:image.jpg|%s]] bar", param))
		test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE_INLINE+"_1$$ bar", content)
	}
}

func TestParseImages_smallSizesProduceInlineImage(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseImages("foo [[file:image.jpg|99x49px]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE_INLINE+"_2$$ bar", content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("foo [[file:image.jpg|101x51px]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_2$$ bar", content)
}

func TestParseImages_withSizes(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseImages("foo [[file:image.jpg|100x200px]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_2$$ bar", content)
	test.AssertMapEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_IMAGE + "_2$$":          fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_IMAGE_FILENAME, 0, TOKEN_IMAGE_SIZE, 1),
		"$$TOKEN_" + TOKEN_IMAGE_FILENAME + "_0$$": "foo/Image.jpg",
		"$$TOKEN_" + TOKEN_IMAGE_SIZE + "_1$$":     "100x200",
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("foo [[file:image.jpg|x200px]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_2$$ bar", content)
	test.AssertMapEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_IMAGE + "_2$$":          fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_IMAGE_FILENAME, 0, TOKEN_IMAGE_SIZE, 1),
		"$$TOKEN_" + TOKEN_IMAGE_FILENAME + "_0$$": "foo/Image.jpg",
		"$$TOKEN_" + TOKEN_IMAGE_SIZE + "_1$$":     "x200",
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("foo [[file:image.jpg|200px]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_2$$ bar", content)
	test.AssertMapEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_IMAGE + "_2$$":          fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_IMAGE_FILENAME, 0, TOKEN_IMAGE_SIZE, 1),
		"$$TOKEN_" + TOKEN_IMAGE_FILENAME + "_0$$": "foo/Image.jpg",
		"$$TOKEN_" + TOKEN_IMAGE_SIZE + "_1$$":     "200x",
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("foo [[file:image.jpg|mini|200px]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_2$$ bar", content)
	test.AssertMapEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_IMAGE + "_2$$":          fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_IMAGE_FILENAME, 0, TOKEN_IMAGE_SIZE, 1),
		"$$TOKEN_" + TOKEN_IMAGE_FILENAME + "_0$$": "foo/Image.jpg",
		"$$TOKEN_" + TOKEN_IMAGE_SIZE + "_1$$":     "200x",
	}, tokenizer.getTokenMap())
}

func TestParseImages_withCaption(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")

	content := tokenizer.parseImages("foo [[file:image.jpg|10x20px|mini|some caption]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_3$$ bar", content)
	test.AssertMapEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_IMAGE + "_3$$":          fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_IMAGE_FILENAME, 0, TOKEN_IMAGE_CAPTION, 2, TOKEN_IMAGE_SIZE, 1),
		"$$TOKEN_" + TOKEN_IMAGE_FILENAME + "_0$$": "foo/Image.jpg",
		"$$TOKEN_" + TOKEN_IMAGE_CAPTION + "_2$$":  "some caption",
		"$$TOKEN_" + TOKEN_IMAGE_SIZE + "_1$$":     "10x20",
	}, tokenizer.getTokenMap())
}

func TestParseImages_withCaptionEndingWithLinks(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseImages("foo [[file:image.jpg|mini|some [https://foo.com link]]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_5$$ bar", content)
	test.AssertMapEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_IMAGE + "_5$$":              fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_IMAGE_FILENAME, 3, TOKEN_IMAGE_CAPTION, 4),
		"$$TOKEN_" + TOKEN_IMAGE_FILENAME + "_3$$":     "foo/Image.jpg",
		"$$TOKEN_" + TOKEN_IMAGE_CAPTION + "_4$$":      "some $$TOKEN_" + TOKEN_EXTERNAL_LINK + "_2$$",
		"$$TOKEN_" + TOKEN_EXTERNAL_LINK + "_2$$":      fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_EXTERNAL_LINK_URL, 0, TOKEN_EXTERNAL_LINK_TEXT, 1),
		"$$TOKEN_" + TOKEN_EXTERNAL_LINK_URL + "_0$$":  "https://foo.com",
		"$$TOKEN_" + TOKEN_EXTERNAL_LINK_TEXT + "_1$$": "link",
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("foo [[file:image.jpg|mini|some [[article]]]]")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_5$$", content)
	test.AssertMapEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_IMAGE + "_5$$":                 fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_IMAGE_FILENAME, 3, TOKEN_IMAGE_CAPTION, 4),
		"$$TOKEN_" + TOKEN_IMAGE_FILENAME + "_3$$":        "foo/Image.jpg",
		"$$TOKEN_" + TOKEN_IMAGE_CAPTION + "_4$$":         "some $$TOKEN_" + TOKEN_INTERNAL_LINK + "_2$$",
		"$$TOKEN_" + TOKEN_INTERNAL_LINK + "_2$$":         fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_INTERNAL_LINK_ARTICLE, 0, TOKEN_INTERNAL_LINK_TEXT, 1),
		"$$TOKEN_" + TOKEN_INTERNAL_LINK_ARTICLE + "_0$$": "article",
		"$$TOKEN_" + TOKEN_INTERNAL_LINK_TEXT + "_1$$":    "article",
	}, tokenizer.getTokenMap())
}

func TestParseImages_withCaptionAndTrailingParameter(t *testing.T) {
	config.Current.IgnoredImageParams = []string{
		"ignoredParam",
	}

	tokenizer := NewTokenizer("foo", "bar")

	content := tokenizer.parseImages("foo [[file:image.jpg|10x20px|mini|some caption|ignoredParam=blubb]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_3$$ bar", content)
	test.AssertMapEqual(t, map[string]string{
		"$$TOKEN_" + TOKEN_IMAGE + "_3$$":          fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_IMAGE_FILENAME, 0, TOKEN_IMAGE_CAPTION, 2, TOKEN_IMAGE_SIZE, 1),
		"$$TOKEN_" + TOKEN_IMAGE_FILENAME + "_0$$": "foo/Image.jpg",
		"$$TOKEN_" + TOKEN_IMAGE_CAPTION + "_2$$":  "some caption",
		"$$TOKEN_" + TOKEN_IMAGE_SIZE + "_1$$":     "10x20",
	}, tokenizer.getTokenMap())
}
