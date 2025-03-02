package config

import (
	"encoding/json"
	"fmt"
	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"runtime"
	"wiki2book/generator"
	"wiki2book/util"
)

const (
	MathConverterNone      = "none"
	MathConverterWikimedia = "wikimedia"
	MathConverterRsvg      = "rsvg"

	OutputTypeEpub2 = "epub2"
	OutputTypeEpub3 = "epub3"

	OutputDriverPandoc   = "pandoc"
	OutputDriverInternal = "internal"
)

var tocDepthDefault = 2
var workerThreadsDefault = 5

// Current config initialized with default values, which allows wiki2book to run without any specified config file.
var Current = &Configuration{
	OutputType:                     OutputTypeEpub2,
	OutputDriver:                   OutputDriverPandoc,
	CacheDir:                       getDefaultCacheDir(),
	StyleFile:                      getDefaultStyleFile(),
	ImagesToGrayscale:              false,
	ConvertPDFsToImages:            false,
	IgnoredTemplates:               []string{},
	TrailingTemplates:              []string{},
	IgnoredImageParams:             []string{},
	IgnoredMediaTypes:              []string{"gif", "mp3", "mp4", "pdf", "oga", "ogg", "ogv", "wav", "webm"},
	WikipediaInstance:              "en",
	WikipediaHost:                  "wikipedia.org",
	WikipediaImageHost:             "upload.wikimedia.org",
	WikipediaMathRestApi:           "https://wikimedia.org/api/rest_v1/media/math",
	WikipediaImageArticleInstances: []string{"commons", "en"},
	FilePrefixe:                    []string{"file", "image", "media"},
	AllowedLinkPrefixes:            []string{"arxiv", "doi"},
	CategoryPrefixes:               []string{"category"},
	MathConverter:                  "wikimedia",
	RsvgConvertExecutable:          "rsvg-convert",
	RsvgMathStylesheet:             getDefaultRsvgStyleFile(),
	ImageMagickExecutable:          "magick",
	PandocExecutable:               "pandoc",
	TocDepth:                       &tocDepthDefault,
	WorkerThreads:                  &workerThreadsDefault,
}

func getDefaultCacheDir() string {
	userCacheDir, err := os.UserCacheDir()
	sigolo.FatalCheck(err)

	return filepath.Join(userCacheDir, "wiki2book")
}

func getDefaultStyleFile() string {
	linuxDefaultFile := "/usr/share/wiki2book/style.css"
	if runtime.GOOS == "linux" && util.PathExists(linuxDefaultFile) {
		return linuxDefaultFile
	}
	return ""
}

func getDefaultRsvgStyleFile() string {
	linuxDefaultFile := "/usr/share/wiki2book/rsvg-math.css"
	if runtime.GOOS == "linux" && util.PathExists(linuxDefaultFile) {
		return linuxDefaultFile
	}
	return ""
}

