package generator

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"wiki2book/cache"
	"wiki2book/config"
	"wiki2book/http"
	"wiki2book/image"
	"wiki2book/parser"
	"wiki2book/util"
	"wiki2book/wikipedia"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

const (
	defaultEpubOutputFile      = "ebook.epub"
	defaultStatsJsonOutputFile = "stats.json"
	defaultStatsTxtOutputFile  = "stats.txt"
)

func CreateProject(projectFile string, outputFile string, cliConfig *config.Configuration) *config.Project {
	var err error

	sigolo.Infof("Use project file: '%s'", projectFile)

	directory, projectFile := filepath.Split(projectFile)
	if directory != "" {
		sigolo.Debugf("Go into folder '%s'", directory)
		err = os.Chdir(directory)
		sigolo.FatalCheck(err)
	}

	proj, err := config.LoadProject(projectFile)
	sigolo.FatalCheck(err)

	if outputFile != "" {
		sigolo.Tracef("Project has no output file set, so I'll use '%s'", outputFile)
		proj.OutputFile = outputFile
	}

	sigolo.Debug("Turn output file path into absolute path")
	proj.OutputFile, err = util.ToAbsolutePath(proj.OutputFile)
	sigolo.FatalCheck(err)

	config.MergeIntoCurrentConfig(&proj.Configuration)
	config.MergeIntoCurrentConfig(cliConfig)

	config.Current.Print()
	proj.Print()

	return proj
}

func GenerateStandaloneEbook(inputFile string, outputFile string) {
	config.Current.Print()

	fileContent, err := os.ReadFile(inputFile)
	sigolo.FatalCheck(err)

	_, inputFileName := path.Split(inputFile)
	title := strings.Split(inputFileName, ".")[0]

	outputFile = ensurePathsAndClearTempDir(outputFile)

	config.Current.AssertFilesAndPathsExists()

	wikipediaService := wikipedia.NewWikipediaService(
		config.Current.WikipediaInstance,
		config.Current.WikipediaHost,
		config.Current.WikipediaImageArticleHosts,
		config.Current.WikipediaImageHost,
		config.Current.WikipediaMathRestApi,
		image.NewImageProcessingService(),
		http.NewDefaultHttpService(),
	)

	tokenizer := parser.NewTokenizer(wikipediaService)
	article, err := tokenizer.Tokenize(string(fileContent), title)
	sigolo.FatalCheck(err)

	err = wikipediaService.DownloadImages(article.Images)
	sigolo.FatalCheck(err)

	// TODO Adjust this when additional non-epub output types are supported.
	htmlFilePath := path.Join(cache.HtmlCacheDirName, article.Title+".html")
	if shouldRecreateHtml(htmlFilePath, config.Current.ForceRegenerateHtml) {
		htmlGenerator := &HtmlGenerator{
			TokenMap:         article.TokenMap,
			WikipediaService: wikipediaService,
		}
		htmlFilePath, err = htmlGenerator.Generate(article)
		sigolo.FatalCheck(err)
	}

	sigolo.Infof("Start generating %s file", config.Current.OutputType)
	metadata := config.Metadata{
		Title: title,
	}

	err = GenerateEpub([]string{htmlFilePath}, outputFile, metadata)
	sigolo.FatalCheck(err)

	absoluteOutputFile, err := util.ToAbsolutePath(outputFile)
	sigolo.FatalCheck(err)
	sigolo.Infof("Successfully created %s file '%s'", config.Current.OutputType, absoluteOutputFile)
}

func GenerateArticleEbook(articleName string, outputFile string) {
	var articles []string
	articles = append(articles, articleName)

	proj := &config.Project{}
	proj.Metadata = config.Metadata{}
	proj.OutputFile = outputFile
	proj.Articles = articles

	config.Current.Print()

	GenerateBookFromProject(proj)
}

