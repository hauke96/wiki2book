package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"wiki2book/config"
	"wiki2book/parser"
	"wiki2book/test"
	"wiki2book/util"
)

func setupCache() *util.MockFile {
	mockFile := util.NewMockFile("mock file")

	fsMock := util.NewDefaultMockFilesystem()
	fsMock.StatFunc = func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	fsMock.CreateTempFunc = func(dir, pattern string) (util.FileLike, error) { return mockFile, nil }
	util.CurrentFilesystem = fsMock
	config.Current.CacheDir = test.TestCacheFolder

	return mockFile
}

func getAndAssertStats(t *testing.T, err error, articleName string, statsOutputFilename string, mockFile *util.MockFile) *articleStats {
	test.AssertNil(t, err)
	test.AssertEqual(t, config.Current.CacheDir+"/stats/"+articleName+".json", statsOutputFilename)

	stats := &articleStats{}
	err = json.Unmarshal(mockFile.WrittenBytes, stats)
	test.AssertNil(t, err)

	return stats
}

func TestGenerate_plainText(t *testing.T) {
	// Arrange
	tokenMap := map[string]parser.Token{}

	article := &parser.Article{
		Title:    "Foobar",
		Content:  "Foó bar\nblübb.",
		TokenMap: tokenMap,
		Images:   []string{},
	}

	mockFile := setupCache()

	statsGenerator := NewStatsGenerator(tokenMap)

	// Act
	statsOutputFilename, err := statsGenerator.Generate(article)

	// Assert
	stats := getAndAssertStats(t, err, article.Title, statsOutputFilename, mockFile)

	test.AssertEqual(t, 14, stats.NumberOfCharacters)
}

func TestGenerate_plainText_countWordsCorrectly(t *testing.T) {
	// Arrange
	tokenMap := map[string]parser.Token{}

	article := &parser.Article{
		Title:    "Foobar",
		Content:  "",
		TokenMap: tokenMap,
		Images:   []string{},
	}

	// Act & Assert
	statsGenerator := NewStatsGenerator(tokenMap)
	mockFile := setupCache()
	statsOutputFilename, err := statsGenerator.Generate(article)
	stats := getAndAssertStats(t, err, article.Title, statsOutputFilename, mockFile)
	test.AssertEqual(t, 0, stats.NumberOfWords)

	// Act & Assert - Spaces
	statsGenerator = NewStatsGenerator(tokenMap)
	mockFile = setupCache()
	article.Content = "Some simple\rtest\nwith\tnormal \rwords \nrand \tonly \r\n\tspaces."
	statsOutputFilename, err = statsGenerator.Generate(article)
	stats = getAndAssertStats(t, err, article.Title, statsOutputFilename, mockFile)
	test.AssertEqual(t, 9, stats.NumberOfWords)

	// Act & Assert - Single letters
	statsGenerator = NewStatsGenerator(tokenMap)
	mockFile = setupCache()
	article.Content = "a b c ö ä ü ß µ ø"
	statsOutputFilename, err = statsGenerator.Generate(article)
	stats = getAndAssertStats(t, err, article.Title, statsOutputFilename, mockFile)
	test.AssertEqual(t, 9, stats.NumberOfWords)

	// Act & Assert - Numbers
	statsGenerator = NewStatsGenerator(tokenMap)
	mockFile = setupCache()
	article.Content = "a 1 b 2 c 3 a1 b2 c3 1a 2b 3c 1a1 2b2 3c3 a1a b2b c3c"
	statsOutputFilename, err = statsGenerator.Generate(article)
	stats = getAndAssertStats(t, err, article.Title, statsOutputFilename, mockFile)
	test.AssertEqual(t, 18, stats.NumberOfWords)

	// Act & Assert - Special characters
	statsGenerator = NewStatsGenerator(tokenMap)
	mockFile = setupCache()
	article.Content = "{{This|is}} [a]] \"test\"\n(with some separate) wórds and/or späcial-characterß."
	statsOutputFilename, err = statsGenerator.Generate(article)
	stats = getAndAssertStats(t, err, article.Title, statsOutputFilename, mockFile)
	test.AssertEqual(t, 12, stats.NumberOfWords)
}

func TestGenerate_headings(t *testing.T) {
	// Arrange
	tokenKey := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_HEADING, 3)
	token := parser.HeadingToken{
		Content: "Some heading",
		Depth:   3,
	}
	tokenMap := map[string]parser.Token{
		tokenKey: token,
	}

	article := &parser.Article{
		Title:    "Foobar",
		Content:  "Abc\n" + tokenKey + "\ndef.",
		TokenMap: tokenMap,
		Images:   []string{},
	}

	statsGenerator := NewStatsGenerator(tokenMap)

	mockFile := setupCache()

	// Act
	statsOutputFilename, err := statsGenerator.Generate(article)

	// Assert
	stats := getAndAssertStats(t, err, article.Title, statsOutputFilename, mockFile)

	test.AssertEqual(t, 21, stats.NumberOfCharacters)
}

