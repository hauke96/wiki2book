package wikipedia

type DummyWikipediaService struct {
	EvaluateTemplateResponse string
}

func (d *DummyWikipediaService) DownloadArticle(title string, cacheFolder string) (*WikiArticleDto, error) {
	return nil, nil
}

func (d *DummyWikipediaService) DownloadImages(images []string, outputFolder string, articleFolder string, svgSizeToViewbox bool, pdfToPng bool, svgToPng bool) error {
	return nil
}

func (d *DummyWikipediaService) EvaluateTemplate(template string, cacheFolder string, cacheFile string) (string, error) {
	return d.EvaluateTemplateResponse, nil
}

func (d *DummyWikipediaService) RenderMath(mathString string, imageCacheFolder string, mathCacheFolder string) (string, string, error) {
	return "", "", nil
}
