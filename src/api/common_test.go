package api

import (
	"testing"
	"wiki2book/test"
)

var apiCacheFolder = test.GetCacheFolder()

func TestMain(m *testing.M) {
	test.CleanRun(m)
}
