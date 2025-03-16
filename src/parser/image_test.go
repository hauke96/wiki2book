package parser

import (
	"fmt"
	"testing"
	"wiki2book/config"
	"wiki2book/test"
)

func TestEscapeImages_removeVideos(t *testing.T) {
	setup()
	tokenizer := NewTokenizer("foo", "bar")

	var content string

	for _, extension := range config.Current.IgnoredMediaTypes {
		content = "file:foo." + extension
		content = tokenizer.escapeImages(content)
		test.AssertEmptyString(t, content)
		test.AssertEqual(t, 0, len(tokenizer.images))

		content = "File:bar." + extension + "|some|further|settings"
		content = tokenizer.escapeImages(content)
		test.AssertEmptyString(t, content)
		test.AssertEqual(t, 0, len(tokenizer.images))
	}
}

func TestEscapeImages_keepPdfsEvenWhenIgnored(t *testing.T) {
	setup()
	tokenizer := NewTokenizer("foo", "bar")

	config.Current.ConvertPdfToPng = true

	var content string

	content = "File:Foo.pdf"
	escapedContent := tokenizer.escapeImages(content)
	test.AssertEqual(t, content, escapedContent)
	test.AssertEqual(t, 1, len(tokenizer.images))
	test.AssertEqual(t, "File:Foo.pdf", tokenizer.images[0])
}

func TestEscapeImages_keepSvgsEvenWhenIgnored(t *testing.T) {
	setup()
	tokenizer := NewTokenizer("foo", "bar")

	config.Current.ConvertSvgToPng = true

	var content string

	content = "File:Foo.svg"
	escapedContent := tokenizer.escapeImages(content)
	test.AssertEqual(t, content, escapedContent)
	test.AssertEqual(t, 1, len(tokenizer.images))
	test.AssertEqual(t, "File:Foo.svg", tokenizer.images[0])
}

func TestEscapeImages_removeVideoWithMultilineCaption(t *testing.T) {
	setup()
	tokenizer := NewTokenizer("foo", "bar")

	content := `file:foo.webm|this caption<br>
is<br>
important!`
	content = tokenizer.escapeImages(content)
	test.AssertEqual(t, "", content)
	test.AssertEqual(t, 0, len(tokenizer.images))

	content = `file:foo.jpg|this caption<br>
is<br>
important!`
	content = tokenizer.escapeImages(content)
	test.AssertEqual(t, `file:Foo.jpg|this caption<br>
is<br>
important!`, content)
	test.AssertEqual(t, []string{"file:Foo.jpg"}, tokenizer.images)
}

func TestEscapeImages_escapeFileNames(t *testing.T) {
	setup()
	tokenizer := NewTokenizer("foo", "bar")

	content := "file:some photo.png|with|properties"
	content = tokenizer.escapeImages(content)
	test.AssertEqual(t, "file:Some_photo.png|with|properties", content)
	test.AssertEqual(t, []string{"file:Some_photo.png"}, tokenizer.images)
}

func TestEscapeImages_leadingNonAscii(t *testing.T) {
	setup()
	tokenizer := NewTokenizer("foo", "bar")

	content := "file:öäü.png|with|properties"
	content = tokenizer.escapeImages(content)
	test.AssertEqual(t, "file:Öäü.png|with|properties", content)
	test.AssertEqual(t, []string{"file:Öäü.png"}, tokenizer.images)
}

func TestEscapeImages_leadingSpecialChar(t *testing.T) {
	setup()
	tokenizer := NewTokenizer("foo", "bar")

	content := "file:\"öäü\".png|with|properties"
	content = tokenizer.escapeImages(content)
	test.AssertEqual(t, "file:\"öäü\".png|with|properties", content)
	test.AssertEqual(t, []string{"file:\"öäü\".png"}, tokenizer.images)
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

	test.AssertMapEqual(t, map[string]Token{}, tokenizer.getTokenMap())
}

func TestParseGalleries_emptyGallery(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseGalleries(`foo
<gallery>
</gallery>
bar`)

	test.AssertEqual(t, `foo
bar`, content)

	test.AssertMapEqual(t, map[string]Token{}, tokenizer.getTokenMap())
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

	test.AssertMapEqual(t, map[string]Token{}, tokenizer.getTokenMap())
}

