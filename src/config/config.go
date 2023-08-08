package config

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"os"
)

var Current = &Configuration{
	IgnoredTemplates:               []string{},
	IgnoredImageParams:             []string{},
	IgnoredMediaTypes:              []string{"gif", "mp3", "mp4", "pdf", "oga", "ogg", "ogv", "wav", "webm"},
	WikipediaInstance:              "en",
	WikipediaImageArticleInstances: []string{"commons", "en"},
	FilePrefixe:                    []string{"file", "image", "media"},
}

// Configuration is a struct with application-wide configurations and language-specific strings (e.g. templates to
// ignore). Some configurations are mandatory, which means that wiki2book will definitely crash if the config entry is
// not given. Entries marked as non-mandatory may also cause a crash.
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
		List of media types to ignore, i.e. list of file extensions. Some media types (e.g. videos) are not of much use
		for a book.

		Default: [ "gif", "mp3", "mp4", "pdf", "oga", "ogg", "ogv", "wav", "webm" ]
		Mandatory: No
	*/
	IgnoredMediaTypes []string `json:"ignored-media-types"`

	/*
		Parameters of images that should be ignored. The list must be in lower case.

		Default: Empty list
		Mandatory: No

		JSON example: "ignored-image-params": [ "alt", "center" ]
		This ignores the image parameters "alt" and "center" including any parameter values like "alt"="some alt text".
	*/
	IgnoredImageParams []string `json:"ignored-image-params"`

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
