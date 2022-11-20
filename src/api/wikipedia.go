package api

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/util"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type WikiArticleDto struct {
	Parse WikiParseArticleDto `json:"parse"`
}

type WikiParseArticleDto struct {
	Title    string              `json:"title"`
	Wikitext WikiWildcardTextDto `json:"wikitext"`
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

func DownloadArticle(language string, title string, cacheFolder string) (*WikiArticleDto, error) {
	titleWithoutWhitespaces := strings.ReplaceAll(title, " ", "_")
	escapedTitle := url.QueryEscape(titleWithoutWhitespaces)
	urlString := fmt.Sprintf("https://%s.wikipedia.org/w/api.php?action=parse&prop=wikitext&format=json&page=%s", language, escapedTitle)

	cachedFile := titleWithoutWhitespaces + ".json"
	cachedFilePath, err := downloadAndCache(urlString, cacheFolder, cachedFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to download article %s", title)
	}

	bodyBytes, err := ioutil.ReadFile(cachedFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read body bytes")
	}

	wikiArticleDto := &WikiArticleDto{}
	err = json.Unmarshal(bodyBytes, wikiArticleDto)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error parsing JSON from article %s/%s", language, title))
	}

	return wikiArticleDto, nil
}

// DownloadImages tries to download the given images from a couple of sources (wikipedia/wikimedia instances). The
// downloaded images will bein the output folder. Some images might be redirects, so the redirect must be resolved,
// that's why the article cache folder is needed as well.
func DownloadImages(images []string, outputFolder string, articleFolder string) error {
	for _, image := range images {
		var downloadErr error = nil
		var outputFilepath string

		for _, source := range imageSources {
			outputFilepath, downloadErr = downloadImage(image, outputFolder, articleFolder, source)
			if downloadErr != nil {
				sigolo.Error("Error downloading image %s from source %s: %s. Try next source.\n%+v", image, source, downloadErr.Error(), downloadErr)
				continue
			}

			// If the file is new, rescale it using ImageMagick.
			if outputFilepath != "" && !strings.HasSuffix(outputFilepath, ".svg") {
				err2 := processImage(outputFilepath)
				if err2 != nil {
					return err2
				}
			}

			break
		}

		if downloadErr != nil {
			return downloadErr
		}
	}
	return nil
}

// downloadImage downloads the given image (e.g. "File:foo.jpg") to the given folder. When the file already exists,
// nothing is done and "", nil will be returned. When the file has been downloaded "filename", nil will be returned.
// The article cache folder is needed as some files might be redirects and such a redirect counts as article.
func downloadImage(imageNameWithPrefix string, outputFolder string, articleFolder string, source string) (string, error) {
	// TODO handle colons in file names
	imageName := strings.Split(imageNameWithPrefix, ":")[1]
	imageNameWithPrefix, err := followRedirectIfNeeded(source, "File:"+imageName, articleFolder)
	if err != nil {
		return "", err
	}
	imageName = strings.Split(imageNameWithPrefix, ":")[1]

	md5sum := fmt.Sprintf("%x", md5.Sum([]byte(imageName)))
	sigolo.Debug(imageName)
	sigolo.Debug(md5sum)

	url := fmt.Sprintf("https://upload.wikimedia.org/wikipedia/%s/%c/%c%c/%s", source, md5sum[0], md5sum[0], md5sum[1], url.QueryEscape(imageName))
	sigolo.Debug(url)

	return downloadAndCache(url, outputFolder, imageName)
}

// followRedirectIfNeeded returns the page name behind a redirect. If the article page with the given title (which can
// also be an image like "File:foo.jpg") is a redirect, the article/file name pointed to is returned. Spaces are
// replaced by underscores. So if "File:foo.jpg" redirects to "File:bar with spaces.jpg", then
// "File:bar_with_spaces.jpg" is returned. If there's no redirect, the original title parameter will be returned.
func followRedirectIfNeeded(source string, title string, cacheFolder string) (string, error) {
	article, err := DownloadArticle(source, title, cacheFolder)
	if err != nil {
		return title, err
	}

	regexMatch := redirectRegex.FindStringSubmatch(article.Parse.Wikitext.Content)
	if regexMatch != nil && regexMatch[1] != "" {
		// Replace spaces by string as the Wikipedia API only handles file names with underscore instead of spaces
		return strings.ReplaceAll(regexMatch[1], " ", "_"), nil
	}

	return title, nil
}

