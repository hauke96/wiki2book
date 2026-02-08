package generator

import (
	"testing"
	"wiki2book/config"

	"github.com/hauke96/sigolo/v2"
)

// This file is used to profile the application with the IntelliJ CPU profiler, which only works on test files.

func testArticleGermanLong(t *testing.T) {
	err := config.LoadConfig("../configs/de.json")
	sigolo.FatalCheck(err)

	GenerateArticleEbook(
		"Commodore 128",
		"../.wiki2book/profiling.epub",
	)
}

func testProjectGerman(t *testing.T) {
	err := config.LoadConfig("../configs/de.json")
	sigolo.FatalCheck(err)

	proj := CreateProject(
		"../projects/de/astronomie/astronomie.json",
		"../.wiki2book/profiling.epub",
		nil,
	)
	GenerateBookFromProject(proj)
}
