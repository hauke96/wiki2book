package api

import (
	"testing"
	"wiki2book/test"
)

const cacheSubFolder = "api-cache"

var apiCacheFolder = test.GetCacheFolder(cacheSubFolder)

func TestMain(m *testing.M) {
	test.CleanRun(m, cacheSubFolder)
}