func EvaluateTemplate(template string, cacheFolder string, cacheFile string) (string, error) {
	sigolo.Info("Evaluate template %s", util.TruncString(template))

	urlString := "https://de.wikipedia.org/w/api.php?action=expandtemplates&format=json&prop=wikitext&text=" + url.QueryEscape(template)
	cacheFilePath, err := downloadAndCache(urlString, cacheFolder, cacheFile)
	if err != nil {
		return "", errors.Wrapf(err, "Error calling evaluation API and caching result for template:\n%s", template)
	}

	evaluatedTemplateString, err := ioutil.ReadFile(cacheFilePath)
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

func RenderMath(mathString string, imageCacheFolder string, mathCacheFolder string) (string, string, error) {
	sigolo.Info("Render math %s", mathString)

	mathString = url.QueryEscape(mathString)

	mathSvgFilename, err := getMathResource(mathString, mathCacheFolder)
	if err != nil {
		return "", "", errors.Wrapf(err, "Unable to get math resource for math string %s", util.TruncString(mathString))
	}

	imageSvgUrl := "https://wikimedia.org/api/rest_v1/media/math/render/svg/" + mathSvgFilename
	cachedSvgFile, err := downloadAndCache(imageSvgUrl, imageCacheFolder, mathSvgFilename+".svg")
	if err != nil {
		return "", "", err
	}

	imagePngUrl := "https://wikimedia.org/api/rest_v1/media/math/render/png/" + mathSvgFilename
	cachedPngFile, err := downloadAndCache(imagePngUrl, imageCacheFolder, mathSvgFilename+".png")
	if err != nil {
		return "", "", err
	}

	return cachedSvgFile, cachedPngFile, nil
}

// getMathResource uses a POST request to generate the SVG from the given math TeX string. This function returns the SVG filename.
func getMathResource(mathString string, cacheFolder string) (string, error) {
	urlString := "https://wikimedia.org/api/rest_v1/media/math/check/tex"
	requestData := fmt.Sprintf("q=%s", mathString)

	// If file exists -> ignore
	filename := util.Hash(mathString)
	outputFilepath := filepath.Join(cacheFolder, filename)
	if _, err := os.Stat(outputFilepath); err == nil {
		mathSvgFilenameBytes, err := ioutil.ReadFile(outputFilepath)
		mathSvgFilename := string(mathSvgFilenameBytes)
		if err != nil {
			return "", errors.Wrapf(err, "Unable to read cache file %s for math string %s", outputFilepath, util.TruncString(mathString))
		}
		sigolo.Info("File %s does already exist. Skip.", outputFilepath)
		return mathSvgFilename, nil
	}

	sigolo.Debug("Rendering math %s", util.TruncString(mathString))

	sigolo.Debug("Make GET request to %s", urlString)
	response, err := httpClient.Post(urlString, "application/x-www-form-urlencoded", strings.NewReader(requestData))
	if err != nil {
		return "", errors.Wrapf(err, "Unable to call render URL for math %s", mathString)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return "", errors.Errorf("Rendering Math: Response returned with status code %d", response.StatusCode)
	}

	locationHeader := response.Header.Get("x-resource-location")
	if locationHeader == "" {
		return "", errors.Errorf("Unable to get location header for math %s", mathString)
	}

	err = cacheToFile(cacheFolder, filename, ioutil.NopCloser(strings.NewReader(locationHeader)))
	if err != nil {
		return "", errors.Wrapf(err, "Unable to cache math resource for math string \"%s\" to %s", util.TruncString(mathString), outputFilepath)
	}

	return locationHeader, nil
}