package config

import (
	"reflect"
	"testing"
	"wiki2book/test"
)

func testCallExpectingPanic(t *testing.T, call func()) {
	defaultValidationErrorHandler = func(err error) {
		panic(err)
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected code to panic but it didn't")
		}
	}()

	call()
}

func TestMergeIntoCurrentConfig(t *testing.T) {
	Current = NewDefaultConfig()
	expectedConfig := &Configuration{
		ForceRegenerateHtml:            true,
		SvgSizeToViewbox:               true,
		OutputType:                     OutputTypeEpub3,
		OutputDriver:                   OutputDriverInternal,
		CacheDir:                       "/cache-dir",
		CacheMaxSize:                   123,
		CacheMaxAge:                    234,
		CacheEvictionStrategy:          CacheEvictionStrategyNone,
		StyleFile:                      "/style-file",
		CoverImage:                     "/cover-image",
		CommandTemplateSvgToPng:        "command-template-svg-to-png" + InputPlaceholder + OutputPlaceholder,
		CommandTemplateMathSvgToPng:    "command-template-math-svg-to-png" + InputPlaceholder + OutputPlaceholder,
		CommandTemplateImageProcessing: "command-template-image-processing" + InputPlaceholder + OutputPlaceholder,
		CommandTemplatePdfToPng:        "command-template-pdf-to-png" + InputPlaceholder + OutputPlaceholder,
		CommandTemplateWebpToPng:       "command-template-webp-to-png" + InputPlaceholder + OutputPlaceholder,
		PandocExecutable:               "pandoc-executable",
		PandocDataDir:                  "/pandoc-data-dir",
		FontFiles:                      []string{"font-files"},
		ConvertPdfToPng:                true,
		ConvertSvgToPng:                true,
		IgnoredTemplates:               []string{"ignored-templates"},
		TrailingTemplates:              []string{"trailing-templates"},
		IgnoredImageParams:             []string{"ignored-image-params"},
		IgnoredMediaTypes:              []string{"ignored-media-types"},
		WikipediaInstance:              "wikipedia-instance",
		WikipediaHost:                  "wikipedia-host",
		WikipediaImageHost:             "wikipedia-image-host",
		WikipediaMathRestApi:           "wikipedia-math-rest-api",
		WikipediaImageArticleHosts:     []string{"wikipedia-image-article-hosts"},
		FilePrefixe:                    []string{"file-prefixe"},
		AllowedLinkPrefixes:            []string{"allowed-link-prefixe"},
		CategoryPrefixes:               []string{"category-prefixes"},
		MathConverter:                  MathConverterWikimedia,
		TocDepth:                       3,
		WorkerThreads:                  234,
		UserAgentTemplate:              "user-agent-template",
	}

	MergeIntoCurrentConfig(expectedConfig)

	vActual := reflect.ValueOf(*Current)
	vExpected := reflect.ValueOf(*expectedConfig)
	actualValues := map[string]string{}
	expectedValues := map[string]string{}
	for i := 0; i < vActual.NumField(); i++ {
		actualValues[vActual.Type().Field(i).Name] = vActual.Field(i).String()
		expectedValues[vExpected.Type().Field(i).Name] = vExpected.Field(i).String()
	}
	test.AssertMapEqual(t, expectedValues, actualValues)
}

func TestMergeIntoCurrentConfig_validEmptyValues(t *testing.T) {
	Current = NewDefaultConfig()
	expectedConfig := NewDefaultConfig()
	expectedConfig.CommandTemplateImageProcessing = ""
	expectedConfig.CommandTemplateSvgToPng = ""
	expectedConfig.CommandTemplatePdfToPng = ""
	expectedConfig.CommandTemplateWebpToPng = ""
	expectedConfig.FontFiles = []string{}
	expectedConfig.IgnoredTemplates = []string{}
	expectedConfig.TrailingTemplates = []string{}
	expectedConfig.IgnoredImageParams = []string{}
	expectedConfig.IgnoredMediaTypes = []string{}
	expectedConfig.WikipediaImageArticleHosts = []string{}
	expectedConfig.FilePrefixe = []string{}
	expectedConfig.AllowedLinkPrefixes = []string{}
	expectedConfig.CategoryPrefixes = []string{}

	MergeIntoCurrentConfig(expectedConfig)

	test.AssertEqual(t, "", Current.CommandTemplateImageProcessing)
	test.AssertEqual(t, "", Current.CommandTemplateWebpToPng)
	test.AssertEqual(t, []string{}, Current.IgnoredTemplates)
	test.AssertEqual(t, []string{}, Current.TrailingTemplates)
	test.AssertEqual(t, []string{}, Current.IgnoredImageParams)
	test.AssertEqual(t, []string{}, Current.IgnoredMediaTypes)
	test.AssertEqual(t, []string{}, Current.WikipediaImageArticleHosts)
	test.AssertEqual(t, []string{}, Current.FilePrefixe)
	test.AssertEqual(t, []string{}, Current.AllowedLinkPrefixes)
	test.AssertEqual(t, []string{}, Current.CategoryPrefixes)
}

func TestMergeIntoCurrentConfig_invalidMathConverter(t *testing.T) {
	// Arrange
	Current = NewDefaultConfig()
	expectedConfig := NewDefaultConfig()

	expectedConfig.MathConverter = "foobar"
	testCallExpectingPanic(t, func() { MergeIntoCurrentConfig(expectedConfig) })

	expectedConfig.MathConverter = "pandoc"
	testCallExpectingPanic(t, func() { MergeIntoCurrentConfig(expectedConfig) })

	expectedConfig.MathConverter = "internal"
	testCallExpectingPanic(t, func() { MergeIntoCurrentConfig(expectedConfig) })
}

func TestMakePathsAbsolute(t *testing.T) {
	actualConfig := &Configuration{
		CacheDir:      "cache-dir",
		StyleFile:     "style-file",
		CoverImage:    "cover-image",
		PandocDataDir: "pandoc-data-dir",
		FontFiles:     []string{"fontA", "fontB"},
	}

	actualConfig.makePathsAbsolute("/foo/file")

	expectedConfig := &Configuration{
		CacheDir:      "/foo/cache-dir",
		StyleFile:     "/foo/style-file",
		CoverImage:    "/foo/cover-image",
		PandocDataDir: "/foo/pandoc-data-dir",
		FontFiles:     []string{"/foo/fontA", "/foo/fontB"},
	}
	test.AssertEqual(t, *expectedConfig, *actualConfig)
}
