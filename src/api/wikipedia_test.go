package api

import (
	"github.com/hauke96/sigolo/v2"
	"os"
	"path/filepath"
	"testing"
	"wiki2book/test"
)

func TestPostProcessImage_freshDownload_noPostProcessing(t *testing.T) {
	serviceMock := newMockImageProcessingService()
	imageProcessingService = serviceMock

	err := postProcessImage("./foo.svg", false, false, true)

	test.AssertNil(t, err)
	test.AssertEqual(t, 0, serviceMock.resizeAndCompressImageCalls)
	test.AssertEqual(t, 0, serviceMock.convertPdfToPngCalls)
	test.AssertEqual(t, 0, serviceMock.convertSvgToPngCalls)
}

func TestPostProcessImage_freshDownload_withSvgToPng(t *testing.T) {
	serviceMock := newMockImageProcessingService()
	imageProcessingService = serviceMock

	err := postProcessImage("./foo.svg", false, true, true)

	test.AssertNil(t, err)
	test.AssertEqual(t, 1, serviceMock.resizeAndCompressImageCalls)
	test.AssertEqual(t, 0, serviceMock.convertPdfToPngCalls)
	test.AssertEqual(t, 1, serviceMock.convertSvgToPngCalls)
}

func TestPostProcessImage_freshDownload_withPdfToPng(t *testing.T) {
	serviceMock := newMockImageProcessingService()
	imageProcessingService = serviceMock

	err := postProcessImage("./foo.svg", true, false, true)

	test.AssertNil(t, err)
	test.AssertEqual(t, 0, serviceMock.resizeAndCompressImageCalls)
	test.AssertEqual(t, 0, serviceMock.convertPdfToPngCalls)
	test.AssertEqual(t, 0, serviceMock.convertSvgToPngCalls)
}

func TestPostProcessImage_noFreshDownload_noPostProcessing(t *testing.T) {
	serviceMock := newMockImageProcessingService()
	imageProcessingService = serviceMock

	err := postProcessImage("./foo.svg", false, false, false)

	test.AssertNil(t, err)
	test.AssertEqual(t, 0, serviceMock.resizeAndCompressImageCalls)
	test.AssertEqual(t, 0, serviceMock.convertPdfToPngCalls)
	test.AssertEqual(t, 0, serviceMock.convertSvgToPngCalls)
}

func TestPostProcessImage_noFreshDownload_withSvgToPng_noExistingPng(t *testing.T) {
	serviceMock := newMockImageProcessingService()
	imageProcessingService = serviceMock

	svgFilepath := filepath.Join(test.TestTempDirName, "foo.svg")
	_, err := os.OpenFile(svgFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	err = postProcessImage(svgFilepath, false, true, false)

	test.AssertNil(t, err)
	test.AssertEqual(t, 1, serviceMock.resizeAndCompressImageCalls)
	test.AssertEqual(t, 0, serviceMock.convertPdfToPngCalls)
	test.AssertEqual(t, 1, serviceMock.convertSvgToPngCalls)
}

func TestPostProcessImage_noFreshDownload_withSvgToPng_alreadyExistingPng(t *testing.T) {
	serviceMock := newMockImageProcessingService()
	imageProcessingService = serviceMock

	svgFilepath := filepath.Join(test.TestTempDirName, "foo.svg")
	_, err := os.OpenFile(svgFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	pngFilepath := filepath.Join(test.TestTempDirName, "foo.svg.png")
	_, err = os.OpenFile(pngFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	err = postProcessImage(svgFilepath, false, true, false)

	test.AssertNil(t, err)
	test.AssertEqual(t, 0, serviceMock.resizeAndCompressImageCalls)
	test.AssertEqual(t, 0, serviceMock.convertPdfToPngCalls)
	test.AssertEqual(t, 0, serviceMock.convertSvgToPngCalls)
}

func TestPostProcessImage_noFreshDownload_withPdfToPng_noExistingPng(t *testing.T) {
	serviceMock := newMockImageProcessingService()
	imageProcessingService = serviceMock

	pdfFilepath := filepath.Join(test.TestTempDirName, "foo.pdf")
	_, err := os.OpenFile(pdfFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	err = postProcessImage(pdfFilepath, true, false, false)

	test.AssertNil(t, err)
	test.AssertEqual(t, 1, serviceMock.resizeAndCompressImageCalls)
	test.AssertEqual(t, 1, serviceMock.convertPdfToPngCalls)
	test.AssertEqual(t, 0, serviceMock.convertSvgToPngCalls)
}

func TestPostProcessImage_noFreshDownload_withPdfToPng_alreadyExistingPng(t *testing.T) {
	serviceMock := newMockImageProcessingService()
	imageProcessingService = serviceMock

	pdfFilepath := filepath.Join(test.TestTempDirName, "foo.pdf")
	_, err := os.OpenFile(pdfFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	pngFilepath := filepath.Join(test.TestTempDirName, "foo.pdf.png")
	_, err = os.OpenFile(pngFilepath, os.O_RDONLY|os.O_CREATE, 0666)
	sigolo.FatalCheck(err)

	err = postProcessImage(pdfFilepath, true, false, false)

	test.AssertNil(t, err)
	test.AssertEqual(t, 0, serviceMock.resizeAndCompressImageCalls)
	test.AssertEqual(t, 0, serviceMock.convertPdfToPngCalls)
	test.AssertEqual(t, 0, serviceMock.convertSvgToPngCalls)
}
