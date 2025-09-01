package http

import (
	"net/http"
	"os"
	"testing"
	"wiki2book/config"
	"wiki2book/test"
	"wiki2book/util"
)

func TestDownloadAndCache_withoutCachedFile(t *testing.T) {
	key := "foobar"
	content := "some interesting stuff"

	mockHttpClient := NewMockHttpClient(content, http.StatusOK)

	httpService := NewDefaultHttpService()
	httpService.httpClient = mockHttpClient

	fsMock := util.NewDefaultMockFilesystem()
	fsMock.StatFunc = func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	util.CurrentFilesystem = fsMock
	config.Current.CacheDir = test.TestCacheFolder

	cachedFilePath, freshlyDownloaded, err := httpService.DownloadAndCache("http://foobar", apiCacheFolder, key)

	test.AssertNil(t, err)
	test.AssertEqual(t, apiCacheFolder+"/"+key, cachedFilePath)
	test.AssertEqual(t, 1, mockHttpClient.GetCalls)
	test.AssertEqual(t, 0, mockHttpClient.PostCalls)
	test.AssertEqual(t, true, freshlyDownloaded)
}

func TestDownloadAndCache_withAlreadyCachedFile(t *testing.T) {
	key := "foobar"
	content := "some interesting stuff"

	mockHttpClient := NewMockHttpClient(content, http.StatusOK)

	httpService := NewDefaultHttpService()
	httpService.httpClient = mockHttpClient

	fsMock := util.NewDefaultMockFilesystem()
	fsMock.ReadFileFunc = func(name string) ([]byte, error) {
		return []byte(content), nil
	}
	util.CurrentFilesystem = fsMock

	config.Current.CacheDir = test.TestCacheFolder

	cachedFilePath, freshlyDownloaded, err := httpService.DownloadAndCache("http://foobar", apiCacheFolder, key)

	test.AssertNil(t, err)
	test.AssertEqual(t, apiCacheFolder+"/"+key, cachedFilePath)
	test.AssertEqual(t, 0, mockHttpClient.GetCalls)
	test.AssertEqual(t, 0, mockHttpClient.PostCalls)
	test.AssertEqual(t, false, freshlyDownloaded)
}
