package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

type WikiPageDto struct {
	Parse WikiParsePageDto `json:"parse"`
}

type WikiParsePageDto struct {
	Title    string `json:"title"`
	Wikitext WikitextDto
}

type WikitextDto struct {
	Content string `json:"*"`
}

func downloadPage(language string, title string) (*WikiPageDto, error) {
	url := fmt.Sprintf("https://%s.wikipedia.org/w/api.php?action=parse&prop=wikitext&format=json&page=%s", language, title)
	response, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to download article content of article "+title)
	}
	if response.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Downloading article %s failed with status code %d", title, response.StatusCode))
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read body bytes")
	}

	wikiPageDto := &WikiPageDto{}
	json.Unmarshal(bodyBytes, wikiPageDto)

	return wikiPageDto, nil
}
