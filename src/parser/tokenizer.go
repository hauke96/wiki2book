package parser

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"regexp"
	"sort"
	"strings"
)

const FILE_PREFIXES = "Datei|File|Bild|Image|Media"

const TOKEN_REGEX = `\$\$TOKEN_([A-Z_0-9]*)_\d+\$\$`
const TOKEN_LINE_REGEX = "^" + TOKEN_REGEX + "$"
const TOKEN_TEMPLATE = "$$TOKEN_%s_%d$$"

const TOKEN_HEADING_TEMPLATE = "HEADING_%d"
const TOKEN_HEADING_1 = "HEADING_1"
const TOKEN_HEADING_2 = "HEADING_2"
const TOKEN_HEADING_3 = "HEADING_3"
const TOKEN_HEADING_4 = "HEADING_4"
const TOKEN_HEADING_5 = "HEADING_5"
const TOKEN_HEADING_6 = "HEADING_6"

const TOKEN_INTERNAL_LINK = "INTERNAL_LINK"
const TOKEN_INTERNAL_LINK_ARTICLE = "INTERNAL_LINK_ARTICLE"
const TOKEN_INTERNAL_LINK_TEXT = "INTERNAL_LINK_TEXT"

const TOKEN_EXTERNAL_LINK = "EXTERNAL_LINK"
const TOKEN_EXTERNAL_LINK_URL = "EXTERNAL_LINK_URL"
const TOKEN_EXTERNAL_LINK_TEXT = "EXTERNAL_LINK_TEXT"

const TOKEN_TABLE = "TABLE"
const TOKEN_TABLE_CAPTION = "TABLE_CAPTION"
const TOKEN_TABLE_HEAD = "TABLE_HEAD"
const TOKEN_TABLE_ROW = "TABLE_ROW"
const TOKEN_TABLE_COL = "TABLE_COL"
const TOKEN_TABLE_COL_ATTRIBUTES = "TABLE_COL_ATTRIB"

const TOKEN_UNORDERED_LIST = "UNORDERED_LIST"
const TOKEN_ORDERED_LIST = "ORDERED_LIST"
const TOKEN_LIST_ITEM = "LIST_ITEM"

const TOKEN_DESCRIPTION_LIST = "DESCRIPTION_LIST"
const TOKEN_DESCRIPTION_LIST_HEAD = "DESCRIPTION_LIST_HEAD" // Head of each description list
const TOKEN_DESCRIPTION_LIST_ITEM = "DESCRIPTION_LIST_ITEM" // Item of a description list (the things with indentation)

const TOKEN_IMAGE = "IMAGE"
const TOKEN_IMAGE_INLINE = "IMAGE_INLINE"
const TOKEN_IMAGE_FILENAME = "IMAGE_FILENAME"
const TOKEN_IMAGE_CAPTION = "IMAGE_CAPTION"
const TOKEN_IMAGE_SIZE = "IMAGE_SIZE"

const TOKEN_REF_USAGE = "REF_USAGE"
const TOKEN_REF_DEF = "REF_DEF"

const TOKEN_MATH = "REF_MATH"

// Marker do not appear in the token map. A marker does not contain further information, it just marks e.g. the start
// and end of a primitive block of content (like a block of bold text)
const MARKER_BOLD_OPEN = "$$MARKER_BOLD_OPEN$$"
const MARKER_BOLD_CLOSE = "$$MARKER_BOLD_CLOSE$$"
const MARKER_ITALIC_OPEN = "$$MARKER_ITALIC_OPEN$$"
const MARKER_ITALIC_CLOSE = "$$MARKER_ITALIC_CLOSE$$"
const MARKER_PARAGRAPH = "$$MARKER_PARAGRAPH$$"

