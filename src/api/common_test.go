package api

import (
	"github.com/hauke96/wiki2book/src/test"
	"testing"
)

const cacheSubFolder = "api-cache"

var apiCacheFolder = test.GetCacheFolder(cacheSubFolder)

func TestMain(m *testing.M) {
	test.CleanRun(m, cacheSubFolder)
}
