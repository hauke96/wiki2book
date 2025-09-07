package image

type mockImageProcessingService struct {
	ResizeAndCompressImageCalls int
	ConvertToPngCalls           int
}

func NewMockImageProcessingService() *mockImageProcessingService {
	return &mockImageProcessingService{}
}

// resizeAndCompressImage will convert and rescale the image so that it's suitable for eBooks.
func (s *mockImageProcessingService) ResizeAndCompressImage(imageFilepath string, commandTemplate string) error {
	s.ResizeAndCompressImageCalls++
	return nil
}

func (s *mockImageProcessingService) ConvertToPng(inputFile string, pngFile string, commandTemplate string) error {
	s.ConvertToPngCalls++
	return nil
}
