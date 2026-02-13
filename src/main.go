package main

import (
	"os"
	"runtime/pprof"
	"runtime/trace"
	"strings"
	"time"
	"wiki2book/cache"
	"wiki2book/config"
	"wiki2book/generator"
	"wiki2book/server"
	"wiki2book/util"

	"github.com/hauke96/sigolo/v2"
	"github.com/spf13/cobra"
)

const (
	defaultEpubOutputFile = "ebook.epub"
)

var cliConfig = config.NewDefaultConfig()

func main() {
	rootCmd := initCli()

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func initCli() *cobra.Command {
	var cliConfigFile = ""
	var cliLogging = ""
	var cliOutputFileArgKey = "output"
	var cliOutputFile = ""
	var cliDiagnosticsProfiling = false
	var cliDiagnosticsTrace = false
	var start time.Time

	configService := config.NewConfigService()
	fileCache := cache.NewCache(configService)
	ebookGeneratorService := generator.NewEbookGenerator(configService, fileCache)

	rootCmd := &cobra.Command{
		Use:   "wiki2book",
		Short: "A CLI tool to turn one or multiple Wikipedia articles into a good-looking eBook.",
		Long:  "A CLI tool to turn one or multiple Wikipedia articles into a good-looking eBook.",
	}

	rootCmd.Version = util.VERSION
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		cliOutputFile = initialize(configService, cliLogging, cliConfig, cliConfigFile, cliDiagnosticsProfiling, cliDiagnosticsTrace, cliOutputFile)
		start = time.Now()
	}
	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		end := time.Now()
		sigolo.Debugf("Start   : %s", start.Format(util.RFC1123Millis))
		sigolo.Debugf("End     : %s", end.Format(util.RFC1123Millis))
		sigolo.Debugf("Duration: %f seconds", end.Sub(start).Seconds())
	}

	rootCmd.PersistentFlags().BoolP("help", "h", false, "This help message.")
	rootCmd.PersistentFlags().StringVarP(&cliConfigFile, "config", "c", "", "The path to the overall application config. If not specified, default values are used.")
	rootCmd.PersistentFlags().StringVarP(&cliLogging, "logging", "l", "info", "Logging verbosity. Possible values: \"info\", \"debug\", \"trace\".")

	rootCmd.PersistentFlags().StringVarP(&cliOutputFile, cliOutputFileArgKey, "o", defaultEpubOutputFile, "The path to the output file.")

	rootCmd.PersistentFlags().BoolVar(&cliDiagnosticsProfiling, "diagnostics-profiling", cliDiagnosticsProfiling, "Enable profiling and write results to ./profiling.prof.")
	rootCmd.PersistentFlags().BoolVar(&cliDiagnosticsTrace, "diagnostics-trace", cliDiagnosticsTrace, "Enable tracing to analyse memory usage and write results to ./trace.out.")

	rootCmd.PersistentFlags().BoolVarP(&cliConfig.ForceRegenerateHtml, "force-regenerate-html", "r", cliConfig.ForceRegenerateHtml, "Forces wiki2book to recreate HTML files even if they exists from a previous run.")
	rootCmd.PersistentFlags().BoolVar(&cliConfig.SvgSizeToViewbox, "svg-size-to-viewbox", cliConfig.SvgSizeToViewbox, "Sets the 'width' and 'height' property of an SimpleSvgAttributes image to its viewbox width and height. This might fix wrong SVG sizes on some eBook-readers.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.OutputType, "output-type", cliConfig.OutputType, "The output file type. Possible values are: 'epub2', 'epub3', 'stats-json' and 'stats.txt'.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.OutputDriver, "output-driver", cliConfig.OutputDriver, "The method to generate the output file. Available driver: 'pandoc', 'internal' (experimental!)")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CacheDir, "cache-dir", cliConfig.CacheDir, "The directory where all cached files will be written to.")
	rootCmd.PersistentFlags().Int64Var(&cliConfig.CacheMaxSize, "cache-max-size", cliConfig.CacheMaxSize, "The maximum size of the file cache in bytes.")
	rootCmd.PersistentFlags().Int64Var(&cliConfig.CacheMaxAge, "cache-max-age", cliConfig.CacheMaxAge, "The maximum age in minutes of files in the cache. All files older than this, will be downloaded/recreated again.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CacheEvictionStrategy, "cache-eviction-strategy", cliConfig.CacheEvictionStrategy, "The strategy by which files are removed from the case when it's full. Can be: 'none', 'lru', 'largest'")
	rootCmd.PersistentFlags().StringVar(&cliConfig.StyleFile, "style-file", cliConfig.StyleFile, "The CSS file that should be used.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CoverImage, "cover-image", cliConfig.CoverImage, "A cover image for the front cover of the eBook.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CommandTemplateSvgToPng, "command-template-svg-to-png", cliConfig.CommandTemplateSvgToPng, "Command template to use for SVG to PNG conversion. Must contain the placeholders '{INPUT}' and '{OUTPUT}'.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CommandTemplateMathSvgToPng, "command-template-math-svg-to-png", cliConfig.CommandTemplateMathSvgToPng, "Command template to use for math SVG to PNG conversion. Must contain the placeholders '{INPUT}' and '{OUTPUT}'.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CommandTemplateImageProcessing, "command-template-image-processing", cliConfig.CommandTemplateImageProcessing, "Command template to use for math SVG to PNG conversion. Disables processing and uses original images when empty. When set, it must contain the placeholders '{INPUT}' and '{OUTPUT}'.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CommandTemplatePdfToPng, "command-template-pdf-to-png", cliConfig.CommandTemplatePdfToPng, "Command template to use for PDF to PNG conversion. Must contain the placeholders '{INPUT}' and '{OUTPUT}'.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CommandTemplateWebpToPng, "command-template-webp-to-png", cliConfig.CommandTemplateWebpToPng, "Command template to use for math WebP to PNG conversion. Disables conversion when empty. When set, it must contain the placeholders '{INPUT}' and '{OUTPUT}'.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.PandocExecutable, "pandoc-executable", cliConfig.PandocExecutable, "The executable name or file for pandoc.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.PandocDataDir, "pandoc-data-dir", cliConfig.PandocDataDir, "The data directory for pandoc. This enables you to override pandocs defaults for HTML and therefore EPUB generation.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.FontFiles, "font-files", cliConfig.FontFiles, "A list of font files that should be used. They are references in your style file.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.IgnoredTemplates, "ignored-templates", cliConfig.IgnoredTemplates, "List of templates that should be ignored and removed from the input wikitext. The list must be in lower case.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.TrailingTemplates, "trailing-templates", cliConfig.TrailingTemplates, "List of templates that will be moved to the end of the document.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.IgnoredImageParams, "ignored-image-params", cliConfig.IgnoredImageParams, "Parameters of images that should be ignored. The list must be in lower case.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.IgnoredMediaTypes, "ignored-media-types", cliConfig.IgnoredMediaTypes, "List of media types to ignore, i.e. list of file extensions.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.WikipediaInstance, "wikipedia-instance", cliConfig.WikipediaInstance, "The subdomain of the Wikipedia instance.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.WikipediaHost, "wikipedia-host", cliConfig.WikipediaHost, "The domain of the Wikipedia instance.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.WikipediaImageHost, "wikipedia-image-host", cliConfig.WikipediaImageHost, "The domain of the Wikipedia image instance.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.WikipediaMathRestApi, "wikipedia-math-rest-api", cliConfig.WikipediaMathRestApi, "The URL to the math API of wikipedia.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.WikipediaImageArticleHosts, "wikipedia-image-article-hosts", cliConfig.WikipediaImageArticleHosts, "Hosts used to search for image article files.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.FilePrefixes, "file-prefixes", cliConfig.FilePrefixes, "A list of prefixes to detect files, e.g. in 'File:picture.jpg' the substring 'File' is the image prefix.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.AllowedLinkPrefixes, "allowed-link-prefixes", cliConfig.AllowedLinkPrefixes, "A list of prefixes that are considered links and are therefore not removed.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.CategoryPrefixes, "category-prefixes", cliConfig.CategoryPrefixes, "A list of category prefixes, which are technically internals links.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.MathConverter, "math-converter", cliConfig.MathConverter, "Converter turning math SVGs into PNGs.")
	rootCmd.PersistentFlags().IntVar(&cliConfig.TocDepth, "toc-depth", cliConfig.TocDepth, "Depth of the table of content. Allowed range is 0 - 6.")
	rootCmd.PersistentFlags().IntVar(&cliConfig.WorkerThreads, "worker-threads", cliConfig.WorkerThreads, "Number of threads to process the articles. Only affects projects but not single articles or the standalone mode. The value must at least be 1.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.UserAgentTemplate, "user-agent-template", cliConfig.UserAgentTemplate, "Template for the user-agent used in HTTP requests.")

	projectCmd := getCommand("project [file]", "Uses a project file to create the eBook.", 1)
	projectCmd.Args = cobra.MatchAll(cobra.ExactArgs(1))
	projectCmd.Run = func(cmd *cobra.Command, args []string) {
		sigolo.Infof("Prepare generating eBook from project")
		if !rootCmd.PersistentFlags().Changed(cliOutputFileArgKey) {
			// In case the output file was not specified, we don't want to use this file path but see if the project
			// file contains the output file. This is handled in the function called here.
			cliOutputFile = ""
		}

		proj := ebookGeneratorService.CreateProject(
			args[0],
			cliOutputFile,
			cliConfig,
		)
		// TODO Should not be necessary since it's done by the CreateProject function: config.MergeIntoCurrentConfig(cliConfig)
		ebookGeneratorService.GenerateBookFromProject(proj)
		fileCache.CleanUpTempDir()
	}

	articleCmd := getCommand("article [name]", "Renders a single article into an eBook.", 1)
	articleCmd.Args = cobra.MatchAll(cobra.ExactArgs(1))
	articleCmd.Run = func(cmd *cobra.Command, args []string) {
		sigolo.Infof("Prepare generating eBook from single article")
		configService.MergeIntoCurrentConfig(cliConfig)
		ebookGeneratorService.GenerateArticleEbook(
			args[0],
			cliOutputFile,
		)
		fileCache.CleanUpTempDir()
	}

	standaloneCmd := getCommand("standalone [file]", "Renders a single mediawiki file into an eBook.", 1)
	standaloneCmd.Args = cobra.MatchAll(cobra.ExactArgs(1))
	standaloneCmd.Run = func(cmd *cobra.Command, args []string) {
		sigolo.Infof("Prepare generating eBook from standalone mediawiki file")
		configService.MergeIntoCurrentConfig(cliConfig)
		ebookGeneratorService.GenerateStandaloneEbook(
			args[0],
			cliOutputFile,
		)
		fileCache.CleanUpTempDir()
	}

	serverCmd := getCommand("server", "Starts wiki2book in server mode handling HTTP requests to create eBooks.", 0)
	serverCmd.PersistentFlags().IntVar(&cliConfig.ServerPort, "server-port", cliConfig.ServerPort, "Port on which wiki2book should receive HTTP requests.")
	serverCmd.Run = func(cmd *cobra.Command, args []string) {
		sigolo.Infof("Prepare starting wiki2book in server mode")
		configService.MergeIntoCurrentConfig(cliConfig)
		serverInstance := server.NewServer(configService, fileCache, ebookGeneratorService)
		serverInstance.Start()
		fileCache.CleanUpTempDir()
	}

	rootCmd.AddCommand(projectCmd, articleCmd, standaloneCmd, serverCmd)

	rootCmd.InitDefaultHelpCmd()
	var helpCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "help" {
			helpCommand = cmd
			break
		}
	}
	if helpCommand == nil {
		sigolo.Fatal("Help command not found")
	}
	helpCommand.Short = "Help about any command."
	helpCommand.Long = "Help provides help for any command in the application. Simply type 'help [command]' for further details on that command."

	return rootCmd
}