func TestGenerate_countRefDefinitionsCorrectly(t *testing.T) {
	// Arrange
	tokenKeyRefA := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_REF_DEF, 0)
	tokenKeyRefB := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_REF_DEF, 1)
	tokenMap := map[string]parser.Token{
		tokenKeyRefA: parser.RefDefinitionToken{
			Token:   tokenKeyRefA,
			Index:   1,
			Content: "ref def A",
		},
		tokenKeyRefB: parser.RefDefinitionToken{
			Token:   tokenKeyRefB,
			Index:   2,
			Content: "ref def B",
		},
	}

	article := &parser.Article{
		Title:    "Foobar",
		Content:  "With" + tokenKeyRefA + " some refs" + tokenKeyRefB + ".",
		TokenMap: tokenMap,
		Images:   []string{},
	}

	statsGenerator := NewStatsGenerator(tokenMap)

	mockFile := setupCache()

	// Act
	statsOutputFilename, err := statsGenerator.Generate(article)

	// Assert
	stats := getAndAssertStats(t, err, article.Title, statsOutputFilename, mockFile)

	test.AssertEqual(t, 2, stats.NumberOfRefDefinitions)
}

func TestGenerate_countRefUsagesCorrectly(t *testing.T) {
	// Arrange
	tokenKeyRefA := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_REF_USAGE, 0)
	tokenKeyRefB := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_REF_USAGE, 1)
	tokenMap := map[string]parser.Token{
		tokenKeyRefA: parser.RefUsageToken{
			Token: tokenKeyRefA,
			Index: 1,
		},
		tokenKeyRefB: parser.RefUsageToken{
			Token: tokenKeyRefB,
			Index: 2,
		},
	}

	article := &parser.Article{
		Title:    "Foobar",
		Content:  "With" + tokenKeyRefA + " some refs" + tokenKeyRefB + ".",
		TokenMap: tokenMap,
		Images:   []string{},
	}

	statsGenerator := NewStatsGenerator(tokenMap)

	mockFile := setupCache()

	// Act
	statsOutputFilename, err := statsGenerator.Generate(article)

	// Assert
	stats := getAndAssertStats(t, err, article.Title, statsOutputFilename, mockFile)

	test.AssertEqual(t, 2, stats.NumberOfRefUsages)
}

func TestGenerate_countMathUsagesCorrectly(t *testing.T) {
	// Arrange
	tokenKeyRefA := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_MATH, 0)
	tokenKeyRefB := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_MATH, 1)
	tokenMap := map[string]parser.Token{
		tokenKeyRefA: parser.MathToken{
			Token:   tokenKeyRefA,
			Content: "a=1",
		},
		tokenKeyRefB: parser.MathToken{
			Token:   tokenKeyRefB,
			Content: "b=2",
		},
	}

	article := &parser.Article{
		Title:    "Foobar",
		Content:  "With" + tokenKeyRefA + " some refs" + tokenKeyRefB + ".",
		TokenMap: tokenMap,
		Images:   []string{},
	}

	statsGenerator := NewStatsGenerator(tokenMap)

	mockFile := setupCache()

	// Act
	statsOutputFilename, err := statsGenerator.Generate(article)

	// Assert
	stats := getAndAssertStats(t, err, article.Title, statsOutputFilename, mockFile)

	test.AssertEqual(t, 2, stats.NumberOfMath)
}

func TestGenerate_table(t *testing.T) {
	// Arrange
	tokenCaptionInternalLink := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_INTERNAL_LINK, 0)
	tokenCaptionExternalLink := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_EXTERNAL_LINK, 1)
	tokenImage := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_IMAGE, 2)
	tokenTable := fmt.Sprintf(parser.TOKEN_TEMPLATE, parser.TOKEN_TABLE, 3)
	tokenMap := map[string]parser.Token{
		tokenCaptionInternalLink: parser.InternalLinkToken{
			Token:       tokenCaptionInternalLink,
			ArticleName: "Foobar",
			LinkText:    "internal-link",
		},
		tokenCaptionExternalLink: parser.ExternalLinkToken{
			Token:    tokenCaptionExternalLink,
			URL:      "https://foo.com",
			LinkText: "external-link",
		},
		tokenImage: parser.ImageToken{
			Filename: "image.jpg",
			Caption:  parser.CaptionToken{Content: "some " + parser.MARKER_BOLD_OPEN + "caption" + parser.MARKER_BOLD_CLOSE},
			SizeX:    10,
			SizeY:    20,
		},
		tokenTable: parser.TableToken{
			Caption: parser.TableCaptionToken{
				Content: "caption with " + tokenCaptionInternalLink + " and " + tokenCaptionExternalLink + ".",
			},
			Rows: []parser.TableRowToken{
				{
					Columns: []parser.TableColToken{
						{
							Attributes: parser.TableColAttributeToken{},
							Content:    "b" + parser.MARKER_BOLD_OPEN + "a" + parser.MARKER_BOLD_CLOSE + "r",
							IsHeading:  false,
						},
						{
							Attributes: parser.TableColAttributeToken{},
							Content:    "some image " + tokenImage + " with caption",
							IsHeading:  false,
						},
					},
				},
			},
		},
	}

	article := &parser.Article{
		Title:    "Foobar",
		Content:  "Abc\n" + tokenTable + "\ndef.",
		TokenMap: tokenMap,
		Images:   []string{},
	}

	statsGenerator := NewStatsGenerator(tokenMap)

	mockFile := setupCache()

	// Act
	statsOutputFilename, err := statsGenerator.Generate(article)

	// Assert
	stats := getAndAssertStats(t, err, article.Title, statsOutputFilename, mockFile)

	test.AssertEqual(t, 93, stats.NumberOfCharacters)
	test.AssertEqual(t, 1, stats.NumberOfInternalLinks)
	test.AssertEqual(t, 1, stats.NumberOfExternalLinks)
}

