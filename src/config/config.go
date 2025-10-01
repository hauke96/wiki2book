package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"wiki2book/util"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

const (
	MathConverterNone      = "none"
	MathConverterWikimedia = "wikimedia"
	MathConverterTemplate  = "template"

	OutputTypeEpub2     = "epub2"
	OutputTypeEpub3     = "epub3"
	OutputTypeStatsJson = "stats-json"
	OutputTypeStatsTxt  = "stats-txt"

	OutputDriverPandoc   = "pandoc"
	OutputDriverInternal = "internal"

	linuxDefaultRsvgMathStyleFile = "/usr/share/wiki2book/rsvg-math.css"
	linuxDefaultStyleFile         = "/usr/share/wiki2book/style.css"

	InputPlaceholder  = "{INPUT}"
	OutputPlaceholder = "{OUTPUT}"

	CacheEvictionStrategyLargest = "largest"
	CacheEvictionStrategyLru     = "lru"
	CacheEvictionStrategyNone    = "none"

	defaultCommandTemplateSvgToPng                   = "rsvg-convert -o " + OutputPlaceholder + " " + InputPlaceholder
	defaultCommandTemplateLinuxMathSvgToPngWithStyle = "rsvg-convert -s " + linuxDefaultRsvgMathStyleFile + " -o " + OutputPlaceholder + " " + InputPlaceholder
	defaultCommandTemplateImageProcessing            = "magick " + InputPlaceholder + " -resize 600x600> -quality 75 -define PNG:compression-level=9 -define PNG:compression-filter=0 -colorspace gray " + OutputPlaceholder
	defaultCommandTemplatePdfToPng                   = "magick -density 300 " + InputPlaceholder + " " + OutputPlaceholder
	defaultCommandTemplateWebpToPng                  = "magick " + InputPlaceholder + " " + OutputPlaceholder
)

var tocDepthDefault = 2
var workerThreadsDefault = 5

// Current config initialized with default values, which allows wiki2book to run without any specified config file.
var Current = NewDefaultConfig()

// Used during merging to only copy values of fields that have non-default values.
var defaultConfig = NewDefaultConfig()

var defaultValidationErrorHandler = func(err error) {
	sigolo.FatalCheck(err)
}

func NewDefaultConfig() *Configuration {
	return &Configuration{
		OutputType:                     OutputTypeEpub2,
		OutputDriver:                   OutputDriverPandoc,
		CacheDir:                       getDefaultCacheDir(),
		CacheMaxSize:                   100_000_000,
		CacheMaxAge:                    40_320,
		CacheEvictionStrategy:          CacheEvictionStrategyLru,
		StyleFile:                      getDefaultStyleFile(),
		IgnoredTemplates:               []string{},
		TrailingTemplates:              []string{},
		IgnoredImageParams:             []string{},
		IgnoredMediaTypes:              []string{"gif", "mp3", "mp4", "pdf", "oga", "ogg", "ogv", "wav", "webm"},
		WikipediaInstance:              "en",
		WikipediaHost:                  "wikipedia.org",
		WikipediaImageHost:             "upload.wikimedia.org",
		WikipediaImageArticleHosts:     []string{"commons.wikimedia.org", "en.wikipedia.org"},
		WikipediaMathRestApi:           "https://wikimedia.org/api/rest_v1/media/math",
		FilePrefixe:                    []string{"file", "image", "media"},
		AllowedLinkPrefixes:            []string{"arxiv", "doi"},
		CategoryPrefixes:               []string{"category"},
		MathConverter:                  "wikimedia",
		CommandTemplateSvgToPng:        defaultCommandTemplateSvgToPng,
		CommandTemplateMathSvgToPng:    getDefaultMathSvgToPngCommandTemplate(),
		CommandTemplateImageProcessing: defaultCommandTemplateImageProcessing,
		CommandTemplatePdfToPng:        defaultCommandTemplatePdfToPng,
		CommandTemplateWebpToPng:       defaultCommandTemplateWebpToPng,
		PandocExecutable:               "pandoc",
		PandocDataDir:                  "",
		FontFiles:                      []string{},
		TocDepth:                       tocDepthDefault,
		WorkerThreads:                  workerThreadsDefault,
		UserAgentTemplate:              "wiki2book {{VERSION}} (https://github.com/hauke96/wiki2book)",
	}
}

func getDefaultCacheDir() string {
	userCacheDir, err := os.UserCacheDir()
	sigolo.FatalCheck(err)

	return filepath.Join(userCacheDir, "wiki2book")
}

