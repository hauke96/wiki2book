package wikipedia

type DummyWikipediaService struct {
	EvaluateTemplateResponse string
}

func (d *DummyWikipediaService) DownloadArticle(host string, title string) (*WikiArticleDto, error) {
	return nil, nil
}

func (d *DummyWikipediaService) DownloadImages(images []string, svgSizeToViewbox bool, pdfToPng bool, svgToPng bool) error {
	return nil
}

func (d *DummyWikipediaService) EvaluateTemplate(template string, cacheFile string) (string, error) {
	return d.EvaluateTemplateResponse, nil
}

func (d *DummyWikipediaService) RenderMath(mathString string) (string, string, error) {
	return "", "", nil
}
