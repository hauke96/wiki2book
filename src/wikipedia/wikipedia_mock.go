package wikipedia

type MockWikipediaService struct {
	DownloadArticleFunc  func(host string, title string) (*WikiArticleDto, error)
	DownloadImagesFunc   func(images []string) error
	EvaluateTemplateFunc func(template string, cacheFile string) (string, error)
	RenderMathFunc       func(mathString string) (string, string, error)
}

func NewMockWikipediaService() *MockWikipediaService {
	return &MockWikipediaService{
		DownloadArticleFunc:  func(host string, title string) (*WikiArticleDto, error) { return nil, nil },
		DownloadImagesFunc:   func(images []string) error { return nil },
		EvaluateTemplateFunc: func(template string, cacheFile string) (string, error) { return "", nil },
		RenderMathFunc:       func(mathString string) (string, string, error) { return "", "", nil },
	}
}

func (m *MockWikipediaService) DownloadArticle(host string, title string) (*WikiArticleDto, error) {
	return m.DownloadArticleFunc(host, title)
}

func (m *MockWikipediaService) DownloadImages(images []string) error {
	return m.DownloadImagesFunc(images)
}

func (m *MockWikipediaService) EvaluateTemplate(template string, cacheFile string) (string, error) {
	return m.EvaluateTemplateFunc(template, cacheFile)
}

func (m *MockWikipediaService) RenderMath(mathString string) (string, string, error) {
	return m.RenderMathFunc(mathString)
}
