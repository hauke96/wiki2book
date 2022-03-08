package api

import (
	"io"
	"net/http"
)

type HttpClient interface {
	Get(url string) (resp *http.Response, err error)
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

type DefaultHttpClient struct {
}

func (h *DefaultHttpClient) Get(url string) (resp *http.Response, err error) {
	return http.Get(url)
}

func (h *DefaultHttpClient) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	return http.Post(url, contentType, body)
}

func GetDefaultHttpClient() HttpClient {
	return &DefaultHttpClient{}
}
