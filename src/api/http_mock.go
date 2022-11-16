package api

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
)

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
