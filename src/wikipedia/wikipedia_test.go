package wikipedia

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
	"io"
	netHttp "net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"wiki2book/http"
	"wiki2book/image"
	"wiki2book/test"
)

func TestPostProcessImage_freshDownload_noPostProcessing(t *testing.T) {
	mockHttpClient := http.NewMockHttpService(
		func(url string, cacheFolder string, filename string) (string, bool, error) {
			return "", true, nil
		},
		nil,
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", "", []string{}, "", "", imageProcessingServiceMock, mockHttpClient)

	err := wikipediaService.postProcessImage("./foo.svg", false, false, true)

	test.AssertNil(t, err)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ConvertPdfToPngCalls)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ConvertSvgToPngCalls)
}

func TestPostProcessImage_freshDownload_withSvgToPng(t *testing.T) {
	mockHttpClient := http.NewMockHttpService(
		func(url string, cacheFolder string, filename string) (string, bool, error) {
			return "", true, nil
		},
		nil,
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", "", []string{}, "", "", imageProcessingServiceMock, mockHttpClient)

	err := wikipediaService.postProcessImage("./foo.svg", false, true, true)

	test.AssertNil(t, err)
	test.AssertEqual(t, 1, imageProcessingServiceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ConvertPdfToPngCalls)
	test.AssertEqual(t, 1, imageProcessingServiceMock.ConvertSvgToPngCalls)
}

func TestPostProcessImage_freshDownload_withPdfToPng(t *testing.T) {
	mockHttpClient := http.NewMockHttpService(
		func(url string, cacheFolder string, filename string) (string, bool, error) {
			return "", true, nil
		},
		nil,
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", "", []string{}, "", "", imageProcessingServiceMock, mockHttpClient)

	err := wikipediaService.postProcessImage("./foo.svg", true, false, true)

	test.AssertNil(t, err)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ConvertPdfToPngCalls)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ConvertSvgToPngCalls)
}

func TestPostProcessImage_noFreshDownload_noPostProcessing(t *testing.T) {
	mockHttpClient := http.NewMockHttpService(
		func(url string, cacheFolder string, filename string) (string, bool, error) {
			return "", true, nil
		},
		nil,
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", "", []string{}, "", "", imageProcessingServiceMock, mockHttpClient)

	err := wikipediaService.postProcessImage("./foo.svg", false, false, false)

	test.AssertNil(t, err)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ConvertPdfToPngCalls)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ConvertSvgToPngCalls)
}

