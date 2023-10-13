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

	test.AssertMapEqual(t, map[string]interface{}{}, tokenizer.getTokenMap())
}

func TestParseGalleries_emptyGallery(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseGalleries(`foo
<gallery>
</gallery>
bar`)

	test.AssertEqual(t, `foo
bar`, content)

	test.AssertMapEqual(t, map[string]interface{}{}, tokenizer.getTokenMap())
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

	test.AssertMapEqual(t, map[string]interface{}{}, tokenizer.getTokenMap())
}

func TestParseImages_inlineHappyPath(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseImages("foo [[file:image.jpg]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE_INLINE+"_0$$ bar", content)
}

func TestParseImages_withEscaping(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseImages("blubb [[file:nice image.jpg]] bar")
	test.AssertEqual(t, "blubb $$TOKEN_"+TOKEN_IMAGE_INLINE+"_0$$ bar", content)
	test.AssertMapEqual(t, map[string]interface{}{
		"$$TOKEN_" + TOKEN_IMAGE_INLINE + "_0$$": InlineImageToken{
			Filename: "foo/Nice_image.jpg",
			SizeX:    -1,
			SizeY:    -1,
		},
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("foo [[file:nice image.gif]] bar")
	test.AssertEqual(t, "foo  bar", content)
	test.AssertMapEqual(t, map[string]interface{}{}, tokenizer.getTokenMap())
}

func TestParseImages_ignoreParameters(t *testing.T) {
	for _, param := range imageNonInlineParameters {
		tokenizer := NewTokenizer("foo", "bar")
		content := tokenizer.parseImages(fmt.Sprintf("foo [[file:image.jpg|%s]] bar", param))
		test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_0$$ bar", content)
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
		test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE_INLINE+"_0$$ bar", content)
	}
}

func TestParseImages_smallSizesProduceInlineImage(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseImages("foo [[file:image.jpg|99x49px]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE_INLINE+"_0$$ bar", content)

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("foo [[file:image.jpg|101x51px]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_0$$ bar", content)
}

func TestParseImages_withSizes(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseImages("blubb [[file:image.jpg|100x200px]] bar")
	test.AssertEqual(t, "blubb $$TOKEN_"+TOKEN_IMAGE+"_0$$ bar", content)
	test.AssertMapEqual(t, map[string]interface{}{
		"$$TOKEN_" + TOKEN_IMAGE + "_0$$": ImageToken{
			Filename:        "foo/Image.jpg",
			CaptionTokenKey: "",
			SizeX:           100,
			SizeY:           200,
		},
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("blubb [[file:image.jpg|x200px]] bar")
	test.AssertEqual(t, "blubb $$TOKEN_"+TOKEN_IMAGE+"_0$$ bar", content)
	test.AssertMapEqual(t, map[string]interface{}{
		"$$TOKEN_" + TOKEN_IMAGE + "_0$$": ImageToken{
			Filename:        "foo/Image.jpg",
			CaptionTokenKey: "",
			SizeX:           -1,
			SizeY:           200,
		},
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("blubb [[file:image.jpg|200px]] bar")
	test.AssertEqual(t, "blubb $$TOKEN_"+TOKEN_IMAGE+"_0$$ bar", content)
	test.AssertMapEqual(t, map[string]interface{}{
		"$$TOKEN_" + TOKEN_IMAGE + "_0$$": ImageToken{
			Filename:        "foo/Image.jpg",
			CaptionTokenKey: "",
			SizeX:           200,
			SizeY:           -1,
		},
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("blubb [[file:image.jpg|mini|200px]] bar")
	test.AssertEqual(t, "blubb $$TOKEN_"+TOKEN_IMAGE+"_0$$ bar", content)
	test.AssertMapEqual(t, map[string]interface{}{
		"$$TOKEN_" + TOKEN_IMAGE + "_0$$": ImageToken{
			Filename:        "foo/Image.jpg",
			CaptionTokenKey: "",
			SizeX:           200,
			SizeY:           -1,
		},
	}, tokenizer.getTokenMap())
}

func TestParseImages_withCaption(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")

	content := tokenizer.parseImages("foo [[file:image.jpg|10x20px|mini|some caption]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_1$$ bar", content)
	test.AssertMapEqual(t, map[string]interface{}{
		"$$TOKEN_" + TOKEN_IMAGE + "_1$$": ImageToken{
			Filename:        "foo/Image.jpg",
			CaptionTokenKey: fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_IMAGE_CAPTION, 0),
			SizeX:           10,
			SizeY:           20,
		},
		"$$TOKEN_" + TOKEN_IMAGE_CAPTION + "_0$$": "some caption",
	}, tokenizer.getTokenMap())
}

func TestParseImages_withCaptionEndingWithLinks(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseImages("foo [[file:image.jpg|mini|some [https://foo.com link]]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_2$$ bar", content)
	test.AssertMapEqual(t, map[string]interface{}{
		"$$TOKEN_" + TOKEN_IMAGE + "_2$$": ImageToken{
			Filename:        "foo/Image.jpg",
			CaptionTokenKey: fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_IMAGE_CAPTION, 1),
			SizeX:           -1,
			SizeY:           -1,
		},
		"$$TOKEN_" + TOKEN_IMAGE_CAPTION + "_1$$": "some $$TOKEN_" + TOKEN_EXTERNAL_LINK + "_0$$",
		"$$TOKEN_" + TOKEN_EXTERNAL_LINK + "_0$$": &ExternalLinkToken{
			URL:      "https://foo.com",
			LinkText: "link",
		},
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("foo [[file:image.jpg|mini|some [[article]]]]")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_2$$", content)
	test.AssertMapEqual(t, map[string]interface{}{
		"$$TOKEN_" + TOKEN_IMAGE + "_2$$": ImageToken{
			Filename:        "foo/Image.jpg",
			CaptionTokenKey: fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_IMAGE_CAPTION, 1),
			SizeX:           -1,
			SizeY:           -1,
		},
		"$$TOKEN_" + TOKEN_IMAGE_CAPTION + "_1$$": "some $$TOKEN_" + TOKEN_INTERNAL_LINK + "_0$$",
		"$$TOKEN_" + TOKEN_INTERNAL_LINK + "_0$$": &InternalLinkToken{
			ArticleName: "article",
			LinkText:    "article",
		},
	}, tokenizer.getTokenMap())
}

func TestParseImages_withCaptionAndTrailingParameter(t *testing.T) {
	config.Current.IgnoredImageParams = []string{
		"ignoredParam",
	}

	tokenizer := NewTokenizer("foo", "bar")

	content := tokenizer.parseImages("foo [[file:image.jpg|10x20px|mini|some caption|ignoredParam=blubb]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_1$$ bar", content)
	test.AssertMapEqual(t, map[string]interface{}{
		"$$TOKEN_" + TOKEN_IMAGE + "_1$$": ImageToken{
			Filename:        "foo/Image.jpg",
			CaptionTokenKey: fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_IMAGE_CAPTION, 0),
			SizeX:           10,
			SizeY:           20,
		},
		"$$TOKEN_" + TOKEN_IMAGE_CAPTION + "_0$$": "some caption",
	}, tokenizer.getTokenMap())
}
