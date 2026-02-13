package generator

import (
	"testing"
	"wiki2book/cache"
	"wiki2book/config"

	"github.com/hauke96/sigolo/v2"
)

// This file is used to profile the application with the IntelliJ CPU profiler, which only works on test files.

func testArticleGermanLong(t *testing.T) {
	configService := config.NewConfigService()
	err := configService.LoadFromConfig("../configs/de.json")
	sigolo.FatalCheck(err)
	fileCache := cache.NewCache(configService)
	ebookGeneratorService := NewEbookGenerator(configService, fileCache)

	ebookGeneratorService.GenerateArticleEbook(
		"Commodore 128",
		"../.wiki2book/profiling.epub",
	)
}

func testProjectGerman(t *testing.T) {
	configService := config.NewConfigService()
	err := configService.LoadFromConfig("../configs/de.json")
	sigolo.FatalCheck(err)
	fileCache := cache.NewCache(configService)
	ebookGeneratorService := NewEbookGenerator(configService, fileCache)

	proj := ebookGeneratorService.CreateProject(
		"../projects/de/astronomie/astronomie.json",
		"../.wiki2book/profiling.epub",
		nil,
	)
	ebookGeneratorService.GenerateBookFromProject(proj)
}
