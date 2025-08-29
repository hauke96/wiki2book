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
}

func NewMockHttpClient(response string, statusCode int) *mockHttpClient {
	return &mockHttpClient{
		response,
		statusCode,
		0,
		0,
	}
}

func (h *mockHttpClient) Do(request *http.Request) (resp *http.Response, err error) {
	if request.Method == "GET" {
		h.GetCalls++
	} else if request.Method == "POST" {
		h.PostCalls++
	}
	return &http.Response{
		Body:       io.NopCloser(bytes.NewReader([]byte(h.Response))),
		StatusCode: h.StatusCode,
	}, nil
}

func (h *mockHttpClient) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	h.PostCalls++
	return &http.Response{
		Body:       io.NopCloser(bytes.NewReader([]byte(h.Response))),
		StatusCode: h.StatusCode,
	}, nil
}

type mockHttpService struct {
	DownloadAndCacheCounter int
	DownloadAndCacheFunc    func(url string, cacheFolder string, filename string) (string, bool, error)
	PostFormEncodedCounter  int
	PostFormEncodedFunc     func(url, contentType string) (resp *http.Response, err error)
}

func (h *mockHttpService) DownloadAndCache(url string, cacheFolder string, filename string) (string, bool, error) {
	h.DownloadAndCacheCounter++
	return h.DownloadAndCacheFunc(url, cacheFolder, filename)
}

func (h *mockHttpService) PostFormEncoded(url, contentType string) (resp *http.Response, err error) {
	h.PostFormEncodedCounter++
	return h.PostFormEncodedFunc(url, contentType)
}

func NewMockHttpService(
	downloadAndCacheFunc func(url string, cacheFolder string, filename string) (string, bool, error),
	postFormEncodedFunc func(url, contentType string) (resp *http.Response, err error),
) *mockHttpService {
	mockedHttpClient := &mockHttpService{
		DownloadAndCacheCounter: 0,
		DownloadAndCacheFunc:    downloadAndCacheFunc,
		PostFormEncodedCounter:  0,
		PostFormEncodedFunc:     postFormEncodedFunc,
	}
	return mockedHttpClient
}
