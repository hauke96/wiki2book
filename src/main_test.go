package main

import (
	"os"
	"reflect"
	"testing"
	"wiki2book/config"
	"wiki2book/test"
)

func TestCliArgs(t *testing.T) {
	os.Args = []string{
		"", "test",
		"--force-regenerate-html", "force-regenerate-html",
		"--svg-size-to-viewbox", "svg-size-to-viewbox",
		"--output-type", "output-type",
		"--output-driver", "output-driver",
		"--cache-dir", "cache-dir",
		"--cache-max-size", "123",
		"--cache-max-age", "234",
		"--cache-eviction-strategy", "cache-eviction-strategy",
		"--style-file", "style-file",
		"--cover-image", "cover-image",
		"--command-template-svg-to-png", "command-template-svg-to-png",
		"--command-template-math-svg-to-png", "command-template-math-svg-to-png",
		"--command-template-image-processing", "command-template-image-processing",
		"--command-template-pdf-to-png", "command-template-pdf-to-png",
		"--command-template-webp-to-png", "command-template-webp-to-png",
		"--pandoc-executable", "pandoc-executable",
		"--pandoc-data-dir", "pandoc-data-dir",
		"--font-files", "font-files",
		"--ignored-templates", "ignored-templates",
		"--trailing-templates", "trailing-templates",
		"--ignored-image-params", "ignored-image-params",
		"--ignored-media-types", "ignored-media-types",
		"--wikipedia-instance", "wikipedia-instance",
		"--wikipedia-host", "wikipedia-host",
		"--wikipedia-image-host", "wikipedia-image-host",
		"--wikipedia-math-rest-api", "wikipedia-math-rest-api",
		"--wikipedia-image-article-hosts", "wikipedia-image-article-hosts",
		"--file-prefixes", "file-prefixes",
		"--allowed-link-prefixes", "allowed-link-prefixes",
		"--category-prefixes", "category-prefixes",
		"--math-converter", "math-converter",
		"--toc-depth", "123",
		"--worker-threads", "234",
		"--user-agent-template", "user-agent-template",
	}
	testCmd := getCommand("test", "", 1)
	cliConfig = &config.Configuration{}
	rootCmd := initCli()
	rootCmd.AddCommand(testCmd)

	err := rootCmd.Execute()

	test.AssertNil(t, err)

	// *2 because each cli flag also has a value
	// -2 because server-port is a cli flag specific to the "server" command and thus not tested in this test.expectedNumberOfArgs := reflect.ValueOf(*cliConfig).NumField()*2 - 2
	expectedNumberOfArgs := reflect.ValueOf(*cliConfig).NumField()*2 - 2
	// -2 because the first two args are: 1) filename and 2) command. Both should be ignored, because the test just compares actual parameters.
	actualNumberOfArgs := len(os.Args) - 2
	// I expect each configuration entry represented in the cli arguments
	test.AssertEqual(t, expectedNumberOfArgs, actualNumberOfArgs)

	test.AssertTrue(t, cliConfig.ForceRegenerateHtml)
	test.AssertTrue(t, cliConfig.SvgSizeToViewbox)
	test.AssertEqual(t, "output-type", cliConfig.OutputType)
	test.AssertEqual(t, "output-driver", cliConfig.OutputDriver)
	test.AssertEqual(t, "cache-dir", cliConfig.CacheDir)
	test.AssertEqual(t, 123, cliConfig.CacheMaxSize)
	test.AssertEqual(t, 234, cliConfig.CacheMaxAge)
	test.AssertEqual(t, "cache-eviction-strategy", cliConfig.CacheEvictionStrategy)
	test.AssertEqual(t, "style-file", cliConfig.StyleFile)
	test.AssertEqual(t, "cover-image", cliConfig.CoverImage)
	test.AssertEqual(t, "command-template-svg-to-png", cliConfig.CommandTemplateSvgToPng)
	test.AssertEqual(t, "command-template-math-svg-to-png", cliConfig.CommandTemplateMathSvgToPng)
	test.AssertEqual(t, "command-template-image-processing", cliConfig.CommandTemplateImageProcessing)
	test.AssertEqual(t, "command-template-pdf-to-png", cliConfig.CommandTemplatePdfToPng)
	test.AssertEqual(t, "command-template-webp-to-png", cliConfig.CommandTemplateWebpToPng)
	test.AssertEqual(t, "pandoc-executable", cliConfig.PandocExecutable)
	test.AssertEqual(t, "pandoc-data-dir", cliConfig.PandocDataDir)
	test.AssertEqual(t, []string{"font-files"}, cliConfig.FontFiles)
	test.AssertEqual(t, []string{"ignored-templates"}, cliConfig.IgnoredTemplates)
	test.AssertEqual(t, []string{"trailing-templates"}, cliConfig.TrailingTemplates)
	test.AssertEqual(t, []string{"ignored-image-params"}, cliConfig.IgnoredImageParams)
	test.AssertEqual(t, []string{"ignored-media-types"}, cliConfig.IgnoredMediaTypes)
	test.AssertEqual(t, "wikipedia-instance", cliConfig.WikipediaInstance)
	test.AssertEqual(t, "wikipedia-host", cliConfig.WikipediaHost)
	test.AssertEqual(t, "wikipedia-image-host", cliConfig.WikipediaImageHost)
	test.AssertEqual(t, "wikipedia-math-rest-api", cliConfig.WikipediaMathRestApi)
	test.AssertEqual(t, []string{"wikipedia-image-article-hosts"}, cliConfig.WikipediaImageArticleHosts)
	test.AssertEqual(t, []string{"file-prefixes"}, cliConfig.FilePrefixes)
	test.AssertEqual(t, []string{"allowed-link-prefixes"}, cliConfig.AllowedLinkPrefixes)
	test.AssertEqual(t, []string{"category-prefixes"}, cliConfig.CategoryPrefixes)
	test.AssertEqual(t, "math-converter", cliConfig.MathConverter)
	test.AssertEqual(t, 123, cliConfig.TocDepth)
	test.AssertEqual(t, 234, cliConfig.WorkerThreads)
	test.AssertEqual(t, "user-agent-template", cliConfig.UserAgentTemplate)
}
