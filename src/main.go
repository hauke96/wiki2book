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
	"path"
	"path/filepath"
	"strings"
	"time"
)

var cli struct {
	Standalone struct {
		File       string `help:"A mediawiki file tha should be rendered to an eBook." type:"existingfile" arg:""`
		OutputDir  string `help:"The directory where all the files should be put into." short:"o"`
		StyleFile  string `help:"The CSS file that should be used." short:"s" type:"existingfile"`
		CoverImage string `help:"A cover image for the front cover of the eBook." short:"c" type:"existingfile"`
	} `cmd:"" help:"Renders a single mediawiki file into an eBook."`
	Project struct {
		ProjectFile string `help:"A project JSON-file tha should be used to create an eBook." type:"existingfile:" arg:""`
	} `cmd:"" help:"Uses a project file to create the eBook."`
	Article struct {
		// TODO How to deal with multiple languages? A new parameter "Language string"?
		ArticleName string `help:"The name of the article to render." arg:""`
		OutputDir   string `help:"The directory where all the files should be put into." short:"o"`
		StyleFile   string `help:"The CSS file that should be used." short:"s" type:"existingfile"`
		CoverImage  string `help:"A cover image for the front cover of the eBook." short:"c" type:"existingfile"`
	} `cmd:"" help:"Renders a specific article into an eBook."`
}

func main() {
	ctx := kong.Parse(&cli)

	start := time.Now()

	switch ctx.Command() {
	case "standalone <file>":
		generateStandaloneEbook(cli.Standalone.File, cli.Standalone.OutputDir, cli.Standalone.StyleFile, cli.Standalone.CoverImage)
	case "project <project-file>":
		generateProjectEbook(cli.Project.ProjectFile)
	default:
		sigolo.Fatal("Unknown command: %v\n%#v", ctx.Command(), ctx)
	}

	end := time.Now()
	sigolo.Debug("Start   : %s", start.Format(time.RFC1123))
	sigolo.Debug("End     : %s", end.Format(time.RFC1123))
	sigolo.Debug("Duration: %f seconds", end.Sub(start).Seconds())
}

func generateProjectEbook(projectFile string) {
	var err error
	// Enable this to create a profiling file. Then use the command "go tool pprof src ./profiling.prof" and enter "web" to open a diagram in your browser.
	//f, err := os.Create("profiling.prof")
	//sigolo.FatalCheck(err)
	//
	//err = pprof.StartCPUProfile(f)
	//sigolo.FatalCheck(err)
	//defer pprof.StopCPUProfile()

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
}

// TODO just create an instance of type "Project" and create an eBook using that faked project.
func generateStandaloneEbook(inputFile string, outputFolder string, styleFile string, coverImage string) {
	imageFolder := path.Join(outputFolder, "images")
	mathFolder := path.Join(outputFolder, "math")
	templateFolder := path.Join(outputFolder, "templates")

	_, inputFileName := path.Split(inputFile)
	title := strings.Split(inputFileName, ".")[0]

	fileContent, err := ioutil.ReadFile(inputFile)
	sigolo.FatalCheck(err)

	tokenizer := parser.NewTokenizer(imageFolder, templateFolder)
	article := parser.Parse(string(fileContent), title, &tokenizer)

	err = api.DownloadImages(article.Images, imageFolder)
	sigolo.FatalCheck(err)

	_, err = html.Generate(article, outputFolder, styleFile, imageFolder, mathFolder)
	sigolo.FatalCheck(err)

	sigolo.Info("Start generating EPUB file")
	metadata := project.Metadata{
		Title: title,
	}

	htmlFile := path.Join(outputFolder, title+".html")
	epubFile := path.Join(outputFolder, title+".epub")

	err = epub.Generate([]string{htmlFile}, epubFile, styleFile, coverImage, metadata)
	sigolo.FatalCheck(err)
	sigolo.Info("Successfully created EPUB file")
}
