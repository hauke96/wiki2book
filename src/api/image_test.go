package api

type mockImageProcessingService struct {
	resizeAndCompressImageCalls int
	convertPdfToPngCalls        int
	convertSvgToPngCalls        int
}

func newMockImageProcessingService() *mockImageProcessingService {
	return &mockImageProcessingService{}
}

// resizeAndCompressImage will convert and rescale the image so that it's suitable for eBooks.
func (s *mockImageProcessingService) resizeAndCompressImage(imageFilepath string, commandTemplate string) error {
	s.resizeAndCompressImageCalls++
	return nil
}

// convertPdfToPng will convert the given PDF file into a PNG image at the given location. This conversion does neither
// rescale nor process the image in any other way, use resizeAndCompressImage accordingly.
func (s *mockImageProcessingService) convertPdfToPng(inputPdfFilepath string, outputPngFilepath string, commandTemplate string) error {
	s.convertPdfToPngCalls++
	return nil
}

func (s *mockImageProcessingService) convertSvgToPng(svgFile string, pngFile string, commandTemplate string) error {
	s.convertSvgToPngCalls++
	return nil
}
