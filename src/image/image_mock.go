package image

type mockImageProcessingService struct {
	ResizeAndCompressImageCalls int
	ConvertPdfToPngCalls        int
	ConvertSvgToPngCalls        int
}

func NewMockImageProcessingService() *mockImageProcessingService {
	return &mockImageProcessingService{}
}

// resizeAndCompressImage will convert and rescale the image so that it's suitable for eBooks.
func (s *mockImageProcessingService) ResizeAndCompressImage(imageFilepath string, commandTemplate string) error {
	s.ResizeAndCompressImageCalls++
	return nil
}

// convertPdfToPng will convert the given PDF file into a PNG image at the given location. This conversion does neither
// rescale nor process the image in any other way, use resizeAndCompressImage accordingly.
func (s *mockImageProcessingService) ConvertPdfToPng(inputPdfFilepath string, outputPngFilepath string, commandTemplate string) error {
	s.ConvertPdfToPngCalls++
	return nil
}

func (s *mockImageProcessingService) ConvertSvgToPng(svgFile string, pngFile string, commandTemplate string) error {
	s.ConvertSvgToPngCalls++
	return nil
}
