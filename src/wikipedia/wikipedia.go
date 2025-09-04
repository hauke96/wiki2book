package wikipedia

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"wiki2book/cache"
	"wiki2book/config"
	ownHttp "wiki2book/http"
	"wiki2book/image"
	"wiki2book/util"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

type WikiArticleDto struct {
	Parse WikiParseArticleDto `json:"parse"`
}

type WikiParseArticleDto struct {
	Title    string              `json:"title"`
	Wikitext WikiWildcardTextDto `json:"wikitext"`

	OriginalTitle string // Not set by Wikipedia but by wiki2book to remember the original title in case of redirects.
}

type WikiWildcardTextDto struct {
	Content string `json:"*"`
}

type WikiExpandedTemplateDto struct {
	ExpandTemplate WikitextDto `json:"expandtemplates"`
}

type WikitextDto struct {
	Content string `json:"wikitext"`
}

type WikipediaService interface {
	DownloadArticle(host string, title string) (*WikiArticleDto, error)
	DownloadImages(images []string, svgSizeToViewbox bool, pdfToPng bool, svgToPng bool) error
	EvaluateTemplate(template string, cacheFile string) (string, error)
	RenderMath(mathString string) (string, string, error)
}

type DefaultWikipediaService struct {
	cacheFolder                string // TODO can be removed?
	wikipediaInstance          string
	wikipediaHost              string
	wikipediaImageArticleHosts []string
	wikipediaImageHost         string
	wikipediaMathRestApi       string
	imageProcessingService     image.ImageProcessingService
	httpService                ownHttp.HttpService
}

func NewWikipediaService(cacheFolder string, wikipediaInstance string, wikipediaHost string, wikipediaImageInstances []string, wikipediaImageHost string, wikipediaMathRestApi string, imageProcessingService image.ImageProcessingService, httpClient ownHttp.HttpService) *DefaultWikipediaService {
	return &DefaultWikipediaService{
		cacheFolder:                cacheFolder,
		wikipediaInstance:          wikipediaInstance,
		wikipediaHost:              wikipediaHost,
		wikipediaImageArticleHosts: wikipediaImageInstances,
		wikipediaImageHost:         wikipediaImageHost,
		wikipediaMathRestApi:       wikipediaMathRestApi,
		imageProcessingService:     imageProcessingService,
		httpService:                httpClient,
	}
}

func (w *DefaultWikipediaService) DownloadArticle(host string, title string) (*WikiArticleDto, error) {
	titleWithoutWhitespaces := strings.ReplaceAll(title, " ", "_")
	escapedTitle := url.QueryEscape(titleWithoutWhitespaces)
	urlString := fmt.Sprintf("https://%s/w/api.php?action=parse&prop=wikitext&redirects=true&format=json&page=%s", host, escapedTitle)

	cachedFile := titleWithoutWhitespaces + ".json"
	cachedFilePath, _, err := w.httpService.DownloadAndCache(urlString, cache.ArticleCacheDirName, cachedFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to download article %s", title)
	}

	cachedResponseBytes, err := util.CurrentFilesystem.ReadFile(cachedFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read body bytes")
	}

	wikiArticleDto := &WikiArticleDto{}
	err = json.Unmarshal(cachedResponseBytes, wikiArticleDto)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error parsing JSON from article %s.%s/%s", w.wikipediaInstance, w.wikipediaHost, title))
	}

	// Use the given title. When the article is behind a redirect, the actual title is used which might be unexpected
	// for a caller of this function.
	wikiArticleDto.Parse.OriginalTitle = title

	return wikiArticleDto, nil
}

// DownloadImages tries to download the given images from a couple of sources (wikipedia/wikimedia instances). The
// downloaded images will be in the output folder. Some images might be redirects, so the redirect will be resolved,
// that's why the article cache folder is needed as well.
func (w *DefaultWikipediaService) DownloadImages(images []string, svgSizeToViewbox bool, pdfToPng bool, svgToPng bool) error {
	sigolo.Debugf("Downloading images or loading them from cache:\n%s", strings.Join(images, "\n"))
	for _, image := range images {
		sigolo.Debugf("Download image %s", image)

		downloadErr := w.downloadImageUsingAllSources(image, svgSizeToViewbox, pdfToPng, svgToPng)
		if downloadErr != nil {
			return downloadErr
		}
	}
	return nil
}

