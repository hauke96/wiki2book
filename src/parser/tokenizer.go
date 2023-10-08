package parser

import (
	"fmt"
	"github.com/hauke96/sigolo"
)

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
const TOKEN_UNKNOWN_LIST_ITEM = "UNKNOWN_LIST_TYPE_%s" // Template for unknown lists

const TOKEN_DESCRIPTION_LIST = "DESCRIPTION_LIST"
const TOKEN_DESCRIPTION_LIST_HEAD = "DESCRIPTION_LIST_HEAD"      // Head of each description list
const TOKEN_DESCRIPTION_LIST_ITEM = "DESCRIPTION_LIST_ITEM"      // Item of a description list (the things with indentation)
const TOKEN_UNKNOWN_LIST_ITEM_TYPE = "UNKNOWN_LIST_ITEM_TYPE_%s" // Template for unknown list items

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

type Tokenizer struct {
	// TODO Change "interface{}" to "*Token" when everything is migrated
	// TODO Use separate new type "type TokenKey string" for token keys instead of "string"
	tokenMap       map[string]interface{}
	tokenCounter   int
	imageFolder    string
	templateFolder string

	tokenizeContent func(tokenizer *Tokenizer, content string) string // TODO Change "interface{}" to "*Token" when everything is migrated
}

type Article struct {
	Title    string
	Content  string
	TokenMap map[string]interface{}
	Images   []string
}

// Token is the abstract type for all sorts of tokens.
type Token interface{}

// StringToken represents a part of the input data that is pure text.
type StringToken struct {
	Token
	String string
}

func NewTokenizer(imageFolder string, templateFolder string) Tokenizer {
	return Tokenizer{
		tokenMap:       map[string]interface{}{},
		tokenCounter:   0,
		imageFolder:    imageFolder,
		templateFolder: templateFolder,

		tokenizeContent: tokenizeContent,
	}
}

func (t *Tokenizer) Tokenize(content string, title string) (*Article, error) {
	var err error

	sigolo.Info("Tokenize article %s [1/2]: Evaluate templates", title)

	content, err = clean(content)
	if err != nil {
		return nil, err
	}

	content, err = t.evaluateTemplates(content)
	if err != nil {
		return nil, err
	}

	content, err = clean(content)
	if err != nil {
		return nil, err
	}

	sigolo.Info("Tokenize article %s [2/2]: Tokenize content", title)
	content = t.tokenizeContent(t, content)

	sigolo.Debug("Token map length: %d", len(t.getTokenMap()))

	// print some debug information if wanted
	// TODO Print these internal information when --trace or similar has been specified (see also #35)
	//if sigolo.LogLevel <= sigolo.LOG_DEBUG {
	//	sigolo.Debug(content)
	//
	//	keys := make([]string, 0, len(t.getTokenMap()))
	//	for k := range t.getTokenMap() {
	//		keys = append(keys, k)
	//	}
	//	sort.Strings(keys)
	//
	//	for _, k := range keys {
	//		sigolo.Debug("%s : %s", k, t.getTokenMap()[k])
	//	}
	//}

	article := Article{
		Title:    title,
		TokenMap: t.getTokenMap(),
		Images:   images,
		Content:  content,
	}
	return &article, nil
}

func (t *Tokenizer) getTokenMap() map[string]interface{} {
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

func (t *Tokenizer) setRawToken(key string, tokenContent interface{}) {
	t.tokenMap[key] = tokenContent
}

// tokenizeContent takes a string and tokenizes it. After this call, the Tokenizer.tokenMap will be filled and parts
// of the input will be replaced by token strings.
func tokenizeContent(t *Tokenizer, content string) string {
	for {
		originalContent := content

		content = t.parseBoldAndItalic(content)
		content = t.parseHeadings(content)
		content = t.parseReferences(content)
		content = t.parseInternalLinks(content)

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