func GenerateBookFromProject(project *config.Project) {
	articles := project.Articles
	metadata := project.Metadata

	outputFile := ensurePathsAndClearTempDir(project.OutputFile)

	config.Current.AssertFilesAndPathsExists()

	numberOfArticles := len(articles)
	articleOutputFiles := make([]string, numberOfArticles)

	articleChan := make(chan string, config.Current.WorkerThreads)
	sigolo.Debugf("Use %d worker threads to process the articles", config.Current.WorkerThreads)

	wikipediaService := wikipedia.NewWikipediaService(
		config.Current.WikipediaInstance,
		config.Current.WikipediaHost,
		config.Current.WikipediaImageArticleHosts,
		config.Current.WikipediaImageHost,
		config.Current.WikipediaMathRestApi,
		image.NewImageProcessingService(),
		http.NewDefaultHttpService(),
	)

	// Create a wait-group that is zero when all threads are done
	threadPoolWaitGroup := &sync.WaitGroup{}
	threadPoolWaitGroup.Add(config.Current.WorkerThreads)

	// Start threads which pick an article from the channel to work on
	for i := 0; i < config.Current.WorkerThreads; i++ {
		sigolo.Debugf("Start worker thread %d", i)

		go func(threadNumber int) {
			for articleName := range articleChan {
				sigolo.Debugf("Processing article %s on thread %d", articleName, threadNumber)

				articleNumber := 0
				for ; articleNumber < len(articles); articleNumber++ {
					if articleName == articles[articleNumber] {
						break
					}
				}

				thisArticleOutputFile := processArticle(articleName, articleNumber+1, numberOfArticles, wikipediaService)
				articleOutputFiles[articleNumber] = thisArticleOutputFile
			}

			// This thread will close, so mark it as done in the wait-group
			sigolo.Debugf("Mark thread %d as done", threadNumber)
			threadPoolWaitGroup.Done()
		}(i)
	}

	for _, articleName := range articles {
		sigolo.Debugf("Pass article %s to worker threads", articleName)
		articleChan <- articleName
	}
	close(articleChan)

	// Wait for threads to be done with processing
	sigolo.Debugf("Wait for threads to finish processing articles ...")
	threadPoolWaitGroup.Wait()
	sigolo.Debugf("Worker threads are done processing articles")

	sigolo.Infof("Start generating %s file", config.Current.OutputType)
	switch config.Current.OutputType {
	case config.OutputTypeEpub2:
		fallthrough
	case config.OutputTypeEpub3:
		err := GenerateEpub(articleOutputFiles, outputFile, metadata)
		sigolo.FatalCheck(err)
	case config.OutputTypeStatsJson:
		fallthrough
	case config.OutputTypeStatsTxt:
		err := GenerateCombinedStats(articleOutputFiles, outputFile)
		sigolo.FatalCheck(err)
	}

	absoluteOutputFile, err := util.ToAbsolutePath(outputFile)
	sigolo.FatalCheck(err)
	sigolo.Infof("Successfully created %s file '%s'", config.Current.OutputType, absoluteOutputFile)
}

// processArticle processes a given article, which means, the content (including images etc.) is downloaded and the
// article will be tokenized, parsed and converted into the output format stored in the current configuration.
func processArticle(articleName string, currentArticleNumber int, totalNumberOfArticles int, wikipediaService *wikipedia.DefaultWikipediaService) string {
	sigolo.Infof("Article '%s' (%d/%d): Start processing", articleName, currentArticleNumber, totalNumberOfArticles)

	wikipediaArticleHost := fmt.Sprintf("%s.%s", config.Current.WikipediaInstance, config.Current.WikipediaHost)
	htmlFilePath := filepath.Join(cache.HtmlCacheDirName, articleName+".html") // TODO use generator to get this file (currently determining the filepath happens twice)
	articleOutputFile := ""
	if !shouldRecreateHtml(htmlFilePath, config.Current.ForceRegenerateHtml) {
		sigolo.Debugf("Article '%s' (%d/%d): HTML for article does already exist. Skip parsing and HTML generation.", articleName, currentArticleNumber, totalNumberOfArticles)
	} else {
		sigolo.Debugf("Article '%s' (%d/%d): Download article", articleName, currentArticleNumber, totalNumberOfArticles)
		wikiArticleDto, err := wikipediaService.DownloadArticle(wikipediaArticleHost, articleName)
		sigolo.FatalCheck(err)

		sigolo.Debugf("Article '%s' (%d/%d): Tokenize content", articleName, currentArticleNumber, totalNumberOfArticles)
		tokenizer := parser.NewTokenizer(wikipediaService)
		article, err := tokenizer.Tokenize(wikiArticleDto.Parse.Wikitext.Content, wikiArticleDto.Parse.OriginalTitle)
		sigolo.FatalCheck(err)

		sigolo.Debugf("Article '%s' (%d/%d): Download images", articleName, currentArticleNumber, totalNumberOfArticles)
		err = wikipediaService.DownloadImages(article.Images)
		sigolo.FatalCheck(err)

		switch config.Current.OutputType {
		case config.OutputTypeEpub2:
			fallthrough
		case config.OutputTypeEpub3:
			sigolo.Debugf("Article '%s' (%d/%d): Generate HTML", articleName, currentArticleNumber, totalNumberOfArticles)
			htmlGenerator := &HtmlGenerator{
				TokenMap:         article.TokenMap,
				WikipediaService: wikipediaService,
			}
			htmlFilePath, err = htmlGenerator.Generate(article)
			articleOutputFile = htmlFilePath
			sigolo.FatalCheck(err)
		case config.OutputTypeStatsJson:
			fallthrough
		case config.OutputTypeStatsTxt:
			sigolo.Debugf("Article '%s' (%d/%d): Generate stats", articleName, currentArticleNumber, totalNumberOfArticles)
			statsGenerator := NewStatsGenerator(article.TokenMap)
			articleOutputFile, err = statsGenerator.Generate(article)
			sigolo.FatalCheck(err)
		}
	}

	sigolo.Debugf("Article '%s' (%d/%d): Finished processing", articleName, currentArticleNumber, totalNumberOfArticles)

	return articleOutputFile
}

