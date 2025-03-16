package config

import (
	"reflect"
	"testing"
	"wiki2book/test"
)

func TestMergeIntoCurrentConfig(t *testing.T) {
	// Set current and default config to an empty config so that all fields will be overwritten by the merge function.
	Current = &Configuration{}
	defaultConfig = &Configuration{}
	expectedConfig := &Configuration{
		ForceRegenerateHtml:            true,
		SvgSizeToViewbox:               true,
		OutputType:                     OutputTypeEpub3,
		OutputDriver:                   OutputDriverInternal,
		CacheDir:                       "/cache-dir",
		StyleFile:                      "/style-file",
		CoverImage:                     "/cover-image",
		CommandTemplateSvgToPng:        "command-template-svg-to-png" + InputPlaceholder + OutputPlaceholder,
		CommandTemplateMathSvgToPng:    "command-template-math-svg-to-png" + InputPlaceholder + OutputPlaceholder,
		CommandTemplateImageProcessing: "command-template-image-processing" + InputPlaceholder + OutputPlaceholder,
		CommandTemplatePdfToPng:        "command-template-pdf-to-png" + InputPlaceholder + OutputPlaceholder,
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
		WikipediaImageArticleInstances: []string{"wikipedia-image-article-instances"},
		FilePrefixe:                    []string{"file-prefixe"},
		AllowedLinkPrefixes:            []string{"allowed-link-prefixe"},
		CategoryPrefixes:               []string{"category-prefixes"},
		MathConverter:                  MathConverterWikimedia,
		TocDepth:                       3,
		WorkerThreads:                  234,
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