func TestParseImages_inlineHappyPath(t *testing.T) {
	setup()
	tokenizer := NewTokenizer("foo", "bar")
	config.Current.IgnoredMediaTypes = []string{}

	content := tokenizer.parseImages("foo [[file:image.jpg]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE_INLINE+"_0$$ bar", content)

	config.Current.ConvertPdfToPng = false
	content = tokenizer.parseImages("foo [[file:image.pdf]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE_INLINE+"_1$$ bar", content)

	config.Current.ConvertSvgToPng = false
	content = tokenizer.parseImages("foo [[file:image.svg]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE_INLINE+"_2$$ bar", content)
}

func TestParseImages_withEscaping(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseImages("blubb [[file:nice image.jpg]] bar")
	test.AssertEqual(t, "blubb $$TOKEN_"+TOKEN_IMAGE_INLINE+"_0$$ bar", content)
	test.AssertMapEqual(t, map[string]Token{
		"$$TOKEN_" + TOKEN_IMAGE_INLINE + "_0$$": InlineImageToken{
			Filename: "foo/Nice_image.jpg",
			SizeX:    -1,
			SizeY:    -1,
		},
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	config.Current.IgnoredMediaTypes = []string{"gif"}
	content = tokenizer.parseImages("foo [[file:nice image.gif]] bar")
	test.AssertEqual(t, "foo  bar", content)
	test.AssertMapEqual(t, map[string]Token{}, tokenizer.getTokenMap())
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
	test.AssertMapEqual(t, map[string]Token{
		"$$TOKEN_" + TOKEN_IMAGE + "_0$$": ImageToken{
			Filename: "foo/Image.jpg",
			Caption:  CaptionToken{},
			SizeX:    100,
			SizeY:    200,
		},
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("blubb [[file:image.jpg|x200px]] bar")
	test.AssertEqual(t, "blubb $$TOKEN_"+TOKEN_IMAGE+"_0$$ bar", content)
	test.AssertMapEqual(t, map[string]Token{
		"$$TOKEN_" + TOKEN_IMAGE + "_0$$": ImageToken{
			Filename: "foo/Image.jpg",
			Caption:  CaptionToken{},
			SizeX:    -1,
			SizeY:    200,
		},
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("blubb [[file:image.jpg|200px]] bar")
	test.AssertEqual(t, "blubb $$TOKEN_"+TOKEN_IMAGE+"_0$$ bar", content)
	test.AssertMapEqual(t, map[string]Token{
		"$$TOKEN_" + TOKEN_IMAGE + "_0$$": ImageToken{
			Filename: "foo/Image.jpg",
			Caption:  CaptionToken{},
			SizeX:    200,
			SizeY:    -1,
		},
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("blubb [[file:image.jpg|mini|200px]] bar")
	test.AssertEqual(t, "blubb $$TOKEN_"+TOKEN_IMAGE+"_0$$ bar", content)
	test.AssertMapEqual(t, map[string]Token{
		"$$TOKEN_" + TOKEN_IMAGE + "_0$$": ImageToken{
			Filename: "foo/Image.jpg",
			Caption:  CaptionToken{},
			SizeX:    200,
			SizeY:    -1,
		},
	}, tokenizer.getTokenMap())
}

func TestParseImages_withCaption(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")

	content := tokenizer.parseImages("foo [[file:image.jpg|10x20px|mini|some caption]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_0$$ bar", content)
	test.AssertMapEqual(t, map[string]Token{
		"$$TOKEN_" + TOKEN_IMAGE + "_0$$": ImageToken{
			Filename: "foo/Image.jpg",
			Caption:  CaptionToken{Content: "some caption"},
			SizeX:    10,
			SizeY:    20,
		},
	}, tokenizer.getTokenMap())
}

func TestParseImages_withCaptionEndingWithLinks(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := tokenizer.parseImages("foo [[file:image.jpg|mini|some [https://foo.com link]]] bar")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_1$$ bar", content)
	test.AssertMapEqual(t, map[string]Token{
		"$$TOKEN_" + TOKEN_IMAGE + "_1$$": ImageToken{
			Filename: "foo/Image.jpg",
			Caption:  CaptionToken{Content: "some $$TOKEN_" + TOKEN_EXTERNAL_LINK + "_0$$"},
			SizeX:    -1,
			SizeY:    -1,
		},
		"$$TOKEN_" + TOKEN_EXTERNAL_LINK + "_0$$": ExternalLinkToken{
			URL:      "https://foo.com",
			LinkText: "link",
		},
	}, tokenizer.getTokenMap())

	tokenizer = NewTokenizer("foo", "bar")
	content = tokenizer.parseImages("foo [[file:image.jpg|mini|some [[article]]]]")
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_1$$", content)
	test.AssertMapEqual(t, map[string]Token{
		"$$TOKEN_" + TOKEN_IMAGE + "_1$$": ImageToken{
			Filename: "foo/Image.jpg",
			Caption:  CaptionToken{Content: "some $$TOKEN_" + TOKEN_INTERNAL_LINK + "_0$$"},
			SizeX:    -1,
			SizeY:    -1,
		},
		"$$TOKEN_" + TOKEN_INTERNAL_LINK + "_0$$": InternalLinkToken{
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
	test.AssertEqual(t, "foo $$TOKEN_"+TOKEN_IMAGE+"_0$$ bar", content)
	test.AssertMapEqual(t, map[string]Token{
		"$$TOKEN_" + TOKEN_IMAGE + "_0$$": ImageToken{
			Filename: "foo/Image.jpg",
			Caption:  CaptionToken{Content: "some caption"},
			SizeX:    10,
			SizeY:    20,
		},
	}, tokenizer.getTokenMap())
}