var (
	imagePrefixRegex                 = regexp.MustCompile("(?i)^(" + FILE_PREFIXES + "):")
	galleryStartRegex                = regexp.MustCompile(`^<gallery.*?>`)
	imagemapStartRegex               = regexp.MustCompile(`^<imagemap.*?>`)
	hasNonInlineParameterRegex       = regexp.MustCompile("(" + strings.Join(imageNonInlineParameters, "|") + ")")
	tableStartRegex                  = regexp.MustCompile(`^(:*)(\{\|.*)`)
	tableRowAndColspanRegex          = regexp.MustCompile(`(colspan|rowspan)="(\d+)"`)
	tableTextAlignRegex              = regexp.MustCompile(`text-align:.+?;`)
	listPrefixRegex                  = regexp.MustCompile(`^([*#:;])`)
	referenceBlockStartRegex         = regexp.MustCompile(`</?references.*?/?>\n?`)
	namedReferenceRegex              = regexp.MustCompile(`<ref[^>]*?name="?([^"^>^/]*)"?([^>]*?=[^>]*?)* ?>((.|\n)*?)</ref>`) // Accept all <ref...name=abc...>...</ref> occurrences. There may be more parameters than "name=..." so we have to consider them as well.
	namedReferenceWithoutGroupsRegex = regexp.MustCompile(`<ref[^>]*?name="?([^"^>^/]*)"?>.*?</ref>`)
	namedReferenceUsageRegex         = regexp.MustCompile(`<ref name="?([^"^>^/]*)"?\s?/>`)
	unnamedReferenceRegex            = regexp.MustCompile(`<ref[^>^/]*?>((.|\n)*?)</ref>`)
	mathRegex                        = regexp.MustCompile(`<math.*?>((.|\n|\r)*?)</math>`)
	tokenLineRegex                   = regexp.MustCompile(TOKEN_LINE_REGEX)
)

type Tokenizer struct {
	tokenMap       map[string]string
	tokenCounter   int
	imageFolder    string
	templateFolder string

	tokenizeContent func(tokenizer *Tokenizer, content string) string
}

type Article struct {
	Title    string
	Content  string
	TokenMap map[string]string
	Images   []string
}

func NewTokenizer(imageFolder string, templateFolder string) Tokenizer {
	return Tokenizer{
		tokenMap:       map[string]string{},
		tokenCounter:   0,
		imageFolder:    imageFolder,
		templateFolder: templateFolder,

		tokenizeContent: tokenizeContent,
	}
}

func (t *Tokenizer) Tokenize(content string, title string) Article {
	content = t.tokenizeContent(t, content)

	sigolo.Debug("Token map length: %d", len(t.getTokenMap()))

	// print some debug information if wanted
	if sigolo.LogLevel <= sigolo.LOG_DEBUG {
		sigolo.Debug(content)

		keys := make([]string, 0, len(t.getTokenMap()))
		for k := range t.getTokenMap() {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			sigolo.Debug("%s : %s", k, t.getTokenMap()[k])
		}
	}

	return Article{
		Title:    title,
		TokenMap: t.getTokenMap(),
		Images:   images,
		Content:  content,
	}
}

func (t *Tokenizer) getTokenMap() map[string]string {
	return t.tokenMap
}

func (t *Tokenizer) getToken(tokenType string) string {
	token := fmt.Sprintf(TOKEN_TEMPLATE, tokenType, t.tokenCounter)
	t.tokenCounter++
	return token
}

func (t *Tokenizer) setToken(key string, tokenContent string) {
	t.setRawToken(key, t.tokenizeContent(t, tokenContent))
}

func (t *Tokenizer) setRawToken(key string, tokenContent string) {
	t.tokenMap[key] = tokenContent
}

// tokenizeContent takes a string and tokenizes it. After this call, the Tokenizer.tokenMap will be filled and parts
// of the input will be replaced by token strings.
func tokenizeContent(t *Tokenizer, content string) string {
	for {
		originalContent := content

		content = clean(content)
		content = t.evaluateTemplates(content)
		content = clean(content)

		content = t.parseBoldAndItalic(content)
		content = t.parseHeadings(content)
		content = t.parseReferences(content)
		content = t.parseInternalLinks(content)

		content = escapeImages(content)
		content = t.parseGalleries(content)
		content = t.parseImageMaps(content)
		content = t.parseImages(content)

		content = t.parseExternalLinks(content)
		content = t.parseMath(content)
		content = t.parseLists(content)
		content = t.parseTables(content)
		content = t.parseParagraphs(content)

		if content == originalContent {
			break
		}
	}

	return content
}

// tokenizeInline is meant for strings that are known to be inline string. Example: The text of an internal link cannot
// contain tables and lists, so we do not want to parse them.
func (t *Tokenizer) tokenizeInline(content string) string {
	content = t.parseBoldAndItalic(content)
	content = t.parseHeadings(content)
	content = t.parseReferences(content)

	content = t.parseInternalLinks(content)
	content = t.parseImages(content)
	content = t.parseExternalLinks(content)

	return content
}