func getDefaultStyleFile() string {
	if runtime.GOOS == "linux" && util.PathExists(linuxDefaultStyleFile) {
		return linuxDefaultStyleFile
	}
	return ""
}

func getDefaultMathSvgToPngCommandTemplate() string {
	if runtime.GOOS == "linux" && util.PathExists(linuxDefaultRsvgMathStyleFile) {
		return defaultCommandTemplateLinuxMathSvgToPngWithStyle
	}
	return defaultCommandTemplateSvgToPng
}

// Configuration is a struct with application-wide configurations and language-specific strings (e.g. templates to
// ignore). Some configurations are mandatory, which means that wiki2book will definitely crash if the config entry is
// not given. Entries marked as non-mandatory may also cause a crash.
// The configuration differs from a project-config by the following rule of thumb: This contains technical and project-
// independent stuff. Some properties, though, might exist in both, this Configuration and the project.Project struct.
type Configuration struct {
	/*
		Forces wiki2book to recreate HTML files even if they exists from a previous run.
		Some
		more
		lines.

		Default: `false`
		JSON example: `"force-regenerate-html": true`
	*/
	ForceRegenerateHtml bool `json:"force-regenerate-html"`

	/*
		Sets the 'width' and 'height' property of an SimpleSvgAttributes image to its viewbox width and height. This might fix wrong SVG sizes on some eBook-readers.

		Default: `false`
		JSON example: `"svg-size-to-viewbox": true`
	*/
	SvgSizeToViewbox bool `json:"svg-size-to-viewbox"`

	/*
		The type of the final result.

		Default: `epub2`
		Allowed values: `epub2`, `epub3`
		JSON example: `"output-type": "epub2"`
	*/
	OutputType string `json:"output-type"`

	/*
		The way the final output is created.

		Default: `pandoc`
		Allowed values: `pandoc`, `internal`
		JSON example: `"output-driver": "pandoc"`
	*/
	OutputDriver string `json:"output-driver"`

	/*
		The directory where all intermediate files are stored. Relative paths are relative to the config file. The
		default value is the default cache directory returned by the golang function os.UserCacheDir().

		Default: `"<user-cache-dir>/wiki2book"`
		JSON example: `"cache-dir": "/path/to/cache"`
	*/
	CacheDir string `json:"cache-dir"`

	/*
		The maximum size of the file cache in bytes.

		Default: `100000000` (100 MiB)
	*/
	CacheMaxSize int64 `json:"cache-max-size"`

	/*
		The maximum age in minutes of files in the cache. All files older than this, will be downloaded/recreated again.
		Note that setting CacheEvictionStrategy to "lru" stays in conflict with this setting, because the LRU cache
		constantly updates timestamps on files.

		Default: `40320` (four weeks)
	*/
	CacheMaxAge int64 `json:"cache-max-age"`

	/*
		The strategy by which files are removed from the case when it's full.

		Default: `"lru"`
		Allowed values:
		<ul>
			<li>`"largest"` - In case the maximum cache size has been reached, the largest file will be removed first.</li>
			<li>`"lru"`     - In case the maximum cache size has been reached, the least recently used file will be removed
			                  first. Note that the LRU cache stays in conflict with the CacheMaxAge setting. Using the
			                  LRU cache constantly updates timestamps on files, which then might stay longer in cache
			                  than CacheMaxAge defines.</li>
			<li>`"none"`    - No cache eviction strategy, i.e. all files are cached and never evicted. Therefore, the
			                  CacheMaxSize setting has no effect.</li>
		</ul>
	*/
	CacheEvictionStrategy string `json:"cache-eviction-strategy"`

	/*
		The CSS style file that should be embedded into the eBook. Relative paths are relative to the config file.

		Default: `"/use/share/wiki2book/style.css"` on Linux when it exists; `""` otherwise
		JSON example: `"style-file": "my-style.css"`
	*/
	StyleFile string `json:"style-file"`

	/*
		The image file that should be the cover of the eBook. Relative paths are relative to the config file.

		Default: `""`
		JSON example: `"cover-image": "nice-picture.jpeg"`
	*/
	CoverImage string `json:"cover-image"`

	/*
		Specifies the template for the command that should be used to convert the SVG files into PNGs. This command
		might use additional parameters in comparison to the normal SVG to PNG command template. An empty value
		deactivates the processing and the original image will be used.

		This template must contain the following placeholders that will be replaced by the actual values before
		executing the command:
		<ul>
			<li>`{INPUT}` : The input SVG file.</li>
			<li>`{OUTPUT}` : The output PNG file.</li>
		</ul>

		Default: `"rsvg-convert -o {OUTPUT} {INPUT}"`
		JSON example: `"command-template-svg-to-png": "my-command --some-arg -i {INPUT} -o {OUTPUT}"`
	*/
	CommandTemplateSvgToPng string `json:"command-template-svg-to-png"`

	/*
		Specifies the template for the command that should be used to convert the SVG files of math expressions into
		PNGs. This template is only used when setting MathConverter to "template". This command might use additional
		parameters in comparison to the normal SVG to PNG command template.

		This template must contain the following placeholders that will be replaced by the actual values before
		executing the command:
		<ul>
			<li>`{INPUT}` : The input SVG file.</li>
			<li>`{OUTPUT}` : The output PNG file.</li>
		</ul>

		Default:
		<ul>
			<li>When the specified CSS file exists: `"rsvg-convert -s /usr/share/wiki2book/rsvg-math.css -o {OUTPUT} {INPUT}"`</li>
			<li>Otherwise: `"rsvg-convert -o {OUTPUT} {INPUT}"`</li>
		</ul>
		JSON example: `"command-template-math-svg-to-png": "my-command --some-arg -i {INPUT} -o {OUTPUT}"`
	*/
	CommandTemplateMathSvgToPng string `json:"command-template-math-svg-to-png"`

	/*
		Specifies the template for the command that should be used to process images. This will be called for each
		downloaded image and can be used to e.g. compress or otherwise process the image. An empty value deactivates
		the processing and the original image will be used.

		This template must contain the following placeholders that will be replaced by the actual values before
		executing the command:
		<ul>
			<li>`{INPUT}` : The input image file.</li>
			<li>`{OUTPUT}` : The output image file.</li>
		</ul>

		Default: `"magick {INPUT} -resize 600x600> -quality 75 -define PNG:compression-level=9 -define PNG:compression-filter=0 -colorspace gray {OUTPUT}"`
		JSON example: `"command-template-image-processing": "my-command --some-arg -i {INPUT} -o {OUTPUT}"`
	*/
	CommandTemplateImageProcessing string `json:"command-template-image-processing"`

	/*
		Specifies the template for the command that should be used to convert PDF into PNG files. An empty value
		deactivates the processing and the original image will be used.

		This template must contain the following placeholders that will be replaced by the actual values before
		executing the command:
		<ul>
			<li>`{INPUT}` : The input PDF file.</li>
			<li>`{OUTPUT}` : The output PNG file.</li>
		</ul>

		Default: `"magick -density 300 {INPUT} {OUTPUT}"`
		JSON example: `"command-template-pdf-to-png": "my-command --some-arg -i {INPUT} -o {OUTPUT}"`
	*/
	CommandTemplatePdfToPng string `json:"command-template-pdf-to-png"`

	/*
		Specifies the template for the command that should be used to convert WebP into PNG files. An empty value
		deactivates the processing and the original image will be used.

		This template must contain the following placeholders that will be replaced by the actual values before
		executing the command:
		<ul>
			<li>`{INPUT}` : The input WebP file.</li>
			<li>`{OUTPUT}` : The output PNG file.</li>
		</ul>

		Default: `"magick {INPUT} {OUTPUT}"`
		JSON example: `"command-template-webp-to-png": "my-command --some-arg -i {INPUT} -o {OUTPUT}"`
	*/
	CommandTemplateWebpToPng string `json:"command-template-webp-to-png"`

	/*
		The executable name or file for pandoc.

		Default: `"pandoc"`
		JSON example: `"pandoc-executable": "/path/to/pandoc"`
	*/
	PandocExecutable string `json:"pandoc-executable"`

	/*
		The data directory for pandoc. Relative paths are relative to the config file.

		Default: `""`
		JSON example: `"pandoc-data-dir": "./my-folder/"`
	*/
	PandocDataDir string `json:"pandoc-data-dir"`

	/*
		A list of font files that should be used. They then can be referenced from the style CSS file. Relative paths are relative to the config file.

		Default: `[]`
		JSON example: `"font-files": ["./fontA.ttf", "/path/to/fontB.ttf"]`
	*/
	FontFiles []string `json:"font-files"`

	/*
		List of templates that should be ignored and removed from the input wikitext. The list must be in lower case.

		Default: `[]`
		JSON example: `"ignored-templates": [ "foo", "bar" ]`
		This ignores `{{foo}}` and `{{bar}}` occurrences in the input text.
	*/
	IgnoredTemplates []string `json:"ignored-templates"`

	/*
		List of templates that will be moved to the end of the document. Theses are e.g. remarks on the article that
		are important but should be shown as a remark after the actual content of the article.

		Default: `[]`
		JSON example: `"trailing-templates": [ "foo", "bar" ]`
		This moves `{{foo}}` and `{{bar}}` to the end of the document.
	*/
	TrailingTemplates []string `json:"trailing-templates"`

	/*
		Parameters of images that should be ignored. The list must be in lower case.

		Default: `[]`
		JSON example: `"ignored-image-params": [ "alt", "center" ]`
		This ignores the image parameters "alt" and "center" including any parameter values like "alt"="some alt text".
	*/
	IgnoredImageParams []string `json:"ignored-image-params"`

	/*
		List of media types to ignore, i.e. list of file extensions. Some media types (e.g. videos) are not of much use
		for a book.

		Default: `[ "gif", "mp3", "mp4", "pdf", "oga", "ogg", "ogv", "wav", "webm" ]`
	*/
	IgnoredMediaTypes []string `json:"ignored-media-types"`

	/*
		The subdomain of the Wikipedia instance.

		Default: `"en"`
		JSON example: `"wikipedia-instance": "de"` This config would then use the German Wikipedia.
	*/
	WikipediaInstance string `json:"wikipedia-instance"`

	/*
		The domain of the Wikipedia instance.

		Default: `"wikipedia.org"`
		JSON example: `"wikipedia-host": "my-server.com"`
	*/
	WikipediaHost string `json:"wikipedia-host"`

	/*
		The domain of the Wikipedia image instance, which should be used to download the actual image files.

		Default: `"upload.wikimedia.org"`
		JSON example: `"wikipedia-image-host": "my-image-server.com"`
	*/
	WikipediaImageHost string `json:"wikipedia-image-host"`

	/*
		Domains used to search for image articles (not the image files themselves, s. WikipediaImageHost). The given
		values are tried in the configured order until a request was successful or the last host has been tried.

		Default: `[ "commons.wikimedia.org", "en.wikipedia.org" ]`
		JSON example: `"wikipedia-image-article-hosts": [ "commons.wikimedia.org" ]`
	*/
	WikipediaImageArticleHosts []string `json:"wikipedia-image-article-hosts"`

	/*
		The URL to the math API of wikipedia. This API provides rendering functionality to turn math-objects into PNGs or SVGs.

		Default: `"https://wikimedia.org/api/rest_v1/media/math"`
		JSON example: `"wikipedia-math-rest-api": "my-math-server.com/api"`
	*/
	WikipediaMathRestApi string `json:"wikipedia-math-rest-api"`

	/*
		A list of prefixes to detect files, e.g. in "File:picture.jpg" the substring "File" is the image prefix. The list
		must be in lower case.

		Default: `[ "file", "image", "media" ]`
		JSON example: `"file-prefixe": [ "file", "datei" ]`
	*/
	FilePrefixe []string `json:"file-prefixe"`

	/*
		A list of prefixes that are considered links and are therefore not removed. All prefixes  specified by
		"FilePrefixe" are considered to be allowed prefixes. Any other not explicitly allowed prefix of a link causes
		the link to get removed. This especially happens for inter-wiki-links if the Wikipedia instance is not
		explicitly allowed using this list.

		Default: `[ "arxiv", "doi" ]`
	*/
	AllowedLinkPrefixes []string `json:"allowed-link-prefixe"`

	/*
		A list of category prefixes, which are technically internals links. However, categories will be removed from
		the input wikitext.

		Default: `[ "category" ]`
	*/
	CategoryPrefixes []string `json:"category-prefixes"`

	/*
		Sets the converter to turn math SVGs into PNGs. This can be one of the following values:
		<ul>
			<li>"none": Uses no converter, instead the plain SVG file is inserted into the ebook.</li>
			<li>"wikimedia": Uses the online API of Wikimedia to get the PNG version of a math expression.</li>
			<li>"template": Uses the CommandTemplateMathSvgToPng to convert math SVG files to PNGs.</li>
		</ul>

		Default: `[ "wikimedia" ]`
	*/
	MathConverter string `json:"math-converter"`

	/*
		Sets the depth of the table of content, i.e. how many sub-headings should be visible.

		Examples:
		<ul>
			<li>A value of 1 means only the h1 headings are visible in the table of content.</li>
			<li>A value of 3 means h1, h2 and h3 are visible.</li>
			<li>A value of 0 means the table of content is not visible at all.</li>
		</ul>

		Default: `2`
		Allowed values: `0` to `6`
	*/
	TocDepth int `json:"toc-depth"`

	/*
		Number of threads to process the articles. Only affects projects but not single articles or the standalone mode.
		A higher number of threads might increase performance, but it also puts more stress on the Wikipedia API, which
		might lead to "too many requests"-errors. These errors are handled by wiki2book, but a high thread count might
		still negatively affect wiki2book. Use a value of 1 to disable parallel processing.

		Default: `5`
		Allowed values: `1` to unlimited
	*/
	WorkerThreads int `json:"worker-threads"`

	/*
		Template string for the user agent used in HTTP requests. There are some placeholders within this template
		string, which are replaced by actual values:
		<ul>
			<li>`{{VERSION}}` - The version of wiki2book as shown by the `--version` CLI argument. Example: `v0.5.0`</li>
		</ul>

		Default: `"wiki2book {{VERSION}} (https://github.com/hauke96/wiki2book)"`
	*/
	UserAgentTemplate string `json:"user-agent-template"`
}

