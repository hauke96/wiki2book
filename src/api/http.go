package api

import (
	"io"
	"net/http"
)

type HttpClient interface {
	Get(url string) (resp *http.Response, err error)
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

func GetDefaultHttpClient() HttpClient {
	return http.DefaultClient
}
