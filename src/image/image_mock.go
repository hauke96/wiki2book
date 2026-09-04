package image

type MockImageProcessingService struct {
	ResizeAndCompressImageCalls int
	ConvertToPngCalls           int
}

func NewMockImageProcessingService() *MockImageProcessingService {
	return &MockImageProcessingService{}
}

// resizeAndCompressImage will convert and rescale the image so that it's suitable for eBooks.
func (s *MockImageProcessingService) ResizeAndCompressImage(imageFilepath string, commandTemplate string) error {
	s.ResizeAndCompressImageCalls++
	return nil
}

func (s *MockImageProcessingService) ConvertToPng(inputFile string, pngFile string, commandTemplate string) error {
	s.ConvertToPngCalls++
	return nil
}
