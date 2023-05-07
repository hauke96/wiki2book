package parser

import (
	"regexp"
	"strings"
)

/*
This file contains all important pattern for various things.
Short explanations (because regex is like magic to some people):

	[^x]  Match everything except the character x (or whatever x is).
	*?    Minimum sized match.
	(?m)  Flag to use multi-line matchings to also match on \n.
*/

// Token
const (
	TOKEN_REGEX      = `\$\$TOKEN_([A-Z_0-9]*)_\d+\$\$`
	TOKEN_LINE_REGEX = "^" + TOKEN_REGEX + "$"
	TOKEN_TEMPLATE   = "$$TOKEN_%s_%d$$"
)

var (
	tokenLineRegex = regexp.MustCompile(TOKEN_LINE_REGEX)
)

// Headings
var (
	headingRegex     = regexp.MustCompile(`^(=*)[^=]+(=*)$`)
	semiHeadingRegex = regexp.MustCompile(`^'''.+'''$`)
)

// Categories, templates, unwanted HTML
var (
	categoryRegex     = regexp.MustCompile(`\[\[(Kategorie|Category):[^]]*?]]\n?`)
	templateNameRegex = regexp.MustCompile(`{{\s*([^\n|}]+)`)
	unwantedHtmlRegex = regexp.MustCompile(`</?(div|span)[^>]*>`)
)

// Links
var (
	interwikiLinkRegex = regexp.MustCompile(`\[\[([A-Za-z\-]+:)[^]]+]]`)
	linkPrefixe        = append(imageTypes,
		[]string{
			"arxiv",
			"doi",
			"vorlage",
		}...,
	)
)

// Lists
var (
	// TODO Edge case: Empty list item at the end of the content (with no trailing newline)
	emptyListItemRegex = regexp.MustCompile(`(?m)^(\s*[*#:;]+\s*\n)`)
	listPrefixRegex    = regexp.MustCompile(`^([*#:;])`)
)

// Images
var (
	imageTypes = []string{
		"datei",
		"file",
		"bild",
		"image",
		"media",
	}
	filePrefixes               = strings.Join(imageTypes, "|")
	nonImageRegex              = regexp.MustCompile(`(?i)\[\[((` + filePrefixes + `):.*?\.(gif|mp3|mp4|pdf|oga|ogg|ogv|wav|webm))(\|([^\]]*))?]]`) // TODO Regarding PDFs: This is temporary, solve issue #33 to handle them properly.
	imageRegex                 = regexp.MustCompile(`(?i)\[\[((` + filePrefixes + `):([^|^\]]*))(\|([^\]]*))?]]`)
	imagePrefixRegex           = regexp.MustCompile("(?i)^(" + filePrefixes + "):")
	galleryStartRegex          = regexp.MustCompile(`^<gallery.*?>`)
	imagemapStartRegex         = regexp.MustCompile(`^<imagemap.*?>`)
	hasNonInlineParameterRegex = regexp.MustCompile("(" + strings.Join(imageNonInlineParameters, "|") + ")")
)

// Tables
var (
	tableStartRegex         = regexp.MustCompile(`^(:*)(\{\|.*)`)
	tableRowAndColspanRegex = regexp.MustCompile(`(colspan|rowspan)="(\d+)"`)
	tableTextAlignRegex     = regexp.MustCompile(`text-align:.+?;`)
)

// References
var (
	referenceBlockStartRegex = regexp.MustCompile(`</?references.*?/?>\n?`)
	namedReferenceRegex      = regexp.MustCompile(`<ref[^>]*?name="?([^"^>]*)"?\s*>((.|\n)*?)</ref>`) // Accept all <ref...name=abc...>...</ref> occurrences. There may be more parameters than "name=..." so we have to consider them as well.
	namedReferenceUsageRegex = regexp.MustCompile(`<ref name="?([^"^>]*)"?\s*/>`)
	generalReferenceRegex    = regexp.MustCompile(`<ref[^>]*?(/>|>((.|\n)*?)</ref>)`)
)

// Math
var (
	mathRegex = regexp.MustCompile(`<math.*?>((.|\n|\r)*?)</math>`)
)