func (w *DefaultWikipediaService) downloadImageUsingAllSources(image string, svgSizeToViewbox bool, pdfToPng bool, svgToPng bool) error {
	var downloadErr error

	for i, articleHost := range w.wikipediaImageArticleHosts {
		var outputFilepath string
		var freshlyDownloaded bool
		isLastSource := i == len(w.wikipediaImageArticleHosts)-1
		// TODO Bug: The image instance is not used to download the image article, which should be the case.
		outputFilepath, freshlyDownloaded, downloadErr = w.downloadImage(articleHost, image, svgSizeToViewbox)
		if downloadErr != nil {
			if isLastSource {
				// We tried every single image source and couldn't find the image.
				sigolo.Errorf("Could not downloading image %s from any image article source. Error of last image source: %s", image, downloadErr.Error())
			} else {
				// This image is not available at the current source. Maybe one of the following sources hold the
				// image. Therefore, this is not a real error that needs to be handled.
				sigolo.Debugf("Could not downloading image %s from source %s.%s: %s", image, articleHost, w.wikipediaHost, downloadErr.Error())
			}
			continue
		}

		err := w.postProcessImage(outputFilepath, pdfToPng, svgToPng, freshlyDownloaded)
		if err != nil {
			return err
		}

		break
	}

	return downloadErr
}

func (w *DefaultWikipediaService) postProcessImage(outputFilepath string, pdfToPng bool, svgToPng bool, freshlyDownloaded bool) error {
	if pdfToPng && filepath.Ext(strings.ToLower(outputFilepath)) == util.FileEndingPdf {
		outputPngFilepath := util.GetPngPathForPdf(outputFilepath)
		pdfAlreadyExists := util.PathExists(outputPngFilepath)
		if !pdfAlreadyExists {
			err := w.imageProcessingService.ConvertPdfToPng(outputFilepath, outputPngFilepath, config.Current.CommandTemplatePdfToPng)
			if err != nil {
				return err
			}

			// We pretend this is a fresh download, because the PNG is indeed fresh
			freshlyDownloaded = true
		}
		outputFilepath = outputPngFilepath
	} else if svgToPng && filepath.Ext(strings.ToLower(outputFilepath)) == util.FileEndingSvg {
		outputPngFilepath := util.GetPngPathForSvg(outputFilepath)
		pngAlreadyExists := util.PathExists(outputPngFilepath)
		if !pngAlreadyExists {
			err := w.imageProcessingService.ConvertSvgToPng(outputFilepath, outputPngFilepath, config.Current.CommandTemplateSvgToPng)
			if err != nil {
				return err
			}

			// We pretend this is a fresh download, because the PNG is indeed fresh
			freshlyDownloaded = true
		}
		outputFilepath = outputPngFilepath
	}

	// If the file is new, rescale it using ImageMagick.
	if freshlyDownloaded && outputFilepath != "" && filepath.Ext(strings.ToLower(outputFilepath)) != util.FileEndingSvg {
		err := w.imageProcessingService.ResizeAndCompressImage(outputFilepath, config.Current.CommandTemplateImageProcessing)
		if err != nil {
			return err
		}
	}

	return nil
}