func TestGenerateCombinedStats(t *testing.T) {
	// Arrange
	config.Current.OutputType = config.OutputTypeStatsJson

	setupCache()

	mockFile := util.NewMockFile("result mock file")
	fsMock := util.CurrentFilesystem.(*util.MockFilesystem)
	fsMock.CreateFunc = func(name string) (util.FileLike, error) { return mockFile, nil }
	fsMock.ReadFileFunc = func(filename string) ([]byte, error) {
		stats := &articleStats{
			ArticleName:            "article " + filename,
			NumberOfCharacters:     50,
			NumberOfWords:          5,
			NumberOfInternalLinks:  6,
			NumberOfExternalLinks:  7,
			NumberOfImages:         8,
			NumberOfMath:           9,
			NumberOfRefDefinitions: 10,
			NumberOfRefUsages:      11,
			InternalLinks:          nil,
			UncoveredInternalLinks: nil,
		}
		return json.Marshal(&stats)
	}

	statsFiles := []string{"stats-file-1", "stats-file-2"}

	// Act
	err := GenerateCombinedStats(statsFiles, "output-path")

	// Assert
	test.AssertNil(t, err)

	stats := &articleStats{}
	err = json.Unmarshal(mockFile.WrittenBytes, stats)
	test.AssertNil(t, err)

	test.AssertEqual(t, 100, stats.NumberOfCharacters)
	test.AssertEqual(t, 10, stats.NumberOfWords)
	test.AssertEqual(t, 12, stats.NumberOfInternalLinks)
	test.AssertEqual(t, 14, stats.NumberOfExternalLinks)
	test.AssertEqual(t, 16, stats.NumberOfImages)
	test.AssertEqual(t, 18, stats.NumberOfMath)
	test.AssertEqual(t, 20, stats.NumberOfRefDefinitions)
	test.AssertEqual(t, 22, stats.NumberOfRefUsages)
}

func TestGenerateCombinedStats_top10UncoveredLinks(t *testing.T) {
	// Arrange
	config.Current.OutputType = config.OutputTypeStatsJson

	setupCache()

	mockFile := util.NewMockFile("result mock file")
	fsMock := util.CurrentFilesystem.(*util.MockFilesystem)
	fsMock.CreateFunc = func(name string) (util.FileLike, error) { return mockFile, nil }
	fsMock.ReadFileFunc = func(filename string) ([]byte, error) {
		stats := &articleStats{
			ArticleName: "article " + filename,
			InternalLinks: map[string]int{
				"a": 20,
				"b": 1,
				"c": 19,
				"d": 2,
				"e": 18,
				"f": 3,
				"g": 17,
				"h": 4,
				"i": 16,
				"j": 5,
				"k": 15,
				"l": 6,
				"m": 14,
				"n": 7,
				"o": 13,
			},
			UncoveredInternalLinks: map[string]int{
				"a": 20,
				"b": 1,
				"c": 19,
				"d": 2,
				"e": 18,
				"f": 3,
				"g": 17,
				"h": 4,
				"i": 16,
				"j": 5,
				"k": 15,
				"l": 6,
				"m": 14,
				"n": 7,
				"o": 13,
			},
		}
		return json.Marshal(&stats)
	}

	statsFiles := []string{"stats-file", "stats-file-2"}

	// Act
	err := GenerateCombinedStats(statsFiles, "output-path")

	// Assert
	test.AssertNil(t, err)

	stats := &articleStats{}
	err = json.Unmarshal(mockFile.WrittenBytes, stats)
	test.AssertNil(t, err)

	test.AssertMapEqual(t, map[string]int{
		"a": 40,
		"b": 2,
		"c": 38,
		"d": 4,
		"e": 36,
		"f": 6,
		"g": 34,
		"h": 8,
		"i": 32,
		"j": 10,
		"k": 30,
		"l": 12,
		"m": 28,
		"n": 14,
		"o": 26,
	}, stats.UncoveredInternalLinks)
	test.AssertMapEqual(t, map[string]int{
		"a": 40,
		"c": 38,
		"e": 36,
		"g": 34,
		"i": 32,
		"k": 30,
		"l": 12,
		"m": 28,
		"n": 14,
		"o": 26,
	}, stats.Top10UncoveredInternalLinks)
}