// Configuration is a struct with application-wide configurations and language-specific strings (e.g. templates to
// ignore). Some configurations are mandatory, which means that wiki2book will definitely crash if the config entry is
// not given. Entries marked as non-mandatory may also cause a crash.
// The configuration differs from a project-config by the following rule of thumb: This contains technical and project-
// independent stuff. Some properties, though, might exist in both, this Configuration and the project.Project struct.
type Configuration struct {
	/*
		Forces wiki2book to recreate HTML files even if they exists from a previous run.

		Default: false
		JSON example: "force-regenerate-html": true
	*/
	ForceRegenerateHtml bool `json:"force-regenerate-html" help:"Forces wiki2book to recreate HTML files even if they exists from a previous run." short:"r"`

	/*
		Sets the 'width' and 'height' property of an SimpleSvgAttributes image to its viewbox width and height. This might fix wrong SVG sizes on some eBook-readers.

		Default: false
		JSON example: "svg-size-to-viewbox": true
	*/
	SvgSizeToViewbox bool `json:"svg-size-to-viewbox" help:"Sets the 'width' and 'height' property of an SimpleSvgAttributes image to its viewbox width and height. This might fix wrong SVG sizes on some eBook-readers."`

	/*
		The type of the final result.

		Default: epub2
		Possible values: epub2, epub3
		JSON example: "output-type": "epub2"
	*/
	OutputType string `json:"output-type" help:"The output file type. Possible values are: \"epub2\" (default), \"epub3\"." placeholder:"<type>"`

	/*
		The way the final output is created.

		Default: pandoc
		Possible values: pandoc, internal
		JSON example: "output-driver": "pandoc"
	*/
	OutputDriver string `json:"output-driver" help:"The method to generate the output file. Available driver: \"pandoc\" (default), \"internal\" (experimental!)" placeholder:"<driver>"`

	/*
		The directory where all intermediate files are stored. Relative paths are relative to the config file. The
		default value is empty and therefore uses the default cache directory returned by the golang function
		os.UserCacheDir().

		Default: "<user-cache-dir>/wiki2book"
		JSON example: "cache-dir": "/path/to/cache"
	*/
	CacheDir string `json:"cache-dir" help:"The directory where all cached files will be written to." placeholder:"<dir>"`

	/*
		The CSS style file that should be embedded into the eBook. Relative paths are relative to the config file.

		Default: "/use/share/wiki2book/style.css" on Linux when it exists; "" otherwise
		JSON example: "style-file": "my-style.css"
	*/
	StyleFile string `json:"style-file" help:"The CSS file that should be used." placeholder:"<file>"`

	/*
		The image file that should be the cover of the eBook. Relative paths are relative to the config file.

		Default: ""
		JSON example: "cover-image": "nice-picture.jpeg"
	*/
	CoverImage string `json:"cover-image" help:"A cover image for the front cover of the eBook." placeholder:"<file>"`

	/*
		The executable name or file for rsvg-convert.

		Default: "rsvg-convert"
		JSON example: "rsvg-convert-executable": "/path/to/rsvg-convert"
	*/
	RsvgConvertExecutable string `json:"rsvg-convert-executable" help:"The executable name or file for rsvg-convert." placeholder:"<file>"`

	/*
		Specifies the path of the CSS file that should be used when converting math SVGs to PNGs using the
		"rsvg-convert" command. Relative paths are relative to the config file.

		Default: "/use/share/wiki2book/rsvg-math.css" on Linux when it exists; "" otherwise
	*/
	RsvgMathStylesheet string `json:"rsvg-math-stylesheet" help:"Stylesheet for rsvg-convert when using the rsvg converter for math SVGs." placeholder:"<file>"`

	/*
		The executable name or file for ImageMagick.

		Default: "magick"
		JSON example: "imagemagick-executable": "/path/to/imagemagick"
	*/
	ImageMagickExecutable string `json:"imagemagick-executable" help:"The executable name or file for ImageMagick." placeholder:"<file>"`

	/*
		The executable name or file for pandoc.

		Default: "pandoc"
		JSON example: "pandoc-executable": "/path/to/pandoc"
	*/
	PandocExecutable string `json:"pandoc-executable" help:"The executable name or file for pandoc." placeholder:"<file>"`

	/*
		The data directory for pandoc. Relative paths are relative to the config file.

		Default: ""
		JSON example: "pandoc-data-dir": "./my-folder/"
	*/
	PandocDataDir string `json:"pandoc-data-dir" help:"The data directory for pandoc. This enables you to override pandocs defaults for HTML and therefore EPUB generation." placeholder:"<dir>"`

	/*
		A list of font files that should be used. They then can be referenced from the style CSS file. Relative paths are relative to the config file.

		Default: []
		JSON example: "font-files": ["./fontA.ttf", "/path/to/fontB.ttf"]
	*/
	FontFiles []string `json:"font-files" help:"A list of font files that should be used. They are references in your style file." placeholder:"<file>"`

	/*
		Set to true in order to convert raster images to grayscale. Relative paths are relative to the config file.

		Default: false
		JSON example: "images-to-grayscale": true
	*/
	ImagesToGrayscale bool `json:"images-to-grayscale" help:"Set to true in order to convert raster images to grayscale."`

	/*
		When set to true, references PDF files, e.g. with "[[File:foo.pdf]]" are treated as images and will be converted
		into a PNG using ImageMagick. PDFs will still be converted into images, even when the "pdf" media type is present
		in the IgnoredMediaTypes list.

		Default: false
		JSON example: "convert-pdfs-to-images": true
	*/
	ConvertPDFsToImages bool `json:"convert-pdfs-to-images" name:"convert-pdfs-to-images" help:"Set to true in order to convert referenced PDFs into images."`

	/*
		List of templates that should be ignored and removed from the input wikitext. The list must be in lower case.

		Default: Empty list
		JSON example: "ignored-templates": [ "foo", "bar" ]
		This ignores {{foo}} and {{bar}} occurrences in the input text.
	*/
	IgnoredTemplates []string `json:"ignored-templates" help:"List of templates that should be ignored and removed from the input wikitext. The list must be in lower case."`

	/*
		List of templates that will be moved to the end of the document. Theses are e.g. remarks on the article that
		are important but should be shown as a remark after the actual content of the article.

		Default: Empty list
		JSON example: "trailing-templates": [ "foo", "bar" ]
		This moves {{foo}} and {{bar}} to the end of the document.
	*/
	TrailingTemplates []string `json:"trailing-templates" help:"List of templates that will be moved to the end of the document."`

	/*
		Parameters of images that should be ignored. The list must be in lower case.

		Default: Empty list
		JSON example: "ignored-image-params": [ "alt", "center" ]
		This ignores the image parameters "alt" and "center" including any parameter values like "alt"="some alt text".
	*/
	IgnoredImageParams []string `json:"ignored-image-params" help:"Parameters of images that should be ignored. The list must be in lower case."`

	/*
		List of media types to ignore, i.e. list of file extensions. Some media types (e.g. videos) are not of much use
		for a book.

		Default: [ "gif", "mp3", "mp4", "pdf", "oga", "ogg", "ogv", "wav", "webm" ]
	*/
	IgnoredMediaTypes []string `json:"ignored-media-types" help:"List of media types to ignore, i.e. list of file extensions."`

	/*
		The subdomain of the Wikipedia instance.

		Default: "en"
		JSON example: "wikipedia-instance": "de"
		This config uses the German Wikipedia.
	*/
	WikipediaInstance string `json:"wikipedia-instance" help:"The subdomain of the Wikipedia instance."`

	/*
		The domain of the Wikipedia instance.

		Default: "wikipedia.org"
		JSON example: "wikipedia-host": "my-server.com"
	*/
	WikipediaHost string `json:"wikipedia-host" help:"The domain of the Wikipedia instance."`

	/*
		The domain of the Wikipedia image instance.

		Default: "wikimedia.org"
		JSON example: "wikipedia-image-host": "my-image-server.com"
	*/
	WikipediaImageHost string `json:"wikipedia-image-host" help:"The domain of the Wikipedia image instance."`

	/*
		The URL to the math API of wikipedia. This API provides rendering functionality to turn math-objects into PNGs or SVGs.

		Default: "https://wikimedia.org/api/rest_v1/media/math"
		JSON example: "wikipedia-math-rest-api": "my-math-server.com/api"
	*/
	WikipediaMathRestApi string `json:"wikipedia-math-rest-api" help:"The URL to the math API of wikipedia."`

	/*
		Wikipedia instances (subdomains) of the wikipedia image host where images should be searched for. Each image has its own article, which is fetched from
		these Wikipedia instances (in the given order).

		Default: [ "commons", "en" ]
		JSON example: "wikipedia-image-article-instances": [ "commons", "de" ]
	*/
	WikipediaImageArticleInstances []string `json:"wikipedia-image-article-instances" help:"Wikipedia instances (subdomains) of the wikipedia image host where images should be searched for."`

	/*
		A list of prefixes to detect files, e.g. in "File:picture.jpg" the substring "File" is the image prefix. The list
		must be in lower case.

		Default: [ "file", "image", "media" ]
		JSON example: "file-prefixe": [ "file", "datei" ]
	*/
	FilePrefixe []string `json:"file-prefixe" help:"A list of prefixes to detect files, e.g. in \"File:picture.jpg\" the substring \"File\" is the image prefix."`

	/*
		A list of prefixes that are considered links and are therefore not removed. All prefixes  specified by
		"FilePrefixe" are considered to be allowed prefixes. Any other not explicitly allowed prefix of a link causes
		the link to get removed. This especially happens for inter-wiki-links if the Wikipedia instance is not
		explicitly allowed using this list.

		Default: [ "arxiv", "doi" ]
	*/
	AllowedLinkPrefixes []string `json:"allowed-link-prefixe" help:"A list of prefixes that are considered links and are therefore not removed."`

	/*
		A list of category prefixes, which are technically internals links. However, categories will be removed from
		the input wikitext.

		Default: [ "category" ]
	*/
	CategoryPrefixes []string `json:"category-prefixes" help:"A list of category prefixes, which are technically internals links."`

	/*
		Sets the converter to turn math SVGs into PNGs. This can be one of the following values:
			- "none": Uses no converter, instead the plain SVG file is inserted into the ebook.
			- "wikimedia": Uses the online API of Wikimedia to get the PNG version of a math expression.
			- "rsvg": Uses "rsvg-convert" to convert SVG files to PNGs.

		Default: [ "wikimedia" ]
	*/
	MathConverter string `json:"math-converter" help:"Converter turning math SVGs into PNGs."`

	/*
		Sets the depth of the table of content, i.e. how many sub-headings should be visible.

		Examples:
			- A value of 1 means only the h1 headings are visible in the table of content.
			- A value of 3 means h1, h2 and h3 are visible.
			- A value of 0 means the table of content is not visible at all.

		Default: 2
		Allowed values: 0 - 6
	*/
	TocDepth *int `json:"toc-depth" help:"Depth of the table of content. Allowed range is 0 - 6."`

	/*
		Number of threads to process the articles. Only affects projects but not single articles or the standalone mode.
		A higher number of threads might increase performance, but it also puts more stress on the Wikipedia API, which
		might lead to "too many requests"-errors. These errors are handled by wiki2book, but a high thread count might
		still negatively affect wiki2book. Use a value of 1 to disable parallel processing.

		Default: 5
		Allowed values: 1 - unlimited
	*/
	WorkerThreads *int `json:"worker-threads" help:"Number of threads to process the articles. Only affects projects but not single articles or the standalone mode. The value must at least be 1."`
}

