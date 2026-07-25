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
	HeaderPromiseNotWrite   = "Promise-Non-Write-API-Action" // Suggested by the mediawiki documentation to use the nearest data center.
	HeaderRetryAfter        = "Retry-After"
	HeaderUserAgent         = "User-Agent"
	HeaderXResourceLocation = "x-resource-location"
)

var (
	sleepFunc = func(seconds int) { time.Sleep(time.Duration(seconds) * time.Second) }
)

// HttpClient is an interface compatible to http.Client from the stdlib. This enables abstraction, especially for tests.
type HttpClient interface {
	Do(request *http.Request) (resp *http.Response, err error)
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

// HttpService is an interface for higher level HTTP operations and additional functionality like caching.
type HttpService interface {
	PerformHttpRequest(url, method, contentType string) (resp *http.Response, err error)
	DownloadAndCache(url string, cacheFolder string, filename string) (string, bool, error)
	PostAndCache(url string, requestBody string, cacheFolder string, filename string) (string, bool, error)
}

// DefaultHttpService is the default implementation of the HttpService using the normal http.Client from the stdlib.
type DefaultHttpService struct {
	httpClient    HttpClient
	configService *config.ConfigService
	cache         *cache.Cache
}

func NewDefaultHttpService(configService *config.ConfigService, cache *cache.Cache) *DefaultHttpService {
	return &DefaultHttpService{
		httpClient:    &http.Client{},
		configService: configService,
		cache:         cache,
	}
}

// DownloadAndCache downloads the data of the given URL and returns the full output path, a flag indicating whether the
// file was downloaded and an error. In case the file is already cached, nothing is downloaded and the cached path
// together with "false" are returned.
func (d *DefaultHttpService) DownloadAndCache(url string, cacheFolderName string, filename string) (string, bool, error) {
	return d.performHttpRequestAndCache(url, cacheFolderName, filename, "GET", "")
}

// PostAndCache downloads the data of the given URL and returns the full output path, a flag indicating whether the
// file was downloaded and an error. In case the file is already cached, nothing is downloaded and the cached path
// together with "false" are returned.
func (d *DefaultHttpService) PostAndCache(url string, requestBody string, cacheFolderName string, filename string) (string, bool, error) {
	return d.performHttpRequestAndCache(url, cacheFolderName, filename, "POST", requestBody)
}

func (d *DefaultHttpService) performHttpRequestAndCache(url string, cacheFolderName string, filename string, method string, body string) (string, bool, error) {
	// If file already cached -> don't download and use cached file
	outputFilepath, fileIsCached, err := d.cache.GetFile(cacheFolderName, filename)
	if err == nil && fileIsCached {
		sigolo.Debugf("File '%s' does already exist -> use this cached file", outputFilepath)
		return outputFilepath, false, nil
	}
	if err != nil {
		return "", false, errors.Wrapf(err, "Unable to check whether file '%s' is already cached or not", outputFilepath)
	}
	sigolo.Debugf("File '%s' not cached -> download fresh one", outputFilepath)

	// Get the data
	response, err := d.PerformHttpRequest(url, method, body)
	if err != nil {
		return "", true, err
	}
	if response == nil || response.Body == nil {
		return "", true, errors.Errorf("Response or response body was nil from %s %s", method, url)
	}

	responseBodyReader := response.Body
	defer responseBodyReader.Close()

	outputFilepath, err = d.cache.CacheToFile(cacheFolderName, filename, responseBodyReader)
	if err != nil {
		return "", true, errors.Wrapf(err, "Unable to cache to '%s'", outputFilepath)
	}

	return outputFilepath, true, nil
}

// PerformHttpRequest send the given request to the given url. It only supports GET and POST requests. The requestBody
// must be filled with x-www-form-urlencoded data when using POST requests. It must be nil for GET requests. When the
// return error is set, the response is nil and vice versa.
func (d *DefaultHttpService) PerformHttpRequest(url, method, requestBody string) (*http.Response, error) {
	var response *http.Response
	var request *http.Request
	var err error

	sigolo.Debugf("Make %s request to %s with form data '%s'", method, url, util.TruncString(requestBody))

	if method != "GET" && method != "POST" {
		return nil, errors.Errorf("Unsupported request method %s", method)
	}
	if method == "GET" && requestBody != "" {
		return nil, errors.New("Request method GET does not support to have a request body")
	}
	if method == "POST" && requestBody == "" {
		return nil, errors.New("Request method POST must have a request body")
	}

	var requestBodyReader io.Reader
	if method == "POST" {
		requestBodyReader = strings.NewReader(requestBody)
	} else {
		requestBodyReader = nil
	}

	request, err = http.NewRequest(method, url, requestBodyReader)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Unable to create POST request for url %s", url))
	}

	userAgentString := d.configService.Get().UserAgentTemplate
	userAgentString = strings.ReplaceAll(userAgentString, "{{VERSION}}", util.VERSION)
	request.Header.Set(HeaderUserAgent, userAgentString)

	if method == "POST" {
		// Details: https://www.mediawiki.org/wiki/API:Etiquette#POST_requests
		request.Header.Set(HeaderContentType, "application/x-www-form-urlencoded")
		request.Header.Set(HeaderPromiseNotWrite, "true")
	}

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

	sigolo.Tracef("Make HTTP Request: %#v", request)

	for {
		response, err = d.httpClient.Do(request)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error executing %s request to url %s", request.Method, url))
		}

		sigolo.Tracef("Got HTTP Response: %#v", response)

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
