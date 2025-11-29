package http

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"wiki2book/cache"
	"wiki2book/config"
	"wiki2book/util"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

const (
	HeaderContentType       = "Content-Type"
	HeaderMediawikiApiError = "mediawiki-api-error"
	HeaderRetryAfter        = "Retry-After"
	HeaderUserAgent         = "User-Agent"
	HeaderXResourceLocation = "x-resource-location"
)

var (
	sleepFunc = func(seconds int) { time.Sleep(time.Duration(seconds) * time.Second) }
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

// DownloadAndCache downloads the data of the given URL and returns the full output path, a flag indicating whether the
// file was downloaded and an error. In case the file is already cached, nothing is downloaded and the cached path
// together with "false" are returned.
func (d *DefaultHttpService) DownloadAndCache(url string, cacheFolderName string, filename string) (string, bool, error) {
	// If file already cached -> don't download and use cached file
	outputFilepath, fileIsCached, err := cache.GetFile(cacheFolderName, filename)
	if err == nil && fileIsCached {
		sigolo.Debugf("File '%s' does already exist -> use this cached file", outputFilepath)
		return outputFilepath, false, nil
	}
	if err != nil {
		return "", false, errors.Wrapf(err, "Unable to check whether file '%s' is already cached or not", outputFilepath)
	}
	sigolo.Debugf("File '%s' not cached -> download fresh one", outputFilepath)

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

	outputFilepath, err = cache.CacheToFile(cacheFolderName, filename, responseBodyReader)
	if err != nil {
		return "", true, errors.Wrapf(err, "Unable to cache to '%s'", outputFilepath)
	}

	return outputFilepath, true, nil
}

func (d *DefaultHttpService) PostFormEncoded(url, requestData string) (resp *http.Response, err error) {
	sigolo.Debugf("Make POST request to %s with form data '%s'", url, util.TruncString(requestData))
	request, err := http.NewRequest("POST", url, strings.NewReader(requestData))
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Unable to create POST request for url %s", url))
	}

	userAgentString := config.Current.UserAgentTemplate
	userAgentString = strings.ReplaceAll(userAgentString, "{{VERSION}}", util.VERSION)
	request.Header.Set(HeaderUserAgent, userAgentString)
	request.Header.Set(HeaderContentType, "application/x-www-form-urlencoded")

	var response *http.Response
	response, err = d.httpClient.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error executing POST request to url %s", url))
	}

	sigolo.Tracef("Response: %#v", response)
	return response, nil
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
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to create GET request for url %s to download file '%s'", url, filename))
		}

		userAgentString := config.Current.UserAgentTemplate
		userAgentString = strings.ReplaceAll(userAgentString, "{{VERSION}}", util.VERSION)
		request.Header.Set(HeaderUserAgent, userAgentString)

		response, err = d.httpClient.Do(request)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error executing GET request to url %s", url))
		}

		sigolo.Tracef("Response: %#v", response)

		if response.StatusCode == http.StatusTooManyRequests {
			// 429 (too many requests): wait a bit and retry
			sigolo.Debugf("Received responst status code %d and try request again", response.StatusCode)
			var waitTime int
			waitTime, err = strconv.Atoi(response.Header.Get(HeaderRetryAfter))
			if err != nil {
				waitTime = 2
				sigolo.Warnf("Unable to parse '%s' header value after receiving HTTP status code %d. Instead I'll wait %d seconds, but this might lead to recurring HTTP errors.", HeaderRetryAfter, http.StatusTooManyRequests, waitTime)
			}
			sigolo.Tracef("Wait %d seconds and try again", waitTime)
			sleepFunc(waitTime)
			continue
		} else if response.StatusCode != http.StatusOK {
			return response.Body, errors.Errorf("Downloading file '%s' failed with status code %d for url %s", filename, response.StatusCode, url)
		} else {
			responseErrorHeader := response.Header.Get(HeaderMediawikiApiError)
			if responseErrorHeader != "" {
				return response.Body, errors.Errorf("Downloading file '%s' failed with error header '%s' value '%s' for url %s", filename, HeaderMediawikiApiError, responseErrorHeader, url)
			}
		}

		break
	}
	return response.Body, nil
}