func TestPostProcessImage_noFreshDownload_withSvgToPng_noExistingPng(t *testing.T) {
	mockHttpClient := http.NewMockHttpService(
		func(url string, cacheFolder string, filename string) (string, bool, error) {
			return "", true, nil
		},
		nil,
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", "", []string{}, "", "", imageProcessingServiceMock, mockHttpClient)

	svgFilepath := filepath.Join(test.TestTempDirName, "foo.svg")
	_, err := os.OpenFile(svgFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	err = wikipediaService.postProcessImage(svgFilepath, false, true, false)

	test.AssertNil(t, err)
	test.AssertEqual(t, 1, imageProcessingServiceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ConvertPdfToPngCalls)
	test.AssertEqual(t, 1, imageProcessingServiceMock.ConvertSvgToPngCalls)
}

func TestPostProcessImage_noFreshDownload_withSvgToPng_alreadyExistingPng(t *testing.T) {
	mockHttpClient := http.NewMockHttpService(
		func(url string, cacheFolder string, filename string) (string, bool, error) {
			return "", true, nil
		},
		nil,
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", "", []string{}, "", "", imageProcessingServiceMock, mockHttpClient)

	svgFilepath := filepath.Join(test.TestTempDirName, "foo.svg")
	_, err := os.OpenFile(svgFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	pngFilepath := filepath.Join(test.TestTempDirName, "foo.svg.png")
	_, err = os.OpenFile(pngFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	err = wikipediaService.postProcessImage(svgFilepath, false, true, false)

	test.AssertNil(t, err)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ConvertPdfToPngCalls)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ConvertSvgToPngCalls)
}

func TestPostProcessImage_noFreshDownload_withPdfToPng_noExistingPng(t *testing.T) {
	mockHttpClient := http.NewMockHttpService(
		func(url string, cacheFolder string, filename string) (string, bool, error) {
			return "", true, nil
		},
		nil,
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", "", []string{}, "", "", imageProcessingServiceMock, mockHttpClient)

	pdfFilepath := filepath.Join(test.TestTempDirName, "foo.pdf")
	_, err := os.OpenFile(pdfFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	err = wikipediaService.postProcessImage(pdfFilepath, true, false, false)

	test.AssertNil(t, err)
	test.AssertEqual(t, 1, imageProcessingServiceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 1, imageProcessingServiceMock.ConvertPdfToPngCalls)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ConvertSvgToPngCalls)
}

func TestPostProcessImage_noFreshDownload_withPdfToPng_alreadyExistingPng(t *testing.T) {
	mockHttpClient := http.NewMockHttpService(
		func(url string, cacheFolder string, filename string) (string, bool, error) {
			return "", true, nil
		},
		nil,
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", "", []string{}, "", "", imageProcessingServiceMock, mockHttpClient)

	pdfFilepath := filepath.Join(test.TestTempDirName, "foo.pdf")
	_, err := os.OpenFile(pdfFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	pngFilepath := filepath.Join(test.TestTempDirName, "foo.pdf.png")
	_, err = os.OpenFile(pngFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	err = wikipediaService.postProcessImage(pdfFilepath, true, false, false)

	test.AssertNil(t, err)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ConvertPdfToPngCalls)
	test.AssertEqual(t, 0, imageProcessingServiceMock.ConvertSvgToPngCalls)
}

func TestDownladImage(t *testing.T) {
	cachedArticleFilepath := filepath.Join(test.TestCacheFolder, "Foo.jpg.json")
	err := os.WriteFile(cachedArticleFilepath, []byte(`{"parse":{"title":"File:Foo.jpg","pageid":80546,"redirects":[],"wikitext":{"*":""}}}`), 0600)
	sigolo.FatalCheck(err)

	cachedImageFilepath := filepath.Join(test.TestCacheFolder, "File:foo.jpg")
	err = os.WriteFile(cachedImageFilepath, []byte(`foobar`), 0600)
	sigolo.FatalCheck(err)

	mockHttpClient := http.NewMockHttpService(
		func(url string, cacheFolder string, filename string) (string, bool, error) {
			if strings.Contains(url, "page=") {
				return cachedArticleFilepath, true, nil
			}
			return cachedImageFilepath, true, nil
		},
		nil,
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", "", []string{}, "", "", imageProcessingServiceMock, mockHttpClient)

	downloadImage, freshlyDownloaded, err := wikipediaService.downloadImage("File:foo.jpg", test.TestCacheFolder, test.TestCacheFolder, false)

	test.AssertNil(t, err)
	test.AssertTrue(t, freshlyDownloaded)
	test.AssertEqual(t, cachedImageFilepath, downloadImage)
}

func TestEvaluateTemplate_newTemplate(t *testing.T) {
	key := "7499ae1f1f8e45a9a95bdeb610ebf13cc4157667"
	expectedTemplateContent := "<div class=\"hauptartikel\" role=\"navigation\"><span class=\"hauptartikel-pfeil\" title=\"siehe\" aria-hidden=\"true\" role=\"presentation\">→ </span>''<span class=\"hauptartikel-text\">Hauptartikel</span>: [[Sternentstehung]]''</div>"
	jsonBytes, _ := json.Marshal(&WikiExpandedTemplateDto{ExpandTemplate: WikitextDto{Content: expectedTemplateContent}})
	expectedTemplateFileContent := string(jsonBytes)

	cachedFilepath := filepath.Join(test.TestCacheFolder, key)
	err := os.WriteFile(cachedFilepath, jsonBytes, 0666)
	sigolo.FatalCheck(err)

	mockHttpService := http.NewMockHttpService(
		func(url string, cacheFolder string, filename string) (string, bool, error) {
			return cachedFilepath, true, nil
		},
		func(url, contentType string) (resp *netHttp.Response, err error) {
			return &netHttp.Response{
				Body:       io.NopCloser(bytes.NewReader(jsonBytes)),
				StatusCode: netHttp.StatusOK,
			}, nil
		},
	)
	imageProcessingServiceMock := image.NewMockImageProcessingService()
	wikipediaService := NewWikipediaService("", "", "", []string{}, "", "", imageProcessingServiceMock, mockHttpService)

	// Evaluate content
	content, err := wikipediaService.EvaluateTemplate("{{Hauptartikel|Sternentstehung}}", test.TestCacheFolder, key)
	test.AssertNil(t, err)
	test.AssertEqual(t, 1, mockHttpService.DownloadAndCacheCounter)
	test.AssertEqual(t, 0, mockHttpService.PostFormEncodedCounter)
	test.AssertEqual(t, expectedTemplateContent, content)
	test.AssertTrue(t, hasLocalTemplate(key, test.TestCacheFolder))

	// Read template content from disk
	expectedContent, err := getLocalTemplate(key, test.TestCacheFolder)
	test.AssertNil(t, err)
	test.AssertEqual(t, expectedTemplateFileContent, expectedContent)
}

func hasLocalTemplate(key string, templateFolder string) bool {
	templateFilepath := filepath.Join(templateFolder, key)

	file, err := os.Open(templateFilepath)
	if file == nil || errors.Is(err, os.ErrNotExist) {
		return false
	}
	defer file.Close()

	return true
}

func getLocalTemplate(key string, templateFolder string) (string, error) {
	templateFilepath := filepath.Join(templateFolder, key)

	content, err := os.ReadFile(templateFilepath)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Error reading template %s from %s", key, templateFilepath))
	}

	return string(content), nil
}
