package api

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
	"io"
	"net/http"
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

func DownloadArticle(wikipediaInstance string, wikipediaHost string, title string, cacheFolder string) (*WikiArticleDto, error) {
	titleWithoutWhitespaces := strings.ReplaceAll(title, " ", "_")
	escapedTitle := url.QueryEscape(titleWithoutWhitespaces)
	urlString := fmt.Sprintf("https://%s.%s/w/api.php?action=parse&prop=wikitext&redirects=true&format=json&page=%s", wikipediaInstance, wikipediaHost, escapedTitle)

	cachedFile := titleWithoutWhitespaces + ".json"
	cachedFilePath, _, err := downloadAndCache(urlString, cacheFolder, cachedFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to download article %s", title)
	}

	cachedResponseBytes, err := os.ReadFile(cachedFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read body bytes")
	}

	wikiArticleDto := &WikiArticleDto{}
	err = json.Unmarshal(cachedResponseBytes, wikiArticleDto)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error parsing JSON from article %s.%s/%s", wikipediaInstance, wikipediaHost, title))
	}

	// Use the given title. When the article is behind a redirect, the actual title is used which might be unexpected
	// for a caller of this function.
	wikiArticleDto.Parse.OriginalTitle = title

	return wikiArticleDto, nil
}

// DownloadImages tries to download the given images from a couple of sources (wikipedia/wikimedia instances). The
// downloaded images will be in the output folder. Some images might be redirects, so the redirect will be resolved,
// that's why the article cache folder is needed as well.
func DownloadImages(images []string, outputFolder string, articleFolder string, svgSizeToViewbox bool, pdfToPng bool, svgToPng bool) error {
	sigolo.Debugf("Downloading images or loading them from cache:\n%s", strings.Join(images, "\n"))
	for _, image := range images {
		sigolo.Infof("Download image %s", image)

		downloadErr := downloadImageUsingAllSources(image, outputFolder, articleFolder, svgSizeToViewbox, pdfToPng, svgToPng)
		if downloadErr != nil {
			return downloadErr
		}
	}
	return nil
}

func downloadImageUsingAllSources(image string, outputFolder string, articleFolder string, svgSizeToViewbox bool, pdfToPng bool, svgToPng bool) error {
	var downloadErr error

	for i, instance := range config.Current.WikipediaImageArticleInstances {
		var outputFilepath string
		var freshlyDownloaded bool
		isLastSource := i == len(config.Current.WikipediaImageArticleInstances)-1
		outputFilepath, freshlyDownloaded, downloadErr = downloadImage(image, outputFolder, articleFolder, config.Current.WikipediaImageHost, instance, config.Current.WikipediaHost, svgSizeToViewbox)
		if downloadErr != nil {
			if isLastSource {
				// We tried every single image source and couldn't find the image.
				sigolo.Errorf("Could not downloading image %s from any image article source. Error of last image source: %s", image, downloadErr.Error())
			} else {
				// This image is not available at the current source. Maybe one of the following sources hold the
				// image. Therefore, this is not a real error that needs to be handled.
				sigolo.Debugf("Could not downloading image %s from source %s.%s: %s", image, instance, config.Current.WikipediaHost, downloadErr.Error())
			}
			continue
		}

		err := postProcessImage(outputFilepath, pdfToPng, svgToPng, freshlyDownloaded)
		if err != nil {
			return err
		}

		break
	}

	return downloadErr
}

