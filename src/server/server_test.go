package server

import (
	"testing"
	"wiki2book/cache"
	"wiki2book/config"
	"wiki2book/test"
)

func TestResetNonUploadableConfigProperties(t *testing.T) {
	// Arrange
	defaultConfig := config.NewDefaultConfig()
	configService := config.NewConfigServiceForConfig(defaultConfig)
	server := &Server{
		configService: configService,
		fileCache:     cache.NewCache(configService),
	}

	// Use two configs: One with default values and one with other values for each property. They will be merged so that
	// any missing attributes in the reset method will lead to a failing test.
	uploadedConfig := &config.Configuration{
		OutputType:                     config.OutputTypeEpub3,
		OutputDriver:                   config.OutputDriverInternal,
		CacheDir:                       "/test-cache",
		CacheMaxSize:                   123,
		CacheMaxAge:                    234,
		CacheEvictionStrategy:          config.CacheEvictionStrategyNone,
		StyleFile:                      "/test-style.css",
		IgnoredTemplates:               []string{"ignored-template"},
		TrailingTemplates:              []string{"trailing-tamplate"},
		IgnoredImageParams:             []string{"ignored-image-params"},
		IgnoredMediaTypes:              []string{"foo"},
		WikipediaInstance:              "de",
		WikipediaHost:                  "wikipedia.host.com",
		WikipediaImageHost:             "wikipedia.image-host.com",
		WikipediaImageArticleHosts:     []string{"wikipedia.image-article-host.com"},
		WikipediaMathRestApi:           "https://wikipedia.math.com/api/rest_v1/media/math",
		FilePrefixes:                   []string{"file-prefixes"},
		AllowedLinkPrefixes:            []string{"allowed-link-prefixes"},
		CategoryPrefixes:               []string{"category-prefixes"},
		MathConverter:                  "wikimedia",
		CommandTemplateSvgToPng:        "command-template-svg-to-png",
		CommandTemplateMathSvgToPng:    "command-template-math-to-png",
		CommandTemplateImageProcessing: "command-template-image-processing",
		CommandTemplatePdfToPng:        "command-template-pdf-to-png",
		CommandTemplateWebpToPng:       "command-template-webp-to-png",
		PandocExecutable:               "pandoc-executable",
		PandocDataDir:                  "/pandoc/data/dir",
		FontFiles:                      []string{"font-files"},
		TocDepth:                       345,
		WorkerThreads:                  456,
		UserAgentTemplate:              "user agent {{VERSION}} template",
		ServerPort:                     5678,
	}

	config := config.NewDefaultConfig()
	config.MergeNonDefaultValues(uploadedConfig)

	// Act
	server.resetNonUploadableConfigProperties(config)

	// Assert
	test.AssertEqual(t, defaultConfig.ForceRegenerateHtml, config.ForceRegenerateHtml)
	test.AssertEqual(t, uploadedConfig.SvgSizeToViewbox, config.SvgSizeToViewbox)
	test.AssertEqual(t, uploadedConfig.OutputType, config.OutputType)
	test.AssertEqual(t, defaultConfig.OutputDriver, config.OutputDriver)
	test.AssertEqual(t, defaultConfig.CacheDir, config.CacheDir)
	test.AssertEqual(t, defaultConfig.CacheMaxSize, config.CacheMaxSize)
	test.AssertEqual(t, defaultConfig.CacheMaxAge, config.CacheMaxAge)
	test.AssertEqual(t, defaultConfig.CacheEvictionStrategy, config.CacheEvictionStrategy)
	test.AssertEqual(t, defaultConfig.StyleFile, config.StyleFile)
	test.AssertEqual(t, defaultConfig.CoverImage, config.CoverImage)
	test.AssertEqual(t, defaultConfig.CommandTemplateSvgToPng, config.CommandTemplateSvgToPng)
	test.AssertEqual(t, defaultConfig.CommandTemplateMathSvgToPng, config.CommandTemplateMathSvgToPng)
	test.AssertEqual(t, defaultConfig.CommandTemplateImageProcessing, config.CommandTemplateImageProcessing)
	test.AssertEqual(t, defaultConfig.CommandTemplatePdfToPng, config.CommandTemplatePdfToPng)
	test.AssertEqual(t, defaultConfig.CommandTemplateWebpToPng, config.CommandTemplateWebpToPng)
	test.AssertEqual(t, defaultConfig.PandocExecutable, config.PandocExecutable)
	test.AssertEqual(t, defaultConfig.PandocDataDir, config.PandocDataDir)
	test.AssertEqual(t, defaultConfig.FontFiles, config.FontFiles)
	test.AssertEqual(t, uploadedConfig.IgnoredTemplates, config.IgnoredTemplates)
	test.AssertEqual(t, uploadedConfig.TrailingTemplates, config.TrailingTemplates)
	test.AssertEqual(t, uploadedConfig.IgnoredImageParams, config.IgnoredImageParams)
	test.AssertEqual(t, uploadedConfig.IgnoredMediaTypes, config.IgnoredMediaTypes)
	test.AssertEqual(t, uploadedConfig.WikipediaInstance, config.WikipediaInstance)
	test.AssertEqual(t, uploadedConfig.WikipediaHost, config.WikipediaHost)
	test.AssertEqual(t, uploadedConfig.WikipediaImageHost, config.WikipediaImageHost)
	test.AssertEqual(t, uploadedConfig.WikipediaMathRestApi, config.WikipediaMathRestApi)
	test.AssertEqual(t, uploadedConfig.WikipediaImageArticleHosts, config.WikipediaImageArticleHosts)
	test.AssertEqual(t, uploadedConfig.FilePrefixes, config.FilePrefixes)
	test.AssertEqual(t, uploadedConfig.AllowedLinkPrefixes, config.AllowedLinkPrefixes)
	test.AssertEqual(t, uploadedConfig.CategoryPrefixes, config.CategoryPrefixes)
	test.AssertEqual(t, defaultConfig.MathConverter, config.MathConverter)
	test.AssertEqual(t, uploadedConfig.TocDepth, config.TocDepth)
	test.AssertEqual(t, defaultConfig.WorkerThreads, config.WorkerThreads)
	test.AssertEqual(t, defaultConfig.UserAgentTemplate, config.UserAgentTemplate)
}
