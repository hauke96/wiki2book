package api

import (
	"bytes"
	"io"
	"net/http"
)

type MockHttpClient struct {
	Response   string
	StatusCode int
	GetCalls   int
	PostCalls  int
}

func (h *MockHttpClient) Get(url string) (*http.Response, error) {
	h.GetCalls++
	return &http.Response{
		Body:       io.NopCloser(bytes.NewReader([]byte(h.Response))),
		StatusCode: h.StatusCode,
	}, nil
}

func (h *MockHttpClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	h.PostCalls++
	return &http.Response{
		Body:       io.NopCloser(bytes.NewReader([]byte(h.Response))),
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