// MergeIntoCurrentConfig goes through all the properties of the given configuration and overwrites the respective field
// in the Current configuration in case the field of the given config is different to the default value.
func MergeIntoCurrentConfig(c *Configuration) {
	if c.ForceRegenerateHtml != defaultConfig.ForceRegenerateHtml {
		sigolo.Tracef("Override ForceRegenerateHtml with %v", c.ForceRegenerateHtml)
		Current.ForceRegenerateHtml = c.ForceRegenerateHtml
	}
	if c.SvgSizeToViewbox != defaultConfig.SvgSizeToViewbox {
		sigolo.Tracef("Override SvgSizeToViewbox with %v", c.SvgSizeToViewbox)
		Current.SvgSizeToViewbox = c.SvgSizeToViewbox
	}
	if c.OutputType != defaultConfig.OutputType {
		sigolo.Tracef("Override OutputType with %s", c.OutputType)
		Current.OutputType = c.OutputType
	}
	if c.OutputDriver != defaultConfig.OutputDriver {
		sigolo.Tracef("Override OutputDriver with %s", c.OutputDriver)
		Current.OutputDriver = c.OutputDriver
	}
	if c.CacheDir != defaultConfig.CacheDir {
		absolutePath, err := util.ToAbsolutePath(c.CacheDir)
		sigolo.FatalCheck(err)
		sigolo.Tracef("Override CacheDir with %s", absolutePath)
		Current.CacheDir = absolutePath
	}
	if c.CacheMaxSize != defaultConfig.CacheMaxSize {
		sigolo.Tracef("Override CacheMaxSize with %d", c.CacheMaxSize)
		Current.CacheMaxSize = c.CacheMaxSize
	}
	if c.CacheMaxAge != defaultConfig.CacheMaxAge {
		sigolo.Tracef("Override CacheMaxAge with %d", c.CacheMaxAge)
		Current.CacheMaxAge = c.CacheMaxAge
	}
	if c.CacheEvictionStrategy != defaultConfig.CacheEvictionStrategy {
		sigolo.Tracef("Override CacheEvictionStrategy with %s", c.CacheEvictionStrategy)
		Current.CacheEvictionStrategy = c.CacheEvictionStrategy
	}
	if c.StyleFile != defaultConfig.StyleFile {
		absolutePath, err := util.ToAbsolutePath(c.StyleFile)
		sigolo.FatalCheck(err)
		sigolo.Tracef("Override StyleFile with %s", absolutePath)
		Current.StyleFile = absolutePath
	}
	if c.CoverImage != defaultConfig.CoverImage {
		absolutePath, err := util.ToAbsolutePath(c.CoverImage)
		sigolo.FatalCheck(err)
		sigolo.Tracef("Override CoverImage with %s", absolutePath)
		Current.CoverImage = absolutePath
	}
	if c.CommandTemplateSvgToPng != defaultConfig.CommandTemplateSvgToPng {
		sigolo.Tracef("Override CommandTemplateSvgToPng with %s", c.CommandTemplateSvgToPng)
		Current.CommandTemplateSvgToPng = c.CommandTemplateSvgToPng
	}
	if c.CommandTemplateMathSvgToPng != defaultConfig.CommandTemplateMathSvgToPng {
		sigolo.Tracef("Override CommandTemplateMathSvgToPng with %s", c.CommandTemplateMathSvgToPng)
		Current.CommandTemplateMathSvgToPng = c.CommandTemplateMathSvgToPng
	}
	if c.CommandTemplateImageProcessing != defaultConfig.CommandTemplateImageProcessing {
		sigolo.Tracef("Override CommandTemplateImageProcessing with %s", c.CommandTemplateImageProcessing)
		Current.CommandTemplateImageProcessing = c.CommandTemplateImageProcessing
	}
	if c.CommandTemplatePdfToPng != defaultConfig.CommandTemplatePdfToPng {
		sigolo.Tracef("Override CommandTemplatePdfToPng with %s", c.CommandTemplatePdfToPng)
		Current.CommandTemplatePdfToPng = c.CommandTemplatePdfToPng
	}
	if c.CommandTemplateWebpToPng != defaultConfig.CommandTemplateWebpToPng {
		sigolo.Tracef("Override CommandTemplateWebpToPng with %s", c.CommandTemplateWebpToPng)
		Current.CommandTemplateWebpToPng = c.CommandTemplateWebpToPng
	}
	if c.PandocExecutable != defaultConfig.PandocExecutable {
		var err error
		newPath := c.PandocExecutable
		if strings.Contains(c.PandocExecutable, "/") {
			// Only convert paths to absolute paths that are actual paths. Just the name of the executable, e.g. just
			// "pandoc", should of course not be converted into a path.
			newPath, err = util.ToAbsolutePath(c.PandocExecutable)
			sigolo.FatalCheck(err)
		}
		sigolo.Tracef("Override PandocExecutable with %s", c.PandocExecutable)
		Current.PandocExecutable = newPath
	}
	if c.PandocDataDir != defaultConfig.PandocDataDir {
		absolutePath, err := util.ToAbsolutePath(c.PandocDataDir)
		sigolo.FatalCheck(err)
		sigolo.Tracef("Override PandocDataDir with %s", absolutePath)
		Current.PandocDataDir = absolutePath
	}
	if !util.EqualsInAnyOrder(c.FontFiles, defaultConfig.FontFiles) {
		absolutePaths, err := util.ToAbsolutePaths(c.FontFiles...)
		sigolo.FatalCheck(err)
		sigolo.Tracef("Override FontFiles with %v", c.SvgSizeToViewbox)
		Current.FontFiles = absolutePaths
	}
	if !util.EqualsInAnyOrder(c.IgnoredTemplates, defaultConfig.IgnoredTemplates) {
		sigolo.Tracef("Override IgnoredTemplates with %v", c.IgnoredTemplates)
		Current.IgnoredTemplates = c.IgnoredTemplates
	}
	if !util.EqualsInAnyOrder(c.TrailingTemplates, defaultConfig.TrailingTemplates) {
		sigolo.Tracef("Override TrailingTemplates with %v", c.TrailingTemplates)
		Current.TrailingTemplates = c.TrailingTemplates
	}
	if !util.EqualsInAnyOrder(c.IgnoredImageParams, defaultConfig.IgnoredImageParams) {
		sigolo.Tracef("Override IgnoredImageParams with %v", c.IgnoredImageParams)
		Current.IgnoredImageParams = c.IgnoredImageParams
	}
	if !util.EqualsInAnyOrder(c.IgnoredMediaTypes, defaultConfig.IgnoredMediaTypes) {
		sigolo.Tracef("Override IgnoredMediaTypes with %v", c.IgnoredMediaTypes)
		Current.IgnoredMediaTypes = c.IgnoredMediaTypes
	}
	if c.WikipediaInstance != defaultConfig.WikipediaInstance {
		sigolo.Tracef("Override WikipediaInstance with %s", c.WikipediaInstance)
		Current.WikipediaInstance = c.WikipediaInstance
	}
	if c.WikipediaHost != defaultConfig.WikipediaHost {
		sigolo.Tracef("Override WikipediaHost with %s", c.WikipediaHost)
		Current.WikipediaHost = c.WikipediaHost
	}
	if c.WikipediaImageHost != defaultConfig.WikipediaImageHost {
		sigolo.Tracef("Override WikipediaImageHost with %s", c.WikipediaImageHost)
		Current.WikipediaImageHost = c.WikipediaImageHost
	}
	if c.WikipediaMathRestApi != defaultConfig.WikipediaMathRestApi {
		sigolo.Tracef("Override WikipediaMathRestApi with %s", c.WikipediaMathRestApi)
		Current.WikipediaMathRestApi = c.WikipediaMathRestApi
	}
	if !util.EqualsInAnyOrder(c.WikipediaImageArticleHosts, defaultConfig.WikipediaImageArticleHosts) {
		sigolo.Tracef("Override WikipediaImageArticleHosts with %v", c.WikipediaImageArticleHosts)
		Current.WikipediaImageArticleHosts = c.WikipediaImageArticleHosts
	}
	if !util.EqualsInAnyOrder(c.FilePrefixe, defaultConfig.FilePrefixe) {
		sigolo.Tracef("Override FilePrefixe with %v", c.FilePrefixe)
		Current.FilePrefixe = c.FilePrefixe
	}
	if !util.EqualsInAnyOrder(c.AllowedLinkPrefixes, defaultConfig.AllowedLinkPrefixes) {
		sigolo.Tracef("Override AllowedLinkPrefixes with %v", c.AllowedLinkPrefixes)
		Current.AllowedLinkPrefixes = c.AllowedLinkPrefixes
	}
	if !util.EqualsInAnyOrder(c.CategoryPrefixes, defaultConfig.CategoryPrefixes) {
		sigolo.Tracef("Override CategoryPrefixes with %v", c.CategoryPrefixes)
		Current.CategoryPrefixes = c.CategoryPrefixes
	}
	if c.MathConverter != defaultConfig.MathConverter {
		sigolo.Tracef("Override MathConverter with %s", c.MathConverter)
		Current.MathConverter = c.MathConverter
	}
	if c.TocDepth != defaultConfig.TocDepth {
		sigolo.Tracef("Override TocDepth with %d", c.TocDepth)
		Current.TocDepth = c.TocDepth
	}
	if c.WorkerThreads != defaultConfig.WorkerThreads {
		sigolo.Tracef("Override WorkerThreads with %d", c.WorkerThreads)
		Current.WorkerThreads = c.WorkerThreads
	}
	if c.UserAgentTemplate != defaultConfig.UserAgentTemplate {
		sigolo.Tracef("Override UserAgentTemplate with %s", c.UserAgentTemplate)
		Current.UserAgentTemplate = c.UserAgentTemplate
	}

	Current.MakePathsAbsoluteToWorkingDir()

	Current.AssertValidity()
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
	for _, f := range Current.FontFiles {
		util.AssertPathExists(f)
	}
}

