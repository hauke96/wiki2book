package api

import (
	"github.com/hauke96/wiki2book/src/test"
	"os"
	"testing"
)

const apiCacheFolder = "../test/api-cache"

// Cleanup from previous runs
func cleanup(t *testing.T, key string) {
	err := os.Remove(apiCacheFolder + "/" + key)
	test.AssertTrue(t, err == nil || os.IsNotExist(err))
}

func TestDownloadAndCache(t *testing.T) {
	key := "foobar"
	content := "some interesting stuff"

	cleanup(t, key)
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

func TestFollowRedirectIfNeeded(t *testing.T) {
	article := "TestFollowRedirectIfNeeded"
	content := "{\"parse\":{\"title\":\"foobar\",\"pageid\":123,\"wikitext\":{\"*\":\"blubb #REDIRECT [[other article]] whatever\"}}}"

	cleanup(t, article+".json")
	mockHttpClient := MockHttp(content, 200)

	redirectedArticle, err := followRedirectIfNeeded("commons", article, apiCacheFolder)

	test.AssertNil(t, err)
	test.AssertEqual(t, "other_article", redirectedArticle)
	test.AssertEqual(t, 1, mockHttpClient.GetCalls)
}

func TestFollowRedirectIfNeeded_noRedirect(t *testing.T) {
	article := "TestFollowRedirectIfNeeded_noRedirect"
	content := "{\"parse\":{\"title\":\"foobar\",\"pageid\":123,\"wikitext\":{\"*\":\"blubb #NO-REDIRECT [[other-article]] whatever\"}}}"

	cleanup(t, article+".json")
	cleanup(t, article+".json")
	mockHttpClient := MockHttp(content, 200)

	redirectedArticle, err := followRedirectIfNeeded("commons", article, apiCacheFolder)

	test.AssertNil(t, err)
	test.AssertEqual(t, article, redirectedArticle)
	test.AssertEqual(t, 1, mockHttpClient.GetCalls)
}
