package api

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/pkg/errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"wiki2book/config"
	"wiki2book/util"
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

func DownloadArticle(wikipediaInstance string, title string, cacheFolder string) (*WikiArticleDto, error) {
	titleWithoutWhitespaces := strings.ReplaceAll(title, " ", "_")
	escapedTitle := url.QueryEscape(titleWithoutWhitespaces)
	urlString := fmt.Sprintf("https://%s.wikipedia.org/w/api.php?action=parse&prop=wikitext&redirects=true&format=json&page=%s", wikipediaInstance, escapedTitle)

	cachedFile := titleWithoutWhitespaces + ".json"
	cachedFilePath, err := downloadAndCache(urlString, cacheFolder, cachedFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to download article %s", title)
	}

	bodyBytes, err := os.ReadFile(cachedFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read body bytes")
	}

	wikiArticleDto := &WikiArticleDto{}
	err = json.Unmarshal(bodyBytes, wikiArticleDto)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error parsing JSON from article %s/%s", wikipediaInstance, title))
	}

	// Use the given title. When the article is behind a redirect, the actual title is used which might be unexpected
	// for a caller of this function.
	wikiArticleDto.Parse.OriginalTitle = title

	return wikiArticleDto, nil
}

// DownloadImages tries to download the given images from a couple of sources (wikipedia/wikimedia instances). The
// downloaded images will be in the output folder. Some images might be redirects, so the redirect will be resolved,
// that's why the article cache folder is needed as well.
func DownloadImages(images []string, outputFolder string, articleFolder string, svgSizeToViewbox bool) error {
	for _, image := range images {
		var downloadErr error = nil
		var outputFilepath string

		for i, source := range config.Current.WikipediaImageArticleInstances {
			isLastSource := i == len(config.Current.WikipediaImageArticleInstances)-1
			outputFilepath, downloadErr = downloadImage(image, outputFolder, articleFolder, source, svgSizeToViewbox)
			if downloadErr != nil {
				if isLastSource {
					sigolo.Error("Could not downloading image %s from any image article source: %s\n", image, downloadErr.Error())
				} else {
					// That an image is not available at one source is a common situation and not an error that needs to be handled.
					sigolo.Debug("Could not downloading image %s from source %s: %s", image, source, downloadErr.Error())
				}
				continue
			}

			// If the file is new, rescale it using ImageMagick.
			if outputFilepath != "" && !strings.HasSuffix(strings.ToLower(outputFilepath), ".svg") {
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
func downloadImage(imageNameWithPrefix string, outputFolder string, articleFolder string, wikipediaInstance string, svgSizeToViewbox bool) (string, error) {
	// TODO handle colons in file names
	imageName := strings.Split(imageNameWithPrefix, ":")[1]
	imageArticle, err := DownloadArticle(wikipediaInstance, "File:"+imageName, articleFolder)
	if err != nil {
		return "", err
	}

	// Replace spaces with underscore because wikimedia doesn't know spaces in file names:

	originalImageNameWithPrefix := imageArticle.Parse.OriginalTitle
	originalImageName := strings.Split(originalImageNameWithPrefix, ":")[1]
	originalImageName = strings.ReplaceAll(originalImageName, " ", "_")

	actualImageNameWithPrefix := imageArticle.Parse.Title
	actualImageName := strings.Split(actualImageNameWithPrefix, ":")[1]
	actualImageName = strings.ReplaceAll(actualImageName, " ", "_")

	md5sum := fmt.Sprintf("%x", md5.Sum([]byte(actualImageName)))
	sigolo.Debug("Original image name: %s", originalImageName)
	sigolo.Debug("Actual image name (after possible redirects): %s", actualImageNameWithPrefix)
	sigolo.Debug("MD5 of redirected image name: %s", md5sum)

	imageUrl := fmt.Sprintf("https://upload.wikimedia.org/wikipedia/%s/%c/%c%c/%s", wikipediaInstance, md5sum[0], md5sum[0], md5sum[1], url.QueryEscape(actualImageName))
	sigolo.Debug(imageUrl)

	cachedFilePath, err := downloadAndCache(imageUrl, outputFolder, originalImageName)
	if err != nil {
		return "", err
	}

	if svgSizeToViewbox && filepath.Ext(cachedFilePath) == ".svg" {
		err = util.MakeSvgSizeAbsolute(cachedFilePath)
		if err != nil {
			return "", err
		}
	}

	return cachedFilePath, nil
}

func EvaluateTemplate(template string, cacheFolder string, cacheFile string) (string, error) {
	sigolo.Debug("Evaluate template %s", util.TruncString(template))

	urlString := fmt.Sprintf("https://%s.wikipedia.org/w/api.php?action=expandtemplates&format=json&prop=wikitext&text=%s", config.Current.WikipediaInstance, url.QueryEscape(template))
	cacheFilePath, err := downloadAndCache(urlString, cacheFolder, cacheFile)
	if err != nil {
		return "", errors.Wrapf(err, "Error calling evaluation API and caching result for template:\n%s", template)
	}

	evaluatedTemplateString, err := os.ReadFile(cacheFilePath)
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
	sigolo.Debug("Render math %s", mathString)

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

// getMathResource uses a POST request to generate the SVG from the given math TeX string. This function returns the SimpleSvgAttributes filename.
func getMathResource(mathString string, cacheFolder string) (string, error) {
	urlString := "https://wikimedia.org/api/rest_v1/media/math/check/tex"
	requestData := fmt.Sprintf("q=%s", mathString)

	// If file exists -> ignore
	filename := util.Hash(mathString)
	outputFilepath := filepath.Join(cacheFolder, filename)
	if _, err := os.Stat(outputFilepath); err == nil {
		mathSvgFilenameBytes, err := os.ReadFile(outputFilepath)
		mathSvgFilename := string(mathSvgFilenameBytes)
		if err != nil {
			return "", errors.Wrapf(err, "Unable to read cache file %s for math string %s", outputFilepath, util.TruncString(mathString))
		}
		sigolo.Debug("File %s does already exist. Skip.", outputFilepath)
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

	err = cacheToFile(cacheFolder, filename, io.NopCloser(strings.NewReader(locationHeader)))
	if err != nil {
		return "", errors.Wrapf(err, "Unable to cache math resource for math string \"%s\" to %s", util.TruncString(mathString), outputFilepath)
	}

	return locationHeader, nil
}
