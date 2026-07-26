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

func TestResetNonUploadableProjectProperties(t *testing.T) {
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

	mergedConfig := config.NewDefaultConfig()
	mergedConfig.MergeNonDefaultValues(uploadedConfig)

	uploadedProject := &config.Project{
		Configuration: *mergedConfig,
		Metadata: config.Metadata{
			Title:    "Title",
			Language: "Language",
			Author:   "Author",
			License:  "License",
			Date:     "Date",
		},
		OutputFile: "./some/output/path.epub",
		Articles:   []string{"Foo", "Bar"},
	}

	// Act
	server.resetNonUploadableProjectProperties(uploadedProject)

	// Assert
	test.AssertEqual(t, "Title", uploadedProject.Metadata.Title)
	test.AssertEqual(t, "Language", uploadedProject.Metadata.Language)
	test.AssertEqual(t, "Author", uploadedProject.Metadata.Author)
	test.AssertEqual(t, "License", uploadedProject.Metadata.License)
	test.AssertEqual(t, "Date", uploadedProject.Metadata.Date)
	test.AssertEqual(t, "", uploadedProject.OutputFile)
	test.AssertEqual(t, []string{"Foo", "Bar"}, uploadedProject.Articles)

	test.AssertEqual(t, defaultConfig.ForceRegenerateHtml, uploadedProject.Configuration.ForceRegenerateHtml)
	test.AssertEqual(t, uploadedConfig.SvgSizeToViewbox, uploadedProject.Configuration.SvgSizeToViewbox)
	test.AssertEqual(t, uploadedConfig.OutputType, uploadedProject.Configuration.OutputType)
	test.AssertEqual(t, defaultConfig.OutputDriver, uploadedProject.Configuration.OutputDriver)
	test.AssertEqual(t, defaultConfig.CacheDir, uploadedProject.Configuration.CacheDir)
	test.AssertEqual(t, defaultConfig.CacheMaxSize, uploadedProject.Configuration.CacheMaxSize)
	test.AssertEqual(t, defaultConfig.CacheMaxAge, uploadedProject.Configuration.CacheMaxAge)
	test.AssertEqual(t, defaultConfig.CacheEvictionStrategy, uploadedProject.Configuration.CacheEvictionStrategy)
	test.AssertEqual(t, defaultConfig.StyleFile, uploadedProject.Configuration.StyleFile)
	test.AssertEqual(t, defaultConfig.CoverImage, uploadedProject.Configuration.CoverImage)
	test.AssertEqual(t, defaultConfig.CommandTemplateSvgToPng, uploadedProject.Configuration.CommandTemplateSvgToPng)
	test.AssertEqual(t, defaultConfig.CommandTemplateMathSvgToPng, uploadedProject.Configuration.CommandTemplateMathSvgToPng)
	test.AssertEqual(t, defaultConfig.CommandTemplateImageProcessing, uploadedProject.Configuration.CommandTemplateImageProcessing)
	test.AssertEqual(t, defaultConfig.CommandTemplatePdfToPng, uploadedProject.Configuration.CommandTemplatePdfToPng)
	test.AssertEqual(t, defaultConfig.CommandTemplateWebpToPng, uploadedProject.Configuration.CommandTemplateWebpToPng)
	test.AssertEqual(t, defaultConfig.PandocExecutable, uploadedProject.Configuration.PandocExecutable)
	test.AssertEqual(t, defaultConfig.PandocDataDir, uploadedProject.Configuration.PandocDataDir)
	test.AssertEqual(t, defaultConfig.FontFiles, uploadedProject.Configuration.FontFiles)
	test.AssertEqual(t, uploadedConfig.IgnoredTemplates, uploadedProject.Configuration.IgnoredTemplates)
	test.AssertEqual(t, uploadedConfig.TrailingTemplates, uploadedProject.Configuration.TrailingTemplates)
	test.AssertEqual(t, uploadedConfig.IgnoredImageParams, uploadedProject.Configuration.IgnoredImageParams)
	test.AssertEqual(t, uploadedConfig.IgnoredMediaTypes, uploadedProject.Configuration.IgnoredMediaTypes)
	test.AssertEqual(t, uploadedConfig.WikipediaInstance, uploadedProject.Configuration.WikipediaInstance)
	test.AssertEqual(t, uploadedConfig.WikipediaHost, uploadedProject.Configuration.WikipediaHost)
	test.AssertEqual(t, uploadedConfig.WikipediaImageHost, uploadedProject.Configuration.WikipediaImageHost)
	test.AssertEqual(t, uploadedConfig.WikipediaMathRestApi, uploadedProject.Configuration.WikipediaMathRestApi)
	test.AssertEqual(t, uploadedConfig.WikipediaImageArticleHosts, uploadedProject.Configuration.WikipediaImageArticleHosts)
	test.AssertEqual(t, uploadedConfig.FilePrefixes, uploadedProject.Configuration.FilePrefixes)
	test.AssertEqual(t, uploadedConfig.AllowedLinkPrefixes, uploadedProject.Configuration.AllowedLinkPrefixes)
	test.AssertEqual(t, uploadedConfig.CategoryPrefixes, uploadedProject.Configuration.CategoryPrefixes)
	test.AssertEqual(t, defaultConfig.MathConverter, uploadedProject.Configuration.MathConverter)
	test.AssertEqual(t, uploadedConfig.TocDepth, uploadedProject.Configuration.TocDepth)
	test.AssertEqual(t, defaultConfig.WorkerThreads, uploadedProject.Configuration.WorkerThreads)
	test.AssertEqual(t, defaultConfig.UserAgentTemplate, uploadedProject.Configuration.UserAgentTemplate)
}
