package config

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"os"
)

// Current config initialized with default values, which allows wiki2book to run without any specified config file.
var Current = &Configuration{
	IgnoredTemplates:               []string{},
	TrailingTemplates:              []string{},
	IgnoredImageParams:             []string{},
	IgnoredMediaTypes:              []string{"gif", "mp3", "mp4", "pdf", "oga", "ogg", "ogv", "wav", "webm"},
	WikipediaInstance:              "en",
	WikipediaImageArticleInstances: []string{"commons", "en"},
	FilePrefixe:                    []string{"file", "image", "media"},
	AllowedLinkPrefixes:            []string{"arxiv", "doi"},
	CategoryPrefixes:               []string{"category"},
}

// Configuration is a struct with application-wide configurations and language-specific strings (e.g. templates to
// ignore). Some configurations are mandatory, which means that wiki2book will definitely crash if the config entry is
// not given. Entries marked as non-mandatory may also cause a crash.
// The configuration differs from a project-config by the following rule of thumb: This contains technical and project-
// independent stuff. Some properties, though, might exist in both, this Configuration and the project.Project struct.
type Configuration struct {
	/*
		List of templates that should be ignored and removed from the input wikitext. The list must be in lower case.

		Default: Empty list
		Mandatory: No

		JSON example: "ignored-templates": [ "foo", "bar" ]
		This ignores {{foo}} and {{bar}} occurrences in the input text.
	*/
	IgnoredTemplates []string `json:"ignored-templates"`

	/*
		List of templates that will be moved to the end of the document. Theses are e.g. remarks on the article that
		are important but should be shown as a remark after the actual content of the article.

		Default: Empty list
		Mandatory: No

		JSON example: "trailing-templates": [ "foo", "bar" ]
		This moves {{foo}} and {{bar}} to the end of the document.
	*/
	TrailingTemplates []string `json:"trailing-templates"`

	/*
		Parameters of images that should be ignored. The list must be in lower case.

		Default: Empty list
		Mandatory: No

		JSON example: "ignored-image-params": [ "alt", "center" ]
		This ignores the image parameters "alt" and "center" including any parameter values like "alt"="some alt text".
	*/
	IgnoredImageParams []string `json:"ignored-image-params"`

	/*
		List of media types to ignore, i.e. list of file extensions. Some media types (e.g. videos) are not of much use
		for a book.

		Default: [ "gif", "mp3", "mp4", "pdf", "oga", "ogg", "ogv", "wav", "webm" ]
		Mandatory: No
	*/
	IgnoredMediaTypes []string `json:"ignored-media-types"`

	/*
		The URL to the Wikipedia (or generally MediaWiki) instance.

		Default: "en"
		Mandatory: Yes

		JSON example: "wikipedia-instance": "de"
		This config uses the German Wikipedia.
	*/
	WikipediaInstance string `json:"wikipedia-instance"`

	/*
		Each image has its own article, which is fetched from these Wikipedia instances (in the given order).

		Default: [ "commons", "en" ]
		Mandatory: Yes

		JSON example: "wikipedia-image-article-instances": [ "commons", "de" ]
	*/
	WikipediaImageArticleInstances []string `json:"wikipedia-image-article-instances"`

	/*
		A list of prefixes for files, e.g. in "File:picture.jpg" the substring "File" is the image prefix. The list
		must be in lower case.

		Default: [ "file", "image", "media" ]
		Mandatory: No

		JSON example: "file-prefixe": [ "file", "datei" ]
	*/
	FilePrefixe []string `json:"file-prefixe"`

	/*
		A list of link prefixes that are allowed. All prefixes  specified by "FilePrefixe" are considered to be allowed
		prefixes. Any other not explicitly allowed prefix of a link causes the link to get removed. This especially
		happens for inter-wiki-links if the Wikipedia instance is not explicitly allowed using this list.

		Default: [ "arxiv", "doi" ]
		Mandatory: No
	*/
	AllowedLinkPrefixes []string `json:"allowed-link-prefixe"`

	/*
		A list of category prefixes, which are technically internals links. However, categories will be removed from
		the input wikitext.

		Default: [ "category" ]
		Mandatory: No
	*/
	CategoryPrefixes []string `json:"category-prefixes"`
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

	return nil
}
