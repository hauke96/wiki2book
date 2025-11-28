package config

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"
	"wiki2book/test"

	"github.com/hauke96/sigolo/v2"
)

func testCallExpectingPanic(t *testing.T, call func()) {
	defaultValidationErrorHandler = func(err error) {
		panic(err)
	}
	defer func() {
		if r := recover(); r == nil {
			sigolo.Fatalb(2, "Expected code to panic but it didn't")
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

func TestAssertValidity_outputType(t *testing.T) {
	config := NewDefaultConfig()

	config.OutputType = "foobar"
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.OutputType = OutputTypeEpub2 + "blubb"
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.OutputType = "blubb" + OutputTypeEpub2
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.OutputType = OutputTypeEpub2
	config.AssertValidity()

	config.OutputType = OutputTypeEpub3
	config.AssertValidity()

	config.OutputDriver = OutputDriverInternal // Needed by the "stats" output types

	config.OutputType = OutputTypeStatsJson
	config.AssertValidity()

	config.OutputType = OutputTypeStatsTxt
	config.AssertValidity()
}

func TestAssertValidity_outputDriver(t *testing.T) {
	config := NewDefaultConfig()
	config.OutputType = OutputTypeEpub3

	config.OutputDriver = "foobar"
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.OutputDriver = OutputDriverInternal + "blubb"
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.OutputDriver = "blubb" + OutputDriverInternal
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.OutputDriver = OutputDriverInternal
	config.AssertValidity()

	config.OutputDriver = OutputDriverPandoc
	config.AssertValidity()
}

func TestAssertValidity_combinationOfOutputTypeAndDriver(t *testing.T) {
	config := NewDefaultConfig()

	// OutputTypeEpub2
	config.OutputType = OutputTypeEpub2

	config.OutputDriver = OutputDriverInternal
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.OutputDriver = OutputDriverPandoc
	config.AssertValidity()

	// OutputTypeEpub3
	config.OutputType = OutputTypeEpub3

	config.OutputDriver = OutputDriverInternal
	config.AssertValidity()

	config.OutputDriver = OutputDriverPandoc
	config.AssertValidity()

	// OutputTypeStatsJson
	config.OutputType = OutputTypeStatsJson

	config.OutputDriver = OutputDriverInternal
	config.AssertValidity()

	config.OutputDriver = OutputDriverPandoc
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	// OutputTypeStatsTxt
	config.OutputType = OutputTypeStatsTxt

	config.OutputDriver = OutputDriverInternal
	config.AssertValidity()

	config.OutputDriver = OutputDriverPandoc
	testCallExpectingPanic(t, func() { config.AssertValidity() })
}

func TestAssertValidity_mathConverter(t *testing.T) {
	config := NewDefaultConfig()

	config.MathConverter = "foobar"
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.MathConverter = MathConverterWikimedia + "blubb"
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.MathConverter = "blubb" + MathConverterWikimedia
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.MathConverter = MathConverterNone
	config.AssertValidity()

	config.MathConverter = MathConverterWikimedia
	config.AssertValidity()

	config.MathConverter = MathConverterTemplate
	config.AssertValidity()
}

func TestAssertValidity_tocDepth(t *testing.T) {
	config := NewDefaultConfig()

	config.TocDepth = -1
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.TocDepth = 7
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.TocDepth = 0
	config.AssertValidity()

	config.TocDepth = 1
	config.AssertValidity()

	config.TocDepth = 2
	config.AssertValidity()

	config.TocDepth = 3
	config.AssertValidity()

	config.TocDepth = 4
	config.AssertValidity()

	config.TocDepth = 5
	config.AssertValidity()

	config.TocDepth = 6
	config.AssertValidity()
}

func TestAssertValidity_workerThreads(t *testing.T) {
	config := NewDefaultConfig()

	config.WorkerThreads = -1
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.WorkerThreads = 0
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.WorkerThreads = 1
	config.AssertValidity()

	config.WorkerThreads = 10
	config.AssertValidity()

	config.WorkerThreads = 100
	config.AssertValidity()
}

func TestAssertValidity_commandTemplateSvgToPng(t *testing.T) {
	config := NewDefaultConfig()

	config.CommandTemplateSvgToPng = "foobar"
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.CommandTemplateSvgToPng = "foo" + InputPlaceholder + "bar"
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.CommandTemplateSvgToPng = "foo" + InputPlaceholder + OutputPlaceholder + "bar"
	config.AssertValidity()

	config.CommandTemplateSvgToPng = "foo" + InputPlaceholder + "blubb" + OutputPlaceholder + "bar"
	config.AssertValidity()

	config.CommandTemplateSvgToPng = InputPlaceholder + OutputPlaceholder
	config.AssertValidity()
}

func TestAssertValidity_commandTemplateMathSvgToPng(t *testing.T) {
	config := NewDefaultConfig()

	config.CommandTemplateMathSvgToPng = ""
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.CommandTemplateMathSvgToPng = "foobar"
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.CommandTemplateMathSvgToPng = "foo" + InputPlaceholder + "bar"
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.CommandTemplateMathSvgToPng = "foo" + InputPlaceholder + OutputPlaceholder + "bar"
	config.AssertValidity()

	config.CommandTemplateMathSvgToPng = "foo" + InputPlaceholder + "blubb" + OutputPlaceholder + "bar"
	config.AssertValidity()

	config.CommandTemplateMathSvgToPng = InputPlaceholder + OutputPlaceholder
	config.AssertValidity()
}

func TestAssertValidity_commandTemplateImageProcessing(t *testing.T) {
	config := NewDefaultConfig()

	config.CommandTemplateImageProcessing = "foobar"
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.CommandTemplateImageProcessing = "foo" + InputPlaceholder + "bar"
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.CommandTemplateImageProcessing = "foo" + InputPlaceholder + OutputPlaceholder + "bar"
	config.AssertValidity()

	config.CommandTemplateImageProcessing = "foo" + InputPlaceholder + "blubb" + OutputPlaceholder + "bar"
	config.AssertValidity()

	config.CommandTemplateImageProcessing = InputPlaceholder + OutputPlaceholder
	config.AssertValidity()
}

func TestAssertValidity_commandTemplatePdfToPng(t *testing.T) {
	config := NewDefaultConfig()

	config.CommandTemplatePdfToPng = "foobar"
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.CommandTemplatePdfToPng = "foo" + InputPlaceholder + "bar"
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.CommandTemplatePdfToPng = "foo" + InputPlaceholder + OutputPlaceholder + "bar"
	config.AssertValidity()

	config.CommandTemplatePdfToPng = "foo" + InputPlaceholder + "blubb" + OutputPlaceholder + "bar"
	config.AssertValidity()

	config.CommandTemplatePdfToPng = InputPlaceholder + OutputPlaceholder
	config.AssertValidity()
}

func TestAssertValidity_commandTemplateWebpToPng(t *testing.T) {
	config := NewDefaultConfig()

	config.CommandTemplateWebpToPng = "foobar"
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.CommandTemplateWebpToPng = "foo" + InputPlaceholder + "bar"
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.CommandTemplateWebpToPng = "foo" + InputPlaceholder + OutputPlaceholder + "bar"
	config.AssertValidity()

	config.CommandTemplateWebpToPng = "foo" + InputPlaceholder + "blubb" + OutputPlaceholder + "bar"
	config.AssertValidity()

	config.CommandTemplateWebpToPng = InputPlaceholder + OutputPlaceholder
	config.AssertValidity()
}

func TestAssertValidity_cacheMaxSize(t *testing.T) {
	config := NewDefaultConfig()

	config.CacheMaxSize = -1
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.CacheMaxSize = 0
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.CacheMaxSize = 1
	config.AssertValidity()

	config.CacheMaxSize = 10
	config.AssertValidity()

	config.CacheMaxSize = 100
	config.AssertValidity()
}

func TestAssertValidity_cacheMaxAge(t *testing.T) {
	config := NewDefaultConfig()

	config.CacheMaxAge = -1
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.CacheMaxAge = 0
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.CacheMaxAge = 1
	config.AssertValidity()

	config.CacheMaxAge = 10
	config.AssertValidity()

	config.CacheMaxAge = 100
	config.AssertValidity()
}

func TestAssertValidity_cacheEvictionStrategy(t *testing.T) {
	config := NewDefaultConfig()

	config.CacheEvictionStrategy = "foobar"
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.CacheEvictionStrategy = CacheEvictionStrategyLargest + "blubb"
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.CacheEvictionStrategy = "blubb" + CacheEvictionStrategyLargest
	testCallExpectingPanic(t, func() { config.AssertValidity() })

	config.CacheEvictionStrategy = CacheEvictionStrategyLargest
	config.AssertValidity()

	config.CacheEvictionStrategy = CacheEvictionStrategyLru
	config.AssertValidity()

	config.CacheEvictionStrategy = CacheEvictionStrategyNone
	config.AssertValidity()
}

// ---------- Script to generate markdown doc ----------

type configEntry struct {
	name          string
	description   string
	defaultValue  string
	allowedValues string
}

func (c *configEntry) toMarkdown() string {
	return fmt.Sprintf("| `%s` | %s | %s | %s |\n", c.name, c.description, c.defaultValue, c.allowedValues)
}

// Not a test, but generates markdown that can be pasted into the "doc/configuration.md" file.
func _TestGenerateDoc(t *testing.T) {
	contentBytes, err := os.ReadFile("config.go")
	sigolo.FatalCheck(err)

	content := string(contentBytes)

	// Collect content of Configuration struct where all the documentation is
	structContent := ""
	isWithinStruct := false
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "type Configuration struct") {
			isWithinStruct = true
			continue // Do not collect this line
		} else if isWithinStruct && strings.HasPrefix(line, "}") {
			isWithinStruct = false
		}

		if isWithinStruct {
			structContent += line + "\n"
		}
	}

	// Group into the separate fields
	var configKeys []string
	configEntries := map[string]*configEntry{}
	entryNameRegex := regexp.MustCompile(".*`json:\"(.*)\".*")
	configEntryContents := strings.Split(structContent, "/*")
	for i, configEntryContent := range configEntryContents {
		if strings.TrimSpace(configEntryContent) == "" {
			continue
		}

		entry := &configEntry{
			name:          "",
			description:   "",
			defaultValue:  "",
			allowedValues: "",
		}
		var descriptionLines []string

		entryContentLines := strings.Split(configEntryContent, "\n")
		for j := 0; j < len(entryContentLines); j++ {
			line := entryContentLines[j]
			if strings.HasPrefix(line, "Default:") {
				line = strings.TrimPrefix(line, "Default:")
				line = strings.TrimSpace(line)
				if line != "" {
					// Single line default value
					entry.defaultValue = line
				} else {
					// Multiline default value
					var defaultDocLines []string
					for ; j < len(entryContentLines); j++ {
						line = entryContentLines[j]
						if strings.Contains(line, "Allowed values:") || strings.Contains(line, "JSON example:") || strings.Contains(line, "`json:\"") || strings.Contains(line, "*/") {
							break
						}
						if line != "" {
							defaultDocLines = append(defaultDocLines, line)
						}
					}
					result := toHtml(defaultDocLines)
					entry.defaultValue = result
					j = j - 1 // To not skip the current line, which does not belong to the default values
				}
			} else if strings.HasPrefix(line, "Allowed values:") {
				line = strings.TrimPrefix(line, "Allowed values:")
				line = strings.TrimSpace(line)
				if line != "" {
					// Single line allowed value
					entry.allowedValues = line
				} else {
					// Multiline allowed value
					var allowedValueLines []string
					for ; j < len(entryContentLines); j++ {
						line = entryContentLines[j]
						if strings.Contains(line, "Default:") || strings.Contains(line, "JSON example:") || strings.Contains(line, "`json:\"") || strings.Contains(line, "*/") {
							break
						}
						if line != "" {
							allowedValueLines = append(allowedValueLines, line)
						}
					}
					result := toHtml(allowedValueLines)
					entry.allowedValues = result
					j = j - 1 // To not skip the current line, which does not belong to the allowed values
				}
			} else if strings.Contains(line, "`json:\"") {
				submatch := entryNameRegex.FindStringSubmatch(line)
				entry.name = submatch[1]
			} else if !strings.Contains(line, "*/") && line != "" {
				lastLine := ""
				if len(descriptionLines) != 0 {
					lastLine = descriptionLines[len(descriptionLines)-1]
				}
				isContinuingTextLine := (len(descriptionLines) != 0 && lastLine != "") && !strings.HasSuffix(line, ":") && !strings.Contains(line, "JSON example:") && !strings.Contains(line, "ul>") && !strings.Contains(line, "li>")
				if isContinuingTextLine {
					descriptionLines[len(descriptionLines)-1] = descriptionLines[len(descriptionLines)-1] + " " + line
				} else {
					descriptionLines = append(descriptionLines, line)
				}
			}
		}

		var cleanedDescriptionLines []string
		for _, line := range descriptionLines {
			if !strings.Contains(line, "*/") && strings.TrimSpace(line) != "" {
				cleanedDescriptionLines = append(cleanedDescriptionLines, line)
			}
		}

		entry.description = toHtml(cleanedDescriptionLines)

		if entry.name == "" {
			sigolo.Fatalf("Entry name must not be empty. Entry index %d with content:\n%s", i, configEntryContent)
		}

		configKeys = append(configKeys, entry.name)
		configEntries[entry.name] = entry
		sigolo.Debugf("Found entry: %+v", entry)
	}

	// Sort and generate markdown
	slices.Sort(configKeys)
	resultMarkdown := "| Name | Description | Default | Allowed values |\n|---|---|---|---|\n"
	for _, key := range configKeys {
		resultMarkdown += configEntries[key].toMarkdown()
	}

	sigolo.Infof("Markdown:\n\n%s", resultMarkdown)
}

func toHtml(lines []string) string {
	result := strings.Join(lines, "</br>")
	result = strings.ReplaceAll(result, "ul></br>", "ul>")
	result = strings.ReplaceAll(result, "</br><ul>", "<ul>")
	result = strings.ReplaceAll(result, "</br><li>", "<li>")
	result = strings.ReplaceAll(result, "li></br>", "li>")
	result = strings.ReplaceAll(result, "  ", " ")
	return result
}
