package http

import (
	"net/http"
	"testing"
	"wiki2book/test"
)

func TestDownloadAndCache(t *testing.T) {
	key := "foobar"
	content := "some interesting stuff"

	mockHttpClient := NewMockHttpClient(content, http.StatusOK)

	httpService := NewDefaultHttpService()
	httpService.httpClient = mockHttpClient

	// First request -> cache file should ve created

	cachedFilePath, freshlyDownloaded, err := httpService.DownloadAndCache("http://foobar", apiCacheFolder, key)

	test.AssertNil(t, err)
	test.AssertEqual(t, apiCacheFolder+"/"+key, cachedFilePath)
	test.AssertEqual(t, 1, mockHttpClient.GetCalls)
	test.AssertEqual(t, 0, mockHttpClient.PostCalls)
	test.AssertEqual(t, true, freshlyDownloaded)

	// Second request -> nothing should change

	cachedFilePath, freshlyDownloaded, err = httpService.DownloadAndCache("http://foobar", apiCacheFolder, key)

	test.AssertNil(t, err)
	test.AssertEqual(t, apiCacheFolder+"/"+key, cachedFilePath)
	test.AssertEqual(t, 1, mockHttpClient.GetCalls)
	test.AssertEqual(t, 0, mockHttpClient.PostCalls)
	test.AssertEqual(t, false, freshlyDownloaded)
}
