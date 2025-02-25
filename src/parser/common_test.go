package parser

import (
	"testing"
	"wiki2book/config"
	"wiki2book/test"
)

func TestMain(m *testing.M) {
	test.CleanRun(m)
}

func setup() {
	images = []string{}
	config.Current.ConvertPDFsToImages = false
}
