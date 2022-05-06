package main

import (
	"github.com/alecthomas/kong"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/api"
	"github.com/hauke96/wiki2book/src/generator/epub"
	"github.com/hauke96/wiki2book/src/generator/html"
	"github.com/hauke96/wiki2book/src/parser"
	"github.com/hauke96/wiki2book/src/project"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

var cli struct {
	Standalone struct {
		File      string `help:"A mediawiki file tha should be rendered to an eBook." type:"existingfile:" arg:""`
		OutputDir string `help:"The directory where all the files should be put into." short:"o" type:"path:"`
	} `cmd:"" help:"Renders a single mediawiki file into an eBook."`
	Project struct {
		ProjectFile string `help:"A project JSON-file tha should be used to create an eBook." type:"existingfile:" arg:""`
	} `cmd:"" help:"Uses a project file to create the eBook."`
}

func main() {
	ctx := kong.Parse(&cli)

	switch ctx.Command() {
	case "standalone":
	case "project <project-file>":
		generateEbook(cli.Project.ProjectFile)
	default:
		sigolo.Fatal("Unknown command: %v\n%#v", ctx.Command(), ctx)
	}
}

func generateEbook(projectFile string) {
	var err error
	start := time.Now()

	// Enable this to create a profiling file. Then use the command "go tool pprof src ./profiling.prof" and enter "web" to open a diagram in your browser.
	//f, err := os.Create("profiling.prof")
	//sigolo.FatalCheck(err)
	//
	//err = pprof.StartCPUProfile(f)
	//sigolo.FatalCheck(err)
	//defer pprof.StopCPUProfile()

	if "test" == projectFile {
		sigolo.Info("Use test file instead of real project file")
		generateTestEbook()
		os.Exit(0)
	}

	sigolo.Info("Use project file: %s", projectFile)

	directory, _ := filepath.Split(projectFile)
	err = os.Chdir(directory)
	sigolo.FatalCheck(err)

	project, err := project.LoadProject(projectFile)
	sigolo.FatalCheck(err)

	var articleFiles []string

	for _, articleName := range project.Articles {
		sigolo.Info("Start processing articleName %s", articleName)

		wikiArticleDto, err := api.DownloadArticle(project.Domain, articleName, project.Caches.Articles)
		sigolo.FatalCheck(err)

		tokenizer := parser.NewTokenizer(project.Caches.Images, project.Caches.Templates)
		article := parser.Parse(wikiArticleDto.Parse.Wikitext.Content, wikiArticleDto.Parse.Title, &tokenizer)

		outputFile, err := html.Generate(article, "./", project.Style, project.Caches.Images, project.Caches.Math)
		sigolo.FatalCheck(err)

		articleFiles = append(articleFiles, outputFile)

		sigolo.Info("Succeesfully created HTML for articleName %s", articleName)
	}

	sigolo.Info("Start generating EPUB file")
	err = epub.Generate(articleFiles, project.OutputFile, project.Style, project.Cover, project.Metadata)
	sigolo.FatalCheck(err)
	sigolo.Info("Successfully created EPUB file")

	end := time.Now()
	sigolo.Debug("Start   : %s", start.Format(time.RFC1123))
	sigolo.Debug("End     : %s", end.Format(time.RFC1123))
	sigolo.Debug("Duration: %f seconds", end.Sub(start).Seconds())
}

func generateTestEbook() {
	imageFolder := "../test/images"
	mathFolder := "../test/math"
	templateFolder := "../test/templates"

	sigolo.LogLevel = sigolo.LOG_DEBUG

	fileContent, err := ioutil.ReadFile("../test/test.mediawiki")
	sigolo.FatalCheck(err)

	tokenizer := parser.NewTokenizer(imageFolder, templateFolder)
	article := parser.Parse(string(fileContent), "test", &tokenizer)

	err = api.DownloadImages(article.Images, imageFolder)
	sigolo.FatalCheck(err)

	_, err = html.Generate(article, "../test/", "../example/style.css", imageFolder, mathFolder)
	sigolo.FatalCheck(err)

	sigolo.Info("Start generating EPUB file")
	metadata := project.Metadata{
		Title: "Foobar",
	}
	err = epub.Generate([]string{"../test/test.html"}, "../test/test.epub", "../example/style.css", "../example/wikipedia-astronomie-cover.png", metadata)
	sigolo.FatalCheck(err)
	sigolo.Info("Successfully created EPUB file")
}