func (c *Configuration) makePathsAbsolute(file string) {
	absoluteConfigPath, err := util.ToAbsolutePath(file)
	sigolo.FatalCheck(err)

	absoluteConfigDir := filepath.Dir(absoluteConfigPath)

	c.CacheDir, err = util.ToAbsolutePathWithBasedir(absoluteConfigDir, c.CacheDir)
	sigolo.FatalCheck(err)

	c.StyleFile, err = util.ToAbsolutePathWithBasedir(absoluteConfigDir, c.StyleFile)
	sigolo.FatalCheck(err)

	c.CoverImage, err = util.ToAbsolutePathWithBasedir(absoluteConfigDir, c.CoverImage)
	sigolo.FatalCheck(err)

	c.PandocDataDir, err = util.ToAbsolutePathWithBasedir(absoluteConfigDir, c.PandocDataDir)
	sigolo.FatalCheck(err)

	c.RsvgMathStylesheet, err = util.ToAbsolutePathWithBasedir(absoluteConfigDir, c.RsvgMathStylesheet)
	sigolo.FatalCheck(err)

	for i, f := range c.FontFiles {
		absoluteFile := filepath.Join(absoluteConfigDir, f)
		sigolo.FatalCheck(err)
		c.FontFiles[i] = absoluteFile
	}
}