func postProcessImage(outputFilepath string, pdfToPng bool, svgToPng bool, freshlyDownloaded bool) error {
	if pdfToPng && filepath.Ext(strings.ToLower(outputFilepath)) == util.FileEndingPdf {
		outputPngFilepath := util.GetPngPathForPdf(outputFilepath)
		if _, err := os.Stat(outputPngFilepath); err != nil {
			err = convertPdfToPng(outputFilepath, outputPngFilepath, config.Current.CommandTemplatePdfToPng)
			if err != nil {
				return err
			}
		}
		outputFilepath = outputPngFilepath
	} else if svgToPng && filepath.Ext(strings.ToLower(outputFilepath)) == util.FileEndingSvg {
		outputPngFilepath := util.GetPngPathForSvg(outputFilepath)
		if _, err := os.Stat(outputPngFilepath); err != nil {
			err = convertSvgToPng(outputFilepath, outputPngFilepath, config.Current.CommandTemplateSvgToPng)
			if err != nil {
				return err
			}
		}
		outputFilepath = outputPngFilepath
	}

	// If the file is new, rescale it using ImageMagick.
	if freshlyDownloaded && outputFilepath != "" && filepath.Ext(strings.ToLower(outputFilepath)) != util.FileEndingSvg {
		err := resizeAndCompressImage(outputFilepath, config.Current.CommandTemplateImageProcessing)
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
func downloadImage(imageNameWithPrefix string, outputFolder string, articleFolder string, wikipediaImageHost string, wikipediaInstance string, wikipediaHost string, svgSizeToViewbox bool) (string, bool, error) {
	// TODO handle colons in file names
	imageName := "File:" + strings.Split(imageNameWithPrefix, ":")[1]
	sigolo.Debugf("Download article file for image '%s' from Wikipedia instance '%s.%s'", imageName, wikipediaInstance, wikipediaHost)
	imageArticle, err := DownloadArticle(wikipediaInstance, wikipediaHost, imageName, articleFolder)
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
	sigolo.Debugf("Download actual image '%s' from Wikimedia instance '%s'", actualImageName, wikipediaInstance)
	sigolo.Tracef("  Original name: %s", originalImageName)
	sigolo.Tracef("  Actual image name (after possible redirects): %s", actualImageNameWithPrefix)
	sigolo.Tracef("  MD5 of redirected image name: %s", md5sum)

	imageUrl := fmt.Sprintf("https://%s/wikipedia/%s/%c/%c%c/%s", wikipediaImageHost, wikipediaInstance, md5sum[0], md5sum[0], md5sum[1], url.QueryEscape(actualImageName))

	cachedFilePath, freshlyDownloaded, err := downloadAndCache(imageUrl, outputFolder, originalImageName)
	if err != nil {
		return "", freshlyDownloaded, err
	}

	if freshlyDownloaded && svgSizeToViewbox && filepath.Ext(cachedFilePath) == util.FileEndingSvg {
		err = util.MakeSvgSizeAbsolute(cachedFilePath)
		if err != nil {
			sigolo.Errorf("Unable to make size of SVG %s absolute. This error will be ignored, since false errors exist for the XML parsing of SVGs. Error: %+v", cachedFilePath, err)
		}
	}

	return cachedFilePath, freshlyDownloaded, nil
}

func EvaluateTemplate(template string, cacheFolder string, cacheFile string) (string, error) {
	sigolo.Debugf("Evaluate template %s (hash/filename: %s)", util.TruncString(template), cacheFile)

	urlString := fmt.Sprintf("https://%s.%s/w/api.php?action=expandtemplates&format=json&prop=wikitext&text=%s", config.Current.WikipediaInstance, config.Current.WikipediaHost, url.QueryEscape(template))
	cacheFilePath, _, err := downloadAndCache(urlString, cacheFolder, cacheFile)
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
	sigolo.Debugf("Render math %s", util.TruncString(mathString))
	sigolo.Tracef("  Complete math text: %s", mathString)

	mathApiUrl := config.Current.WikipediaMathRestApi

	mathSvgFilename, err := getMathResource(mathString, mathCacheFolder)
	if err != nil {
		return "", "", err
	}

	imageSvgUrl := mathApiUrl + "/render/svg/" + mathSvgFilename
	cachedSvgFile, _, err := downloadAndCache(imageSvgUrl, imageCacheFolder, mathSvgFilename+util.FileEndingSvg)
	if err != nil {
		return "", "", err
	}

	if config.Current.MathConverter == config.MathConverterNone {
		return cachedSvgFile, cachedSvgFile, nil
	} else if config.Current.MathConverter == config.MathConverterWikimedia {
		imagePngUrl := mathApiUrl + "/render/png/" + mathSvgFilename
		cachedPngFile, _, err := downloadAndCache(imagePngUrl, imageCacheFolder, mathSvgFilename+util.FileEndingPng)
		if err != nil {
			return "", "", err
		}
		return cachedSvgFile, cachedPngFile, nil
	} else if config.Current.MathConverter == config.MathConverterInternal {
		cachedPngFile := filepath.Join(imageCacheFolder, mathSvgFilename+util.FileEndingPng)
		err = convertSvgToPng(cachedSvgFile, cachedPngFile, config.Current.CommandTemplateMathSvgToPng)
		if err != nil {
			return "", "", err
		}
		return cachedSvgFile, cachedPngFile, nil
	}

	return "", "", errors.New("No supported math converter found")
}

// getMathResource uses a POST request to generate the SVG from the given math TeX string. This function returns the SimpleSvgAttributes filename.
func getMathResource(mathString string, cacheFolder string) (string, error) {
	urlString := config.Current.WikipediaMathRestApi + "/check/tex"

	// Wikipedia itself adds the "{\displaystyle ...}" part. Having this here as well generated the same IDs for the
	// formulae as in the original article. This is not only nice for debugging but also might increase speed due to
	// caching on the Wikimedia servers.
	requestData := "q=" + url.QueryEscape(fmt.Sprintf(`{\displaystyle %s}`, mathString))

	// If file exists -> ignore
	filename := util.Hash(mathString)
	outputFilepath := filepath.Join(cacheFolder, filename)
	if _, err := os.Stat(outputFilepath); err == nil {
		mathSvgFilenameBytes, err := os.ReadFile(outputFilepath)
		mathSvgFilename := string(mathSvgFilenameBytes)
		if err != nil {
			return "", errors.Wrapf(err, "Unable to read cache file %s for math string %s", outputFilepath, util.TruncString(mathString))
		}
		sigolo.Debugf("File %s does already exist. Skip.", outputFilepath)
		return mathSvgFilename, nil
	}

	sigolo.Debugf("Make POST request to %s with request data: %s", urlString, requestData)
	response, err := httpClient.Post(urlString, "application/x-www-form-urlencoded", strings.NewReader(requestData))

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

	err = cacheToFile(cacheFolder, filename, io.NopCloser(strings.NewReader(locationHeader)))
	if err != nil {
		return "", errors.Wrapf(err, "Unable to cache math resource for math string \"%s\" to %s", util.TruncString(mathString), outputFilepath)
	}

	return locationHeader, nil
}
