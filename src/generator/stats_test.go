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
	test.AssertNil(t, err)
	test.AssertEqual(t, config.Current.CacheDir+"/stats/"+article.Title+".json", statsOutputFilename)

	stats := &articleStats{}
	err = json.Unmarshal(mockFile.WrittenBytes, stats)
	test.AssertNil(t, err)

	test.AssertEqual(t, 14, stats.NumberOfCharacters)
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
	test.AssertNil(t, err)
	test.AssertEqual(t, config.Current.CacheDir+"/stats/"+article.Title+".json", statsOutputFilename)

	stats := &articleStats{}
	err = json.Unmarshal(mockFile.WrittenBytes, stats)
	test.AssertNil(t, err)

	test.AssertEqual(t, 21, stats.NumberOfCharacters)
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
	test.AssertNil(t, err)
	test.AssertEqual(t, config.Current.CacheDir+"/stats/"+article.Title+".json", statsOutputFilename)

	stats := &articleStats{}
	err = json.Unmarshal(mockFile.WrittenBytes, stats)
	test.AssertNil(t, err)

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
			NumberOfInternalLinks:  6,
			NumberOfExternalLinks:  7,
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
	test.AssertEqual(t, 12, stats.NumberOfInternalLinks)
	test.AssertEqual(t, 14, stats.NumberOfExternalLinks)
}
