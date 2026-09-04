package wikipedia

import (
	"bytes"
	"encoding/json"
	"io"
	netHttp "net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"wiki2book/config"
	"wiki2book/http"
	"wiki2book/image"
	"wiki2book/test"
	"wiki2book/util"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

func TestPostProcessImage_freshDownload_noPostProcessing(t *testing.T) {
	// Arrange
	mockHttpClient := http.NewMockHttpService(
		func(url string, cacheFolder string, filename string) (string, bool, error) {
			return "", true, nil
		},
		nil,
		nil,
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", []string{}, "", imageProcessingServiceMock, mockHttpClient)

	config.Current = config.NewDefaultConfig()
	config.Current.CommandTemplateImageProcessing = ""
	config.Current.CommandTemplateSvgToPng = ""

	// Act
	err := wikipediaService.postProcessImage("./foo.svg", true)

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ConvertToPngCalls)
}

func TestPostProcessImage_freshDownload_withSvgToPng(t *testing.T) {
	// Arrange
	mockHttpClient := http.NewMockHttpService(
		func(url string, cacheFolder string, filename string) (string, bool, error) {
			return "", true, nil
		},
		nil,
		nil,
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", []string{}, "", imageProcessingServiceMock, mockHttpClient)

	config.Current = config.NewDefaultConfig()
	config.Current.CommandTemplateImageProcessing = "some-command"
	config.Current.CommandTemplateSvgToPng = "some-command"

	// Act
	err := wikipediaService.postProcessImage("./foo.svg", true)

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, 1, imageProcessingServiceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 1, imageProcessingServiceMock.ConvertToPngCalls)
}

func TestPostProcessImage_freshDownload_withPdfToPng(t *testing.T) {
	// Arrange
	mockHttpClient := http.NewMockHttpService(
		func(url string, cacheFolder string, filename string) (string, bool, error) {
			return "", true, nil
		},
		nil,
		nil,
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", []string{}, "", imageProcessingServiceMock, mockHttpClient)

	config.Current = config.NewDefaultConfig()
	config.Current.CommandTemplateImageProcessing = "some-command"
	config.Current.CommandTemplatePdfToPng = "some-command"

	// Act
	err := wikipediaService.postProcessImage("./foo.pdf", true)

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, 1, imageProcessingServiceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 1, imageProcessingServiceMock.ConvertToPngCalls)
}

func TestPostProcessImage_freshDownload_withWebpToPng(t *testing.T) {
	// Arrange
	mockHttpClient := http.NewMockHttpService(
		func(url string, cacheFolder string, filename string) (string, bool, error) {
			return "", true, nil
		},
		nil,
		nil,
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", []string{}, "", imageProcessingServiceMock, mockHttpClient)

	config.Current = config.NewDefaultConfig()
	config.Current.CommandTemplateImageProcessing = "some-command"
	config.Current.CommandTemplateWebpToPng = "some-command"

	// Act
	err := wikipediaService.postProcessImage("./foo.webp", true)

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, 1, imageProcessingServiceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 1, imageProcessingServiceMock.ConvertToPngCalls)
}

func TestPostProcessImage_noFreshDownload_noPostProcessing(t *testing.T) {
	// Arrange
	mockHttpClient := http.NewMockHttpService(
		func(url string, cacheFolder string, filename string) (string, bool, error) {
			return "", true, nil
		},
		nil,
		nil,
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", []string{}, "", imageProcessingServiceMock, mockHttpClient)

	config.Current = config.NewDefaultConfig()
	config.Current.CommandTemplateImageProcessing = ""
	config.Current.CommandTemplateSvgToPng = ""

	// Act
	err := wikipediaService.postProcessImage("./foo.svg", false)

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ConvertToPngCalls)
}

func TestPostProcessImage_noFreshDownload_withSvgToPng_noExistingPng(t *testing.T) {
	// Arrange
	mockHttpClient := http.NewMockHttpService(
		func(url string, cacheFolder string, filename string) (string, bool, error) {
			return "", true, nil
		},
		nil,
		nil,
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", []string{}, "", imageProcessingServiceMock, mockHttpClient)

	svgFilepath := filepath.Join(test.TestTempDirName, "foo.svg")
	_, err := os.OpenFile(svgFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	config.Current = config.NewDefaultConfig()
	config.Current.CommandTemplateImageProcessing = "some-command"
	config.Current.CommandTemplateSvgToPng = "some-command"

	// Act
	err = wikipediaService.postProcessImage(svgFilepath, false)

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, 1, imageProcessingServiceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 1, imageProcessingServiceMock.ConvertToPngCalls)
}

func TestPostProcessImage_noFreshDownload_withSvgToPng_alreadyExistingPng(t *testing.T) {
	mockHttpClient := http.NewMockHttpService(
		func(url string, cacheFolder string, filename string) (string, bool, error) {
			return "", true, nil
		},
		nil,
		nil,
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", []string{}, "", imageProcessingServiceMock, mockHttpClient)

	svgFilepath := filepath.Join(test.TestTempDirName, "foo.svg")
	_, err := os.OpenFile(svgFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	pngFilepath := filepath.Join(test.TestTempDirName, "foo.svg.png")
	_, err = os.OpenFile(pngFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	err = wikipediaService.postProcessImage(svgFilepath, false)

	test.AssertNil(t, err)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ConvertToPngCalls)
}

func TestPostProcessImage_noFreshDownload_withPdfToPng_noExistingPng(t *testing.T) {
	mockHttpClient := http.NewMockHttpService(
		func(url string, cacheFolder string, filename string) (string, bool, error) {
			return "", true, nil
		},
		nil,
		nil,
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", []string{}, "", imageProcessingServiceMock, mockHttpClient)

	pdfFilepath := filepath.Join(test.TestTempDirName, "foo.pdf")
	_, err := os.OpenFile(pdfFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	err = wikipediaService.postProcessImage(pdfFilepath, false)

	test.AssertNil(t, err)
	test.AssertEqual(t, 1, imageProcessingServiceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 1, imageProcessingServiceMock.ConvertToPngCalls)
}

func TestPostProcessImage_noFreshDownload_withPdfToPng_alreadyExistingPng(t *testing.T) {
	mockHttpClient := http.NewMockHttpService(
		func(url string, cacheFolder string, filename string) (string, bool, error) {
			return "", true, nil
		},
		nil,
		nil,
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", []string{}, "", imageProcessingServiceMock, mockHttpClient)

	pdfFilepath := filepath.Join(test.TestTempDirName, "foo.pdf")
	_, err := os.OpenFile(pdfFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	pngFilepath := filepath.Join(test.TestTempDirName, "foo.pdf.png")
	_, err = os.OpenFile(pngFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	err = wikipediaService.postProcessImage(pdfFilepath, false)

	test.AssertNil(t, err)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ConvertToPngCalls)
}

func TestDownladImage(t *testing.T) {
	cachedArticleFilepath := filepath.Join(test.TestCacheFolder, "Foo.jpg.json")
	err := os.WriteFile(cachedArticleFilepath, []byte(`{"parse":{"title":"File:Foo.jpg","pageid":80546,"redirects":[],"wikitext":{"*":""}}}`), 0600)
	sigolo.FatalCheck(err)

	cachedImageFilepath := filepath.Join(test.TestCacheFolder, "File:foo.jpg")
	err = os.WriteFile(cachedImageFilepath, []byte(`foobar`), 0600)
	sigolo.FatalCheck(err)

	requestedArticleUrl := ""
	requestedImageUrl := ""

	mockHttpClient := http.NewMockHttpService(
		func(url string, cacheFolder string, filename string) (string, bool, error) {
			if strings.Contains(url, "api.php") {
				requestedArticleUrl = url
				return cachedArticleFilepath, true, nil
			} else if strings.Contains(url, "upload") {
				requestedImageUrl = url
				return cachedImageFilepath, true, nil
			}

			return "", false, errors.New("no mock behavior for url " + url)
		},
		nil,
		nil,
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", []string{}, "upload.wikimedia.org", imageProcessingServiceMock, mockHttpClient)

	downloadImage, freshlyDownloaded, err := wikipediaService.downloadImage("en.wikipedia.org", "File:foo.jpg")

	test.AssertNil(t, err)
	test.AssertTrue(t, freshlyDownloaded)
	test.AssertEqual(t, cachedImageFilepath, downloadImage)
	test.AssertEqual(t, "https://en.wikipedia.org/w/api.php?action=parse&prop=wikitext&redirects=true&format=json&page=File%3Afoo.jpg", requestedArticleUrl)
	test.AssertEqual(t, "https://upload.wikimedia.org/wikipedia/en/0/06/Foo.jpg", requestedImageUrl)
}

func TestEvaluateTemplate_newTemplate(t *testing.T) {
	key := "7499ae1f1f8e45a9a95bdeb610ebf13cc4157667"
	expectedTemplateContent := "<div class=\"hauptartikel\" role=\"navigation\"><span class=\"hauptartikel-pfeil\" title=\"siehe\" aria-hidden=\"true\" role=\"presentation\">→ </span>''<span class=\"hauptartikel-text\">Hauptartikel</span>: [[Sternentstehung]]''</div>"
	jsonBytes, _ := json.Marshal(&WikiExpandedTemplateDto{ExpandTemplate: WikitextDto{Content: expectedTemplateContent}})

	cachedFilepath := filepath.Join(test.TestCacheFolder, key)

	fsMock := util.NewDefaultMockFilesystem()
	fsMock.ReadFileFunc = func(name string) ([]byte, error) { return jsonBytes, nil }
	util.CurrentFilesystem = fsMock

	mockHttpService := http.NewMockHttpService(
		nil,
		func(url string, method, cacheFolder string, filename string) (string, bool, error) {
			return cachedFilepath, true, nil
		},
		func(url, method, contentType string) (resp *netHttp.Response, err error) {
			return &netHttp.Response{
				Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
				StatusCode: netHttp.StatusOK,
			}, nil
		},
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", []string{}, "", imageProcessingServiceMock, mockHttpService)

	// Evaluate content
	content, err := wikipediaService.EvaluateTemplate("{{Hauptartikel|Sternentstehung}}", key)
	test.AssertNil(t, err)
	test.AssertEqual(t, 0, mockHttpService.DownloadAndCacheCounter)
	test.AssertEqual(t, 1, mockHttpService.PostAndCacheCounter)
	test.AssertEqual(t, 0, mockHttpService.PerformHttpRequestCounter)
	test.AssertEqual(t, expectedTemplateContent, content)
}
