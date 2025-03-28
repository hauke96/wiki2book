package wikipedia

import (
	"encoding/json"
	"fmt"
	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"testing"
	"wiki2book/http"
	"wiki2book/image"
	"wiki2book/test"
)

func TestPostProcessImage_freshDownload_noPostProcessing(t *testing.T) {
	serviceMock := image.NewMockImageProcessingService()
	imageProcessingService = serviceMock

	err := postProcessImage("./foo.svg", false, false, true)

	test.AssertNil(t, err)
	test.AssertEqual(t, 0, serviceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 0, serviceMock.ConvertPdfToPngCalls)
	test.AssertEqual(t, 0, serviceMock.ConvertSvgToPngCalls)
}

func TestPostProcessImage_freshDownload_withSvgToPng(t *testing.T) {
	serviceMock := image.NewMockImageProcessingService()
	imageProcessingService = serviceMock

	err := postProcessImage("./foo.svg", false, true, true)

	test.AssertNil(t, err)
	test.AssertEqual(t, 1, serviceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 0, serviceMock.ConvertPdfToPngCalls)
	test.AssertEqual(t, 1, serviceMock.ConvertSvgToPngCalls)
}

func TestPostProcessImage_freshDownload_withPdfToPng(t *testing.T) {
	serviceMock := image.NewMockImageProcessingService()
	imageProcessingService = serviceMock

	err := postProcessImage("./foo.svg", true, false, true)

	test.AssertNil(t, err)
	test.AssertEqual(t, 0, serviceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 0, serviceMock.ConvertPdfToPngCalls)
	test.AssertEqual(t, 0, serviceMock.ConvertSvgToPngCalls)
}

func TestPostProcessImage_noFreshDownload_noPostProcessing(t *testing.T) {
	serviceMock := image.NewMockImageProcessingService()
	imageProcessingService = serviceMock

	err := postProcessImage("./foo.svg", false, false, false)

	test.AssertNil(t, err)
	test.AssertEqual(t, 0, serviceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 0, serviceMock.ConvertPdfToPngCalls)
	test.AssertEqual(t, 0, serviceMock.ConvertSvgToPngCalls)
}

func TestPostProcessImage_noFreshDownload_withSvgToPng_noExistingPng(t *testing.T) {
	serviceMock := image.NewMockImageProcessingService()
	imageProcessingService = serviceMock

	svgFilepath := filepath.Join(test.TestTempDirName, "foo.svg")
	_, err := os.OpenFile(svgFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	err = postProcessImage(svgFilepath, false, true, false)

	test.AssertNil(t, err)
	test.AssertEqual(t, 1, serviceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 0, serviceMock.ConvertPdfToPngCalls)
	test.AssertEqual(t, 1, serviceMock.ConvertSvgToPngCalls)
}

func TestPostProcessImage_noFreshDownload_withSvgToPng_alreadyExistingPng(t *testing.T) {
	serviceMock := image.NewMockImageProcessingService()
	imageProcessingService = serviceMock

	svgFilepath := filepath.Join(test.TestTempDirName, "foo.svg")
	_, err := os.OpenFile(svgFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	pngFilepath := filepath.Join(test.TestTempDirName, "foo.svg.png")
	_, err = os.OpenFile(pngFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	err = postProcessImage(svgFilepath, false, true, false)

	test.AssertNil(t, err)
	test.AssertEqual(t, 0, serviceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 0, serviceMock.ConvertPdfToPngCalls)
	test.AssertEqual(t, 0, serviceMock.ConvertSvgToPngCalls)
}

func TestPostProcessImage_noFreshDownload_withPdfToPng_noExistingPng(t *testing.T) {
	serviceMock := image.NewMockImageProcessingService()
	imageProcessingService = serviceMock

	pdfFilepath := filepath.Join(test.TestTempDirName, "foo.pdf")
	_, err := os.OpenFile(pdfFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	err = postProcessImage(pdfFilepath, true, false, false)

	test.AssertNil(t, err)
	test.AssertEqual(t, 1, serviceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 1, serviceMock.ConvertPdfToPngCalls)
	test.AssertEqual(t, 0, serviceMock.ConvertSvgToPngCalls)
}

func TestPostProcessImage_noFreshDownload_withPdfToPng_alreadyExistingPng(t *testing.T) {
	serviceMock := image.NewMockImageProcessingService()
	imageProcessingService = serviceMock

	pdfFilepath := filepath.Join(test.TestTempDirName, "foo.pdf")
	_, err := os.OpenFile(pdfFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	pngFilepath := filepath.Join(test.TestTempDirName, "foo.pdf.png")
	_, err = os.OpenFile(pngFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	err = postProcessImage(pdfFilepath, true, false, false)

	test.AssertNil(t, err)
	test.AssertEqual(t, 0, serviceMock.ResizeAndCompressImageCalls)
	test.AssertEqual(t, 0, serviceMock.ConvertPdfToPngCalls)
	test.AssertEqual(t, 0, serviceMock.ConvertSvgToPngCalls)
}

func TestEvaluateTemplate_newTemplate(t *testing.T) {
	key := "7499ae1f1f8e45a9a95bdeb610ebf13cc4157667"
	expectedTemplateContent := "<div class=\"hauptartikel\" role=\"navigation\"><span class=\"hauptartikel-pfeil\" title=\"siehe\" aria-hidden=\"true\" role=\"presentation\">→ </span>''<span class=\"hauptartikel-text\">Hauptartikel</span>: [[Sternentstehung]]''</div>"
	jsonBytes, _ := json.Marshal(&WikiExpandedTemplateDto{ExpandTemplate: WikitextDto{Content: expectedTemplateContent}})
	expectedTemplateFileContent := string(jsonBytes)

	mockHttpClient := http.NewMockHttp(expectedTemplateFileContent, 200)
	httpClient = mockHttpClient

	wikipediaService := NewWikipediaService("", "", "", []string{}, "", "")

	// Evaluate content
	content, err := wikipediaService.EvaluateTemplate("{{Hauptartikel|Sternentstehung}}", test.TestCacheFolder, key)
	test.AssertNil(t, err)
	test.AssertEqual(t, 1, mockHttpClient.GetCalls)
	test.AssertEqual(t, 0, mockHttpClient.PostCalls)
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
