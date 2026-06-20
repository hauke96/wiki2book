package http

import (
	"bytes"
	"io"
	"net/http"
)

type mockHttpClient struct {
	Response   string
	StatusCode int
	GetCalls   int
	PostCalls  int
	doFunc     func(request *http.Request) (resp *http.Response, err error)
	postFunc   func(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

func NewMockHttpClient(response string, statusCode int) *mockHttpClient {
	return &mockHttpClient{
		response,
		statusCode,
		0,
		0,
		func(request *http.Request) (resp *http.Response, err error) { return nil, nil },
		func(url, contentType string, body io.Reader) (resp *http.Response, err error) { return nil, nil },
	}
}

func (h *mockHttpClient) Do(request *http.Request) (resp *http.Response, err error) {
	if request.Method == "GET" {
		h.GetCalls++
	} else if request.Method == "POST" {
		h.PostCalls++
	}

	response, err := h.doFunc(request)
	if response != nil || err != nil {
		return response, err
	}

	// Deprecated:
	return &http.Response{
		Body:       io.NopCloser(bytes.NewReader([]byte(h.Response))),
		StatusCode: h.StatusCode,
	}, nil
}

func (h *mockHttpClient) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	h.PostCalls++

	response, err := h.postFunc(url, contentType, body)
	if response != nil || err != nil {
		return response, err
	}

	// Deprecated:

	return &http.Response{
		Body:       io.NopCloser(bytes.NewReader([]byte(h.Response))),
		StatusCode: h.StatusCode,
	}, nil
}

type mockHttpService struct {
	DownloadAndCacheCounter   int
	DownloadAndCacheFunc      func(url string, cacheFolder string, filename string) (string, bool, error)
	PostAndCacheCounter       int
	PostAndCacheFunc          func(url string, requestBody string, cacheFolder string, filename string) (string, bool, error)
	PerformHttpRequestCounter int
	PerformHttpRequestFunc    func(url, method, contentType string) (resp *http.Response, err error)
}

func (h *mockHttpService) DownloadAndCache(url string, cacheFolder string, filename string) (string, bool, error) {
	h.DownloadAndCacheCounter++
	return h.DownloadAndCacheFunc(url, cacheFolder, filename)
}

func (h *mockHttpService) PostAndCache(url string, requestBody string, cacheFolder string, filename string) (string, bool, error) {
	h.PostAndCacheCounter++
	return h.PostAndCacheFunc(url, requestBody, cacheFolder, filename)
}

func (h *mockHttpService) PerformHttpRequest(url string, method string, contentType string) (*http.Response, error) {
	h.PerformHttpRequestCounter++
	return h.PerformHttpRequestFunc(url, method, contentType)
}

func NewMockHttpService(
	downloadAndCacheFunc func(url string, cacheFolder string, filename string) (string, bool, error),
	postAndCacheFunc func(url string, requestBody string, cacheFolder string, filename string) (string, bool, error),
	performHttpRequestFunc func(url string, method, contentType string) (*http.Response, error),
) *mockHttpService {
	mockedHttpClient := &mockHttpService{
		DownloadAndCacheCounter:   0,
		DownloadAndCacheFunc:      downloadAndCacheFunc,
		PostAndCacheCounter:       0,
		PostAndCacheFunc:          postAndCacheFunc,
		PerformHttpRequestCounter: 0,
		PerformHttpRequestFunc:    performHttpRequestFunc,
	}
	return mockedHttpClient
}
