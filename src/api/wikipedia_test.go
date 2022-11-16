package api

import (
	"github.com/hauke96/wiki2book/src/test"
	"testing"
)

func TestFollowRedirectIfNeeded(t *testing.T) {
	article := "TestFollowRedirectIfNeeded"
	content := "{\"parse\":{\"title\":\"foobar\",\"pageid\":123,\"wikitext\":{\"*\":\"blubb #REDIRECT [[other article]] whatever\"}}}"

	mockHttpClient := MockHttp(content, 200)

	redirectedArticle, err := followRedirectIfNeeded("commons", article, apiCacheFolder)

	test.AssertNil(t, err)
	test.AssertEqual(t, "other_article", redirectedArticle)
	test.AssertEqual(t, 1, mockHttpClient.GetCalls)
}

func TestFollowRedirectIfNeeded_noRedirect(t *testing.T) {
	article := "TestFollowRedirectIfNeeded_noRedirect"
	content := "{\"parse\":{\"title\":\"foobar\",\"pageid\":123,\"wikitext\":{\"*\":\"blubb #NO-REDIRECT [[other-article]] whatever\"}}}"

	mockHttpClient := MockHttp(content, 200)

	redirectedArticle, err := followRedirectIfNeeded("commons", article, apiCacheFolder)

	test.AssertNil(t, err)
	test.AssertEqual(t, article, redirectedArticle)
	test.AssertEqual(t, 1, mockHttpClient.GetCalls)
}