// downloadImage downloads the given image (e.g. "File:foo.jpg") to the given folder and returns the filepath as first
// return value. When the file already exists, then the second value is false, otherwise true (for fresh downloads or
// in case of errors). Whenever an error is returned, the article cache folder is needed as some files might be
// redirects and such a redirect counts as article.
func (w *DefaultWikipediaService) downloadImage(imageArticleHost string, imageNameWithPrefix string, svgSizeToViewbox bool) (string, bool, error) {
	// TODO handle colons in file names
	imageName := "File:" + strings.Split(imageNameWithPrefix, ":")[1]
	sigolo.Debugf("Download article file for image '%s' from '%s'", imageName, imageArticleHost)
	imageArticle, err := w.DownloadArticle(imageArticleHost, imageName)
	if err != nil {
		return "", true, err
	}

	// Replace spaces with underscore because wikimedia doesn't know spaces in file names:

	originalImageNameWithPrefix := imageArticle.Parse.OriginalTitle
	originalImageName := strings.Split(originalImageNameWithPrefix, ":")[1]
	originalImageName = strings.ReplaceAll(originalImageName, " ", "_")

	actualImageNameWithPrefix := imageArticle.Parse.Title
	actualImageName := strings.Split(actualImageNameWithPrefix, ":")[1]
	actualImageName = strings.ReplaceAll(actualImageName, " ", "_")

	md5sum := fmt.Sprintf("%x", md5.Sum([]byte(actualImageName)))
	sigolo.Debugf("Download actual image '%s' from Wikimedia instance '%s'", actualImageName, imageArticleHost)
	sigolo.Tracef("  Original name: %s", originalImageName)
	sigolo.Tracef("  Actual image name (after possible redirects): %s", actualImageNameWithPrefix)
	sigolo.Tracef("  MD5 of redirected image name: %s", md5sum)

	// In case the host ist a wiki host with subdomain (i.e. "de.wikipedia.org"), we use the subdomain as the instance
	// path segment below. Image URLs are e.g. "upload.wikimedia.org/wikipedia/de/..." where the "de" is the instance,
	// which should be used when the image article has been located at "de.wikipedia.org".
	isWikiDomainWithSubdomain, err := regexp.MatchString(".+\\.wiki.+\\..+", imageArticleHost)
	wikiInstancePathSegment := ""
	if isWikiDomainWithSubdomain {
		wikiInstancePathSegment = strings.SplitN(imageArticleHost, ".", 2)[0]
	}

	imageUrl := fmt.Sprintf("https://%s/wikipedia/%s/%c/%c%c/%s", w.wikipediaImageHost, wikiInstancePathSegment, md5sum[0], md5sum[0], md5sum[1], url.QueryEscape(actualImageName))

	cachedFilePath, freshlyDownloaded, err := w.httpService.DownloadAndCache(imageUrl, cache.ImageCacheDirName, originalImageName)
	if err != nil {
		return "", freshlyDownloaded, err
	}

	if freshlyDownloaded && svgSizeToViewbox && filepath.Ext(cachedFilePath) == util.FileEndingSvg {
		err = image.MakeSvgSizeAbsolute(cachedFilePath)
		if err != nil {
			sigolo.Errorf("Unable to make size of SVG %s absolute. This error will be ignored, since false errors exist for the XML parsing of SVGs. Error: %+v", cachedFilePath, err)
		}
	}

	return cachedFilePath, freshlyDownloaded, nil
}

func (w *DefaultWikipediaService) EvaluateTemplate(template string, cacheFile string) (string, error) {
	sigolo.Debugf("Evaluate template %s (hash/filename: %s)", util.TruncString(template), cacheFile)

	urlString := fmt.Sprintf("https://%s.%s/w/api.php?action=expandtemplates&format=json&prop=wikitext&text=%s", w.wikipediaInstance, w.wikipediaHost, url.QueryEscape(template))
	cacheFilePath, _, err := w.httpService.DownloadAndCache(urlString, cache.TemplateCacheDirName, cacheFile)
	if err != nil {
		return "", errors.Wrapf(err, "Error calling evaluation API and caching result for template:\n%s", template)
	}

	evaluatedTemplateString, err := util.CurrentFilesystem.ReadFile(cacheFilePath)
	if err != nil {
		return "", errors.Wrapf(err, "Reading cached template file %s failed", cacheFilePath)
	}

	evaluatedTemplate := &WikiExpandedTemplateDto{}
	err = json.Unmarshal(evaluatedTemplateString, evaluatedTemplate)
	if err != nil {
		return "", errors.Wrapf(err, "Unable to unmarshal template string:\n%s", evaluatedTemplateString)
	}

	return evaluatedTemplate.ExpandTemplate.Content, nil
}

