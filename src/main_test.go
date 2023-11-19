package main

import (
	"github.com/hauke96/sigolo"
	"testing"
	"wiki2book/config"
)

// This file is used to profile the application with the IntelliJ CPU profiler, which only works on test files.

func TestArticleGermanLong(t *testing.T) {
	err := config.LoadConfig("../configs/de.json")
	sigolo.FatalCheck(err)

	generateArticleEbook(
		"Commodore 128",
		"../.wiki2book/profiling.epub",
		"../.wiki2book",
		"",
		"",
		"../pandoc",
		true,
		true,
	)
}

func TestProjectGerman(t *testing.T) {
	err := config.LoadConfig("../configs/de.json")
	sigolo.FatalCheck(err)

	generateProjectEbook(
		"../projects/de/astronomie/astronomie.json",
		true,
		true,
	)
}