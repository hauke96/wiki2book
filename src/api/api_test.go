package api

import (
	"github.com/hauke96/wiki2book/src/test"
	"testing"
)

func TestDownloadAndCache(t *testing.T) {
	key := "foobar"
	content := "some interesting stuff"

	mockHttpClient := MockHttp(content, 200)

	// First request -> cache file should ve created

	cachedFilePath, err := downloadAndCache("http://foobar", apiCacheFolder, key)

	test.AssertNil(t, err)
	test.AssertEqual(t, apiCacheFolder+"/"+key, cachedFilePath)
	test.AssertEqual(t, 1, mockHttpClient.GetCalls)
	test.AssertEqual(t, 0, mockHttpClient.PostCalls)

	// Second request -> nothing should change

	cachedFilePath, err = downloadAndCache("http://foobar", apiCacheFolder, key)

	test.AssertNil(t, err)
	test.AssertEqual(t, apiCacheFolder+"/"+key, cachedFilePath)
	test.AssertEqual(t, 1, mockHttpClient.GetCalls)
	test.AssertEqual(t, 0, mockHttpClient.PostCalls)
}
