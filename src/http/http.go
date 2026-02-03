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
	responseBodyReader, err := d.download(url)
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

// download returns the open response body of the GET request for the given URL. The article name is just there for
// logging purposes.
func (d *DefaultHttpService) download(url string) (io.ReadCloser, error) {
	var response *http.Response
	var request *http.Request
	var err error

	sigolo.Debugf("Make GET request to %s", url)
	request, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Unable to create GET request for url %s", url))
	}

	userAgentString := config.Current.UserAgentTemplate
	userAgentString = strings.ReplaceAll(userAgentString, "{{VERSION}}", util.VERSION)
	request.Header.Set(HeaderUserAgent, userAgentString)

	response, err = d.doRequest(url, request)
	if err != nil {
		return nil, err
	}
	sigolo.Tracef("Response: %#v", response)
	return response.Body, nil
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
	response, err = d.doRequest(url, request)
	if err != nil {
		return nil, err
	}
	sigolo.Tracef("Response: %#v", response)
	return response, nil
}

func (d *DefaultHttpService) doRequest(url string, request *http.Request) (*http.Response, error) {
	var response *http.Response
	var err error

	for {
		response, err = d.httpClient.Do(request)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error executing %s request to url %s", request.Method, url))
		}

		sigolo.Tracef("Response: %#v", response)

		if response.StatusCode == http.StatusTooManyRequests {
			// 429 (too many requests): wait a bit and retry
			var waitTime int
			waitTime, err = strconv.Atoi(response.Header.Get(HeaderRetryAfter))
			if err != nil {
				waitTime = 2
				sigolo.Warnf("Unable to parse '%s' header value after receiving HTTP status code %d. Instead I'll wait %d seconds, but this might lead to recurring HTTP errors.", HeaderRetryAfter, http.StatusTooManyRequests, waitTime)
			}
			sigolo.Debugf("Received response status code %d and try request again in %d seconds", response.StatusCode, waitTime)
			sleepFunc(waitTime)
			continue
		} else if response.StatusCode != http.StatusOK {
			return nil, errors.Errorf("%s request to url %s failed with status code %d", request.Method, url, response.StatusCode)
		}

		responseErrorHeader := response.Header.Get(HeaderMediawikiApiError)
		if responseErrorHeader != "" {
			return nil, errors.Errorf("%s request to url %s failed with error header '%s' value '%s'", request.Method, url, HeaderMediawikiApiError, responseErrorHeader)
		}
		break
	}

	return response, nil
}