func (c *Configuration) AssertValidity() {
	isOutputTypeValid := true
	if c.OutputType != OutputTypeEpub2 && c.OutputType != OutputTypeEpub3 && c.OutputType != OutputTypeStatsJson && c.OutputType != OutputTypeStatsTxt {
		isOutputTypeValid = false
		defaultValidationErrorHandler(errors.Errorf("Invalid output type '%s'", c.OutputType))
	}
	isOutputDriverValid := true
	if c.OutputDriver != OutputDriverPandoc && c.OutputDriver != OutputDriverInternal {
		isOutputDriverValid = false
		defaultValidationErrorHandler(errors.Errorf("Invalid output driver '%s'", c.OutputDriver))
	}
	if isOutputTypeValid && isOutputDriverValid {
		err := VerifyOutputAndDriver(c.OutputType, c.OutputDriver)
		if err != nil {
			defaultValidationErrorHandler(errors.Errorf("Output type '%s' and driver '%s' are not valid: %+v", c.OutputType, c.OutputDriver, err))
		}
	}

	if c.MathConverter != MathConverterNone && c.MathConverter != MathConverterWikimedia && c.MathConverter != MathConverterTemplate {
		defaultValidationErrorHandler(errors.Errorf("Invalid math converter '%s'", c.MathConverter))
	}
	if c.TocDepth < 0 || c.TocDepth > 6 {
		defaultValidationErrorHandler(errors.Errorf("Invalid toc-depth '%d'", c.TocDepth))
	}
	if c.WorkerThreads < 1 {
		defaultValidationErrorHandler(errors.Errorf("Invalid number of worker threads '%d'", c.WorkerThreads))
	}

	if c.CommandTemplateSvgToPng != "" {
		if !strings.Contains(c.CommandTemplateSvgToPng, InputPlaceholder) {
			defaultValidationErrorHandler(errors.Errorf("CommandTemplateSvgToPng must contain the '" + InputPlaceholder + "' placeholder"))
		}
		if !strings.Contains(c.CommandTemplateSvgToPng, OutputPlaceholder) {
			defaultValidationErrorHandler(errors.Errorf("CommandTemplateSvgToPng must contain the '" + OutputPlaceholder + "' placeholder"))
		}
	}

	if c.CommandTemplateMathSvgToPng == "" {
		defaultValidationErrorHandler(errors.Errorf("CommandTemplateMathSvgToPng must not be empty"))
	}
	if !strings.Contains(c.CommandTemplateMathSvgToPng, InputPlaceholder) {
		defaultValidationErrorHandler(errors.Errorf("CommandTemplateMathSvgToPng must contain the '" + InputPlaceholder + "' placeholder"))
	}
	if !strings.Contains(c.CommandTemplateMathSvgToPng, OutputPlaceholder) {
		defaultValidationErrorHandler(errors.Errorf("CommandTemplateMathSvgToPng must contain the '" + OutputPlaceholder + "' placeholder"))
	}

	if c.CommandTemplateImageProcessing != "" {
		if !strings.Contains(c.CommandTemplateImageProcessing, InputPlaceholder) {
			defaultValidationErrorHandler(errors.Errorf("CommandTemplateImageProcessing is set and therefore must contain the '" + InputPlaceholder + "' placeholder"))
		}
		if !strings.Contains(c.CommandTemplateImageProcessing, OutputPlaceholder) {
			defaultValidationErrorHandler(errors.Errorf("CommandTemplateImageProcessing is set and therefore must contain the '" + OutputPlaceholder + "' placeholder"))
		}
	}

	if c.CommandTemplatePdfToPng != "" {
		if !strings.Contains(c.CommandTemplatePdfToPng, InputPlaceholder) {
			defaultValidationErrorHandler(errors.Errorf("CommandTemplatePdfToPng must contain the '" + InputPlaceholder + "' placeholder"))
		}
		if !strings.Contains(c.CommandTemplatePdfToPng, OutputPlaceholder) {
			defaultValidationErrorHandler(errors.Errorf("CommandTemplatePdfToPng must contain the '" + OutputPlaceholder + "' placeholder"))
		}
	}

	if c.CommandTemplateWebpToPng != "" {
		if !strings.Contains(c.CommandTemplateWebpToPng, InputPlaceholder) {
			defaultValidationErrorHandler(errors.Errorf("CommandTemplateWebpToPng is set and therefore must contain the '" + InputPlaceholder + "' placeholder"))
		}
		if !strings.Contains(c.CommandTemplateWebpToPng, OutputPlaceholder) {
			defaultValidationErrorHandler(errors.Errorf("CommandTemplateWebpToPng is set and therefore must contain the '" + OutputPlaceholder + "' placeholder"))
		}
	}

	if c.CacheMaxSize <= 0 {
		defaultValidationErrorHandler(errors.Errorf("CacheMaxSize must be larger than 0 but was %d", c.CacheMaxSize))
	}
	if c.CacheMaxAge <= 0 {
		defaultValidationErrorHandler(errors.Errorf("CacheMaxAge must be larger than 0 but was %d", c.CacheMaxAge))
	}
	if c.CacheEvictionStrategy != CacheEvictionStrategyNone && c.CacheEvictionStrategy != CacheEvictionStrategyLru && c.CacheEvictionStrategy != CacheEvictionStrategyLargest {
		defaultValidationErrorHandler(errors.Errorf("CacheEvictionStrategy '%s' is invalid", c.CacheEvictionStrategy))
	}
}

