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
	(?s)  Flag to let . also match on \n.
	(?i)  Flag to use case-insensitive matching.
*/

// Token
const (
	TOKEN_REGEX      = `\$\$TOKEN_[A-Z_0-9]+_\d+\$\$`
	TOKEN_LINE_REGEX = "^" + TOKEN_REGEX + "$"
	TOKEN_TEMPLATE   = "$$TOKEN_%s_%d$$"
)

var (
	tokenLineRegex = regexp.MustCompile(TOKEN_LINE_REGEX)
)

// Categories, templates, unwanted HTML
var (
	templateNameRegex = regexp.MustCompile(`{{\s*([^\n|}]+)`)
	unwantedHtmlRegex = regexp.MustCompile(`</?(div|span)[^>]*>`)
)

// Links
var (
	internalLinkStartRegex = regexp.MustCompile(`(?s)\[\[([^]]+?):`)
)

// Media files
var (
	galleryStartRegex          = regexp.MustCompile(`^<gallery.*?>`)
	imagemapStartRegex         = regexp.MustCompile(`^<imagemap.*?>`)
	hasNonInlineParameterRegex = regexp.MustCompile("(" + strings.Join(imageNonInlineParameters, "|") + ")")
)

// Tables
var (
	tableStartRegex         = regexp.MustCompile(`^(:*)(\{\|.*)`)
	tableRowAndColspanRegex = regexp.MustCompile(`(colspan|rowspan)="?(\d+)"?`)
	tableTextAlignRegex     = regexp.MustCompile(`text-align:.+?;`)
)

// References
var (
	referencePlaceholderShortRegex = regexp.MustCompile(`<references.*?/\s*>`) // <references />
	referencePlaceholderStartRegex = regexp.MustCompile(`<references.*?\s*>`)  // <references group="foo" >
	referencePlaceholderEndRegex   = regexp.MustCompile(`</references\s*>`)    // </references>
)

// Math
var (
	mathRegex = regexp.MustCompile(`<math.*?>((.|\n|\r)*?)</math>`)
)
