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