// VerifyOutputAndDriver returns an error if the output type and driver are not compatible and returns nil if they are.
func VerifyOutputAndDriver(outputType string, outputDriver string) error {
	sigolo.Tracef("Verify compatibility of outputType '%s' and outputDriver '%s'", outputType, outputDriver)

	switch outputType {
	case OutputTypeEpub2:
		if outputDriver == OutputDriverPandoc {
			return nil
		}
		return errors.Errorf("Incompatible output type '%s' with output driver '%s'", outputType, outputDriver)
	case OutputTypeEpub3:
		if outputDriver == OutputDriverPandoc ||
			outputDriver == OutputDriverInternal {
			return nil
		}
		return errors.Errorf("Incompatible output type '%s' with output driver '%s'", outputType, outputDriver)
	case OutputTypeStatsJson:
		return nil
	case OutputTypeStatsTxt:
		return nil
	}

	return errors.Errorf("Unknown output type '%s'", outputType)
}

func (c *Configuration) Print() {
	jsonBytes, err := json.MarshalIndent(c, "", "  ")
	sigolo.FatalCheck(err)
	sigolo.Debugf("Configuration:\n%s", string(jsonBytes))
}

func (c *Configuration) ShouldConvertSvgToPng() bool {
	return c.CommandTemplateSvgToPng != ""
}

func (c *Configuration) ShouldConvertPdfToPng() bool {
	return c.CommandTemplatePdfToPng != ""
}

func (c *Configuration) ShouldConvertWebpToPng() bool {
	return c.CommandTemplateWebpToPng != ""
}

func LoadConfig(file string) error {
	sigolo.Debugf("Load config from file %s", file)

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