func initialize(configService *config.ConfigService, cliLogging string, cliConfig *config.Configuration, cliConfigFile string, cliDiagnosticsProfiling bool, cliDiagnosticsTrace bool, cliOutputFile string) string {
	if strings.ToLower(cliLogging) == "debug" {
		sigolo.SetDefaultLogLevel(sigolo.LOG_DEBUG)
	} else if strings.ToLower(cliLogging) == "trace" {
		sigolo.SetDefaultLogLevel(sigolo.LOG_TRACE)
	} else if strings.ToLower(cliLogging) == "info" {
		sigolo.SetDefaultLogLevel(sigolo.LOG_INFO)
		sigolo.SetDefaultFormatFunctionAll(sigolo.LogPlain)
	} else {
		sigolo.SetDefaultFormatFunctionAll(sigolo.LogPlain)
		sigolo.Fatalf("Unknown logging level '%s'", cliLogging)
	}

	sigolo.Tracef("CLI config:\n%+v", cliConfig)

	if cliConfigFile != "" {
		err := configService.LoadFromConfig(cliConfigFile)
		sigolo.FatalCheck(err)
	}

	if cliDiagnosticsProfiling {
		f, err := os.Create("profiling.prof")
		sigolo.FatalCheck(err)

		err = pprof.StartCPUProfile(f)
		sigolo.FatalCheck(err)
		defer pprof.StopCPUProfile()
	}

	if cliDiagnosticsTrace {
		f, err := os.Create("trace.out")
		sigolo.FatalCheck(err)

		err = trace.Start(f)
		sigolo.FatalCheck(err)
		defer trace.Stop()
	}

	cliOutputFile, err := util.ToAbsolutePath(cliOutputFile)
	sigolo.FatalCheck(err)

	return cliOutputFile
}

func getCommand(use string, shortDoc string, numberOfArgs int) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: shortDoc,
		Long:  shortDoc,
		Args:  cobra.ExactArgs(numberOfArgs),
	}

	return cmd
}
