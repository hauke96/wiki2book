package api

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

func (h *mockHttpClient) Get(url string) (*http.Response, error) {
	h.GetCalls++
	return &http.Response{
		Body:       io.NopCloser(bytes.NewReader([]byte(h.Response))),
		StatusCode: h.StatusCode,
	}, nil
}

func (h *mockHttpClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	h.PostCalls++
	return &http.Response{
		Body:       io.NopCloser(bytes.NewReader([]byte(h.Response))),
		StatusCode: h.StatusCode,
	}, nil
}

func NewMockHttp(response string, statusCode int) *mockHttpClient {
	mockedHttpClient := &mockHttpClient{
		response,
		statusCode,
		0,
		0,
	}
	httpClient = mockedHttpClient
	return mockedHttpClient
}
