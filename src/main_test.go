package main

import (
	"github.com/hauke96/sigolo/v2"
	"testing"
	"wiki2book/config"
)

// This file is used to profile the application with the IntelliJ CPU profiler, which only works on test files.

func testArticleGermanLong(t *testing.T) {
	err := config.LoadConfig("../configs/de.json")
	sigolo.FatalCheck(err)

	generateArticleEbook(
		"Commodore 128",
		"../.wiki2book/profiling.epub",
		"epub2",
		"pandoc",
		"../.wiki2book",
		"",
		"",
		"../pandoc",
		[]string{},
		true,
		true,
		true,
	)
}

func testProjectGerman(t *testing.T) {
	err := config.LoadConfig("../configs/de.json")
	sigolo.FatalCheck(err)

	generateProjectEbook(
		"../projects/de/astronomie/astronomie.json",
		"../.wiki2book/profiling.epub",
		"epub2",
		"pandoc",
		"../.wiki2book",
		"",
		"",
		"../pandoc",
		[]string{},
		true,
		true,
		true,
	)
}
