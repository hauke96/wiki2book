package http

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
	"wiki2book/config"
	"wiki2book/test"
	"wiki2book/util"

	"github.com/pkg/errors"
)

func TestDownloadAndCache_withoutCachedFile(t *testing.T) {
	// Arrange
	key := "foobar"
	content := "some interesting stuff"

	mockHttpClient := NewMockHttpClient(content, http.StatusOK)

	httpService := NewDefaultHttpService()
	httpService.httpClient = mockHttpClient

	fsMock := util.NewDefaultMockFilesystem()
	fsMock.StatFunc = func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	util.CurrentFilesystem = fsMock
	config.Current.CacheDir = test.TestCacheFolder

	// Act
	cachedFilePath, freshlyDownloaded, err := httpService.DownloadAndCache("http://foobar", apiCacheFolder, key)

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, apiCacheFolder+"/"+key, cachedFilePath)
	test.AssertEqual(t, 1, mockHttpClient.GetCalls)
	test.AssertEqual(t, 0, mockHttpClient.PostCalls)
	test.AssertEqual(t, true, freshlyDownloaded)
}

func TestDownloadAndCache_withAlreadyCachedFile(t *testing.T) {
	// Arrange
	key := "foobar"
	content := "some interesting stuff"

	mockHttpClient := NewMockHttpClient(content, http.StatusOK)

	httpService := NewDefaultHttpService()
	httpService.httpClient = mockHttpClient

	config.Current.CacheDir = test.TestCacheFolder
	config.Current.CacheMaxAge = 9999999

	fsMock := util.NewDefaultMockFilesystem()
	fsMock.ReadFileFunc = func(name string) ([]byte, error) {
		return []byte(content), nil
	}
	fsMock.StatFunc = func(path string) (os.FileInfo, error) {
		return util.NewMockFileInfoWithTime("file", time.Now()), nil
	}
	util.CurrentFilesystem = fsMock

	// Act
	cachedFilePath, freshlyDownloaded, err := httpService.DownloadAndCache("http://foobar", apiCacheFolder, key)

	// Assert
	test.AssertNil(t, err)
	test.AssertEqual(t, apiCacheFolder+"/"+key, cachedFilePath)
	test.AssertEqual(t, 0, mockHttpClient.GetCalls)
	test.AssertEqual(t, 0, mockHttpClient.PostCalls)
	test.AssertEqual(t, false, freshlyDownloaded)
}

func TestDownloadAndCache_tooManyRequestsResponse(t *testing.T) {
	// Arrange
	expectedSleepCallParam := 123
	var sleepFuncCallParams []int
	sleepFunc = func(seconds int) {
		sleepFuncCallParams = append(sleepFuncCallParams, seconds)
	}

	doCall := 0
	mockHttpClient := NewMockHttpClient("", http.StatusOK)
	mockHttpClient.doFunc = func(req *http.Request) (*http.Response, error) {
		doCall++
		if doCall == 1 {
			return &http.Response{
				Body:       io.NopCloser(bytes.NewReader([]byte("response of call1"))),
				StatusCode: http.StatusTooManyRequests,
				Header:     map[string][]string{HeaderRetryAfter: {"some invalid number"}},
			}, nil
		} else if doCall == 2 {
			return &http.Response{
				Body:       io.NopCloser(bytes.NewReader([]byte("response of call2"))),
				StatusCode: http.StatusTooManyRequests,
				Header:     map[string][]string{HeaderRetryAfter: {strconv.Itoa(expectedSleepCallParam)}},
			}, nil
		} else if doCall == 3 {
			return &http.Response{
				Body:       io.NopCloser(bytes.NewReader([]byte("response of call3"))),
				StatusCode: http.StatusOK,
			}, nil
		}
		return nil, errors.New("no more than 3 requests expected")
	}

	httpService := NewDefaultHttpService()
	httpService.httpClient = mockHttpClient

	// Act
	reader, err := httpService.download("http://foobar", "testfile")

	// Assert
	test.AssertNil(t, err)

	all, err := io.ReadAll(reader)
	test.AssertNil(t, err)
	test.AssertEqual(t, "response of call3", string(all))

	test.AssertEqual(t, []int{2, expectedSleepCallParam}, sleepFuncCallParams)
}
