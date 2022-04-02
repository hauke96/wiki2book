package api

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/util"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type WikiPageDto struct {
	Parse WikiParsePageDto `json:"parse"`
}

type WikiParsePageDto struct {
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

type MockHttpClient struct {
	Response   string
	StatusCode int
	GetCalls   int
	PostCalls  int
}

func (h *MockHttpClient) Get(url string) (resp *http.Response, err error) {
	h.GetCalls++
	return &http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(h.Response))),
		StatusCode: h.StatusCode,
	}, nil
}

func (h *MockHttpClient) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	h.PostCalls++
	return &http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(h.Response))),
		StatusCode: h.StatusCode,
	}, nil
}

func MockHttp(response string, statusCode int) *MockHttpClient {
	mockedHttpClient := &MockHttpClient{
		response,
		statusCode,
		0,
		0,
	}
	httpClient = mockedHttpClient
	return mockedHttpClient
}

var imageSources = []string{"commons", "de"}
var httpClient = GetDefaultHttpClient()

func DownloadPage(language string, title string, cacheFolder string) (*WikiPageDto, error) {
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

	wikiPageDto := &WikiPageDto{}
	err = json.Unmarshal(bodyBytes, wikiPageDto)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error parsing JSON from article %s/%s", language, title))
	}

	return wikiPageDto, nil
}

func DownloadImages(images []string, outputFolder string) error {
	for _, image := range images {
		var err error = nil
		var outputFilepath string

		for _, source := range imageSources {
			outputFilepath, err = downloadImage(image, outputFolder, source)
			if err != nil {
				sigolo.Error("Error downloading image %s from source %s: %s. Try next source.", image, source, err.Error())
				continue
			}

			// If the file is new, rescale it using ImageMagick.
			if outputFilepath != "" && !strings.HasSuffix(outputFilepath, ".svg") {
				const imgSize = 600
				cmd := exec.Command("convert", outputFilepath, "-colorspace", "gray", "-separate", "-average", "-resize", fmt.Sprintf("%dx%d>", imgSize, imgSize), "-quality", "75",
					"-define", "PNG:compression-level=9", "-define", "PNG:compression-filter=0", outputFilepath)
				err = cmd.Run()
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("Error rescaling image %s", outputFilepath))
				}
			}

			break
		}

		if err != nil {
			return err
		}
	}
	return nil
}

// downloadImage downloads the given image (e.g. "File:foo.jpg") to the given folder.
// When the file already exists, nothing is done and "", nil will be returned.
// When the file has been downloaded "filename", nil will be returned.
func downloadImage(fileDescriptor string, outputFolder string, source string) (string, error) {
	filename := strings.Split(fileDescriptor, ":")[1]
	md5sum := fmt.Sprintf("%x", md5.Sum([]byte(filename)))
	sigolo.Debug(filename)
	sigolo.Debug(md5sum)

	url := fmt.Sprintf("https://upload.wikimedia.org/wikipedia/%s/%c/%c%c/%s", source, md5sum[0], md5sum[0], md5sum[1], filename)
	sigolo.Debug(url)

	return downloadAndCache(url, outputFolder, filename)
}

// downloadAndCache fires an GET request to the given url and saving the result in cacheFolder/filename. The return
// value is this resulting filepath or an error. If the file already exists, no HTTP request is made.
func downloadAndCache(url string, cacheFolder string, filename string) (string, error) {
	// If file exists -> ignore
	outputFilepath := filepath.Join(cacheFolder, filename)
	_, err := os.Stat(outputFilepath)
	if err == nil {
		sigolo.Info("File %s does already exist. Skip.", outputFilepath)
		return outputFilepath, nil
	}

	// Get the data
	var response *http.Response
	for {
		response, err = httpClient.Get(url)
		if err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("Unable to get file %s with url %s", filename, url))
		}

		// Handle 429 (too many requests): wait a bit and retry
		if response.StatusCode == 429 {
			time.Sleep(2 * time.Second)
			continue
		} else if response.StatusCode != 200 {
			return "", errors.Errorf("Downloading file %s failed with status code %d for url %s", filename, response.StatusCode, url)
		}

		break
	}
	defer response.Body.Close()
	reader := response.Body

	err = cacheToFile(cacheFolder, filename, reader)
	if err != nil {
		return "", errors.Wrapf(err, "Unable to cache to %s", outputFilepath)
	}

	return outputFilepath, nil
}

func cacheToFile(cacheFolder string, filename string, reader io.ReadCloser) error {
	// Create the output folder
	err := os.MkdirAll(cacheFolder, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, fmt.Sprintf("Unable to create output folder %s", cacheFolder))
	}

	outputFilepath := filepath.Join(cacheFolder, filename)

	// Create the output file
	outputFile, err := os.Create(outputFilepath)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to create output file for file %s", outputFilepath))
	}
	defer outputFile.Close()

	// Write the body to file
	_, err = io.Copy(outputFile, reader)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable copy downloaded content to output file %s", outputFilepath))
	}

	sigolo.Info("Cached file to %s", outputFilepath)

	return nil
}

func EvaluateTemplate(template string, cacheFolder string, cacheFile string) (string, error) {
	sigolo.Info("Evaluate template %s", template)

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
