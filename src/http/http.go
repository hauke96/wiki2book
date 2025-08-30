package http

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"wiki2book/cache"
	"wiki2book/config"
	"wiki2book/util"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

type HttpClient interface {
	Do(request *http.Request) (resp *http.Response, err error)
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

type HttpService interface {
	PostFormEncoded(url, contentType string) (resp *http.Response, err error)
	DownloadAndCache(url string, cacheFolder string, filename string) (string, bool, error)
}

type DefaultHttpService struct {
	httpClient HttpClient
}

func NewDefaultHttpService() *DefaultHttpService {
	return &DefaultHttpService{
		httpClient: &http.Client{},
	}
}

// DownloadAndCache fires an GET request to the given url and saving the result in cacheFolder/filename. The return
// value is this resulting filepath and a bool (true = file was (tried to be) downloaded, false = file already exists in
// cache) or an error. If the file already exists, no HTTP request is made.
func (d *DefaultHttpService) DownloadAndCache(url string, cacheFolderName string, filename string) (string, bool, error) {
	// If file exists -> ignore
	// TODO extract to something like "cache.GetFile(cacheFolderName, filename)" returning nil or the file path
	outputFilepath := filepath.Join(config.Current.CacheDir, cacheFolderName, filename)
	_, err := os.Stat(outputFilepath)
	if err == nil {
		sigolo.Debugf("File %s does already exist -> use this cached file", outputFilepath)
		return outputFilepath, false, nil
	}
	sigolo.Debugf("File %s not cached -> download fresh one", outputFilepath)

	// Get the data
	responseBodyReader, err := d.download(url, filename)
	if responseBodyReader != nil {
		defer responseBodyReader.Close()
		if err != nil {
			responseBodyText := util.ReaderToString(responseBodyReader)
			sigolo.Errorf("Response body of failed request:\n%s", responseBodyText)
			return "", true, err
		}
	}
	if err != nil {
		return "", true, err
	}

	err = cache.CacheToFile(cacheFolderName, filename, responseBodyReader)
	if err != nil {
		return "", true, errors.Wrapf(err, "Unable to cache to %s", outputFilepath)
	}

	return outputFilepath, true, nil
}

func (d *DefaultHttpService) PostFormEncoded(url, requestData string) (resp *http.Response, err error) {
	return d.httpClient.Post(url, "application/x-www-form-urlencoded", strings.NewReader(requestData))
}

// download returns the open response body of the GET request for the given URL. The article name is just there for
// logging purposes.
func (d *DefaultHttpService) download(url string, filename string) (io.ReadCloser, error) {
	var response *http.Response
	var request *http.Request
	var err error

	for {
		sigolo.Debugf("Make GET request to %s", url)
		request, err = http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to creqte GET request for url %s to download file %s", url, filename))
		}

		request.Header.Set("User-Agent", fmt.Sprintf("wiki2book %s (https://github.com/hauke96/wiki2book)", util.VERSION))

		response, err = d.httpClient.Do(request)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to get file %s with url %s", filename, url))
		}

		sigolo.Tracef("Response: %#v", response)

		// Handle 429 (too many requests): wait a bit and retry
		if response.StatusCode == http.StatusTooManyRequests {
			sigolo.Tracef("Got %d response (too many requests). Wait some time and try again...", http.StatusTooManyRequests)
			time.Sleep(2 * time.Second)
			continue
		} else if response.StatusCode != http.StatusOK {
			return response.Body, errors.Errorf("Downloading file '%s' failed with status code %d for url %s", filename, response.StatusCode, url)
		} else {
			errorHeaderName := "mediawiki-api-error"
			responseErrorHeader := response.Header.Get(errorHeaderName)
			if responseErrorHeader != "" {
				return response.Body, errors.Errorf("Downloading file '%s' failed with error header '%s' value '%s' for url %s", filename, errorHeaderName, responseErrorHeader, url)
			}
		}

		break
	}
	return response.Body, nil
}
