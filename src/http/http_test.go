package http

import (
	"net/http"
	"testing"
	"wiki2book/test"
)

func TestDownloadAndCache(t *testing.T) {
	key := "foobar"
	content := "some interesting stuff"

	mockHttpClient := NewMockHttp(content, http.StatusOK)

	// First request -> cache file should ve created

	cachedFilePath, freshlyDownloaded, err := DownloadAndCache("http://foobar", apiCacheFolder, key)

	test.AssertNil(t, err)
	test.AssertEqual(t, apiCacheFolder+"/"+key, cachedFilePath)
	test.AssertEqual(t, 1, mockHttpClient.GetCalls)
	test.AssertEqual(t, 0, mockHttpClient.PostCalls)
	test.AssertEqual(t, true, freshlyDownloaded)

	// Second request -> nothing should change

	cachedFilePath, freshlyDownloaded, err = DownloadAndCache("http://foobar", apiCacheFolder, key)

	test.AssertNil(t, err)
	test.AssertEqual(t, apiCacheFolder+"/"+key, cachedFilePath)
	test.AssertEqual(t, 1, mockHttpClient.GetCalls)
	test.AssertEqual(t, 0, mockHttpClient.PostCalls)
	test.AssertEqual(t, false, freshlyDownloaded)
}