func shouldRecreateHtml(htmlFilePath string, forceHtmlRecreate bool) bool {
	if forceHtmlRecreate || config.Current.OutputType == config.OutputTypeStatsJson || config.Current.OutputType == config.OutputTypeStatsTxt {
		return true
	}

	// Check if HTML file already exists. If so, no recreate is wanted.
	_, err := os.Stat(htmlFilePath)
	htmlFileExists := err == nil

	return !htmlFileExists
}

// ensurePathsAndClearTempDir ensures that the output folder for the given outputFile exists and clears up any
// temporary files in the temp files folder that might still exist from previous runs.
func ensurePathsAndClearTempDir(outputFile string) string {
	var err error

	if config.Current.OutputType == config.OutputTypeStatsJson && !strings.HasSuffix(outputFile, "json") {
		// For stats, the output file is not an EPUB, therefore we change the default file in case it's the default EPUB one.
		outputFile = defaultStatsJsonOutputFile
		sigolo.Infof("Notice: Changing output file from default '%s' to '%s'", defaultEpubOutputFile, outputFile)
		outputFile, err = util.ToAbsolutePath(outputFile)
		sigolo.FatalCheck(err)
	}
	if config.Current.OutputType == config.OutputTypeStatsTxt && !strings.HasSuffix(outputFile, "txt") {
		// For stats, the output file is not an EPUB, therefore we change the default file in case it's the default EPUB one.
		outputFile = defaultStatsTxtOutputFile
		sigolo.Infof("Notice: Changing output file from default '%s' to '%s'", defaultEpubOutputFile, outputFile)
		outputFile, err = util.ToAbsolutePath(outputFile)
		sigolo.FatalCheck(err)
	}

	var file *os.File
	if _, err := os.Stat(outputFile); err != nil {
		// Output file does not exist
		sigolo.Debugf("Output file %s does not exists, I'll create it", outputFile)

		err = os.MkdirAll(filepath.Dir(outputFile), os.ModePerm)
		sigolo.FatalCheck(errors.Wrapf(err, "Error creating output file %s", outputFile))

		file, err = os.Create(outputFile)
		sigolo.FatalCheck(err)
	} else {
		file, err = os.Open(outputFile)
		sigolo.FatalCheck(err)
	}

	// Assign default output file name if given path is a directory
	fileInfo, err := file.Stat()
	sigolo.FatalCheck(err)

	if fileInfo.IsDir() {
		switch config.Current.OutputType {
		case config.OutputTypeEpub2:
			fallthrough
		case config.OutputTypeEpub3:
			outputFile = path.Join(outputFile, defaultEpubOutputFile)
		case config.OutputTypeStatsJson:
			outputFile = path.Join(outputFile, defaultStatsJsonOutputFile)
		case config.OutputTypeStatsTxt:
			outputFile = path.Join(outputFile, defaultStatsTxtOutputFile)
		}
	}

	// Make all relevant paths absolute
	outputFile, err = util.ToAbsolutePath(outputFile)
	sigolo.FatalCheck(err)

	sigolo.Debug("Ensure cache directories exist")
	util.EnsureDirectory(cache.GetTempPath())
	util.EnsureDirectory(cache.GetDirPathInCache(cache.ArticleCacheDirName))
	util.EnsureDirectory(cache.GetDirPathInCache(cache.HtmlCacheDirName))
	util.EnsureDirectory(cache.GetDirPathInCache(cache.ImageCacheDirName))
	util.EnsureDirectory(cache.GetDirPathInCache(cache.MathCacheDirName))
	util.EnsureDirectory(cache.GetDirPathInCache(cache.TemplateCacheDirName))
	util.EnsureDirectory(cache.GetDirPathInCache(cache.TempDirName))

	return outputFile
}