func (w *DefaultWikipediaService) RenderMath(mathString string) (string, string, error) {
	sigolo.Debugf("Render math %s", util.TruncString(mathString))
	sigolo.Tracef("  Complete math text: %s", mathString)

	mathApiUrl := w.wikipediaMathRestApi

	mathSvgFilename, err := w.getMathResource(mathString)
	if err != nil {
		return "", "", err
	}

	imageSvgUrl := mathApiUrl + "/render/svg/" + mathSvgFilename
	cachedSvgFile, _, err := w.httpService.DownloadAndCache(imageSvgUrl, cache.ImageCacheDirName, mathSvgFilename+util.FileEndingSvg)
	if err != nil {
		return "", "", err
	}

	if config.Current.MathConverter == config.MathConverterNone {
		return cachedSvgFile, cachedSvgFile, nil
	} else if config.Current.MathConverter == config.MathConverterWikimedia {
		imagePngUrl := mathApiUrl + "/render/png/" + mathSvgFilename
		cachedPngFile, _, err := w.httpService.DownloadAndCache(imagePngUrl, cache.ImageCacheDirName, mathSvgFilename+util.FileEndingPng)
		if err != nil {
			return "", "", err
		}
		return cachedSvgFile, cachedPngFile, nil
	} else if config.Current.MathConverter == config.MathConverterInternal {
		cachedPngFile := filepath.Join(cache.ImageCacheDirName, mathSvgFilename+util.FileEndingPng)
		err = w.imageProcessingService.ConvertSvgToPng(cachedSvgFile, cachedPngFile, config.Current.CommandTemplateMathSvgToPng)
		if err != nil {
			return "", "", err
		}
		return cachedSvgFile, cachedPngFile, nil
	}

	return "", "", errors.New("No supported math converter found")
}

// getMathResource uses a POST request to generate the SVG from the given math TeX string. This function returns the SimpleSvgAttributes filename.
func (w *DefaultWikipediaService) getMathResource(mathString string) (string, error) {
	urlString := w.wikipediaMathRestApi + "/check/tex"

	// Wikipedia itself adds the "{\displaystyle ...}" part. Having this here as well generated the same IDs for the
	// formulae as in the original article. This is not only nice for debugging but also might increase speed due to
	// caching on the Wikimedia servers.
	requestData := "q=" + url.QueryEscape(fmt.Sprintf(`{\displaystyle %s}`, mathString))

	// If file exists -> ignore
	filename := util.Hash(mathString)
	outputFilepath := cache.GetFilePathInCache(cache.MathCacheDirName, filename)
	if _, err := util.CurrentFilesystem.Stat(outputFilepath); err == nil {
		mathSvgFilenameBytes, err := util.CurrentFilesystem.ReadFile(outputFilepath)
		mathSvgFilename := string(mathSvgFilenameBytes)
		if err != nil {
			return "", errors.Wrapf(err, "Unable to read cache file %s for math string %s", outputFilepath, util.TruncString(mathString))
		}
		sigolo.Debugf("File %s does already exist. Skip.", outputFilepath)
		return mathSvgFilename, nil
	}

	response, err := w.httpService.PostFormEncoded(urlString, requestData)

	responseBodyText := ""
	if response != nil {
		responseBodyReader := response.Body
		if responseBodyReader != nil {
			defer responseBodyReader.Close()
		}
		responseBodyText = util.ReaderToString(response.Body)
	}

	if err != nil {
		return "", errors.Wrapf(err, "Response body for math '%s' on URL %s : %s", mathString, urlString, responseBodyText)
	}

	if response == nil {
		return "", errors.Errorf("No error but empty response returned for math '%s' on URL %s", mathString, urlString)
	}

	if response.StatusCode != http.StatusOK {
		return "", errors.Errorf("Rendering math failed with status code %d for math '%s' on URL %s with body: %s", response.StatusCode, mathString, urlString, responseBodyText)
	}

	locationHeader := response.Header.Get("x-resource-location")
	if locationHeader == "" {
		return "", errors.Errorf("Unable to get location header for math '%s' on URL %s with body: %s", mathString, urlString, responseBodyText)
	}

	err = cache.CacheToFile(cache.MathCacheDirName, filename, io.NopCloser(strings.NewReader(locationHeader)))
	if err != nil {
		return "", errors.Wrapf(err, "Unable to cache math resource for math string \"%s\" to %s", util.TruncString(mathString), outputFilepath)
	}

	return locationHeader, nil
}