func (c *Configuration) MakePathsAbsoluteToWorkingDir() {
	var err error

	c.CacheDir, err = util.ToAbsolutePath(c.CacheDir)
	sigolo.FatalCheck(err)

	c.StyleFile, err = util.ToAbsolutePath(c.StyleFile)
	sigolo.FatalCheck(err)

	c.CoverImage, err = util.ToAbsolutePath(c.CoverImage)
	sigolo.FatalCheck(err)

	c.PandocDataDir, err = util.ToAbsolutePath(c.PandocDataDir)
	sigolo.FatalCheck(err)

	c.RsvgMathStylesheet, err = util.ToAbsolutePath(c.RsvgMathStylesheet)
	sigolo.FatalCheck(err)

	for i, f := range c.FontFiles {
		absoluteFile, err := util.ToAbsolutePath(f)
		sigolo.FatalCheck(err)
		c.FontFiles[i] = absoluteFile
	}
}

func (c *Configuration) AssertFilesAndPathsExists() {
	util.AssertPathExists(Current.CacheDir)
	util.AssertPathExists(Current.StyleFile)
	util.AssertPathExists(Current.CoverImage)
	util.AssertPathExists(Current.PandocDataDir)
	util.AssertPathExists(Current.RsvgMathStylesheet)
	for _, f := range Current.FontFiles {
		util.AssertPathExists(f)
	}
}

func (c *Configuration) AssertValidity() {
	if c.OutputType != OutputTypeEpub2 && c.OutputType != OutputTypeEpub3 {
		sigolo.Fatalf("Invalid output type '%s'", c.OutputType)
	}
	if c.OutputDriver != OutputDriverPandoc && c.OutputDriver != OutputDriverInternal {
		sigolo.Fatalf("Invalid output driver '%s'", c.OutputDriver)
	}
	err := generator.VerifyOutputAndDriver(c.OutputType, c.OutputDriver)
	if err != nil {
		sigolo.Fatalf("Output type '%s' and driver '%s' are not valid: %+v", c.OutputType, c.OutputDriver, err)
	}
	if c.MathConverter != MathConverterNone && c.MathConverter != MathConverterWikimedia && c.MathConverter != MathConverterRsvg {
		sigolo.Fatalf("Invalid math converter '%s'", c.OutputDriver)
	}
	if *c.TocDepth < 0 || *c.TocDepth > 6 {
		sigolo.Fatalf("Invalid toc-depth '%d'", c.TocDepth)
	}
	if *c.WorkerThreads < 1 {
		sigolo.Fatalf("Invalid number of worker threads '%d'", c.WorkerThreads)
	}
}

func (c *Configuration) Print() {
	jsonBytes, err := json.MarshalIndent(c, "", "  ")
	sigolo.FatalCheck(err)
	sigolo.Debugf("Configuration:\n%s", string(jsonBytes))
}

func LoadConfig(file string) error {
	projectString, err := os.ReadFile(file)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error reading config file %s", file))
	}

	err = json.Unmarshal(projectString, Current)
	if err != nil {
		return errors.Wrap(err, "Error parsing config file content")
	}

	Current.makePathsAbsolute(file)
	Current.AssertValidity()

	return nil
}
