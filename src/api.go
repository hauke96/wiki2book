package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

type WikiPage struct {
	Parse WikiParsePage `json:"parse"`
}

type WikiParsePage struct {
	Title    string `json:"title"`
	Wikitext Wikitext
}

type Wikitext struct {
	Content string `json:"*"`
}

func downloadPage(title string) (*WikiPage, error) {
	response, err := http.Get("https://de.wikipedia.org/w/api.php?action=parse&prop=wikitext&format=json&page=" + title)
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

	wikiPage := &WikiPage{}
	json.Unmarshal(bodyBytes, wikiPage)

	return wikiPage, nil
}
