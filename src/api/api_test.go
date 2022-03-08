package api

import (
	"bytes"
	"github.com/hauke96/wiki2book/src/test"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

const apiCacheFolder = "../test/api-cache"

// Cleanup from previous runs
func cleanup(t *testing.T, key string) {
	err := os.Remove(apiCacheFolder + "/" + key)
	test.AssertTrue(t, err == nil || os.IsNotExist(err))
}

type MockHttpClient struct {
	response   string
	statusCode int
	getCalls   int
	postCalls  int
}

func (h *MockHttpClient) Get(url string) (resp *http.Response, err error) {
	h.getCalls++
	return &http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(h.response))),
		StatusCode: h.statusCode,
	}, nil
}

func (h *MockHttpClient) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	h.postCalls++
	return &http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(h.response))),
		StatusCode: h.statusCode,
	}, nil
}

func mockHttp(response string, statusCode int) *MockHttpClient {
	mockedHttpClient := &MockHttpClient{
		response,
		statusCode,
		0,
		0,
	}
	httpClient = mockedHttpClient
	return mockedHttpClient
}

func TestDownloadAndCache(t *testing.T) {
	key := "foobar"
	content := "some interesting stuff"

	cleanup(t, key)
	mockHttpClient := mockHttp(content, 200)

	// First request -> cache file should ve created

	cachedFilePath, err := downloadAndCache("http://foobar", apiCacheFolder, key)

	test.AssertNil(t, err)
	test.AssertEqual(t, apiCacheFolder+"/"+key, cachedFilePath)
	test.AssertEqual(t, 1, mockHttpClient.getCalls)
	test.AssertEqual(t, 0, mockHttpClient.postCalls)

	// Second request -> nothing should change

	cachedFilePath, err = downloadAndCache("http://foobar", apiCacheFolder, key)

	test.AssertNil(t, err)
	test.AssertEqual(t, apiCacheFolder+"/"+key, cachedFilePath)
	test.AssertEqual(t, 1, mockHttpClient.getCalls)
	test.AssertEqual(t, 0, mockHttpClient.postCalls)
}
