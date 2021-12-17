package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/wiki"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

func downloadImages(images []wiki.Image, outputFolder string) error {
	for _, image := range images {
		err := downloadImage(image.Filename, outputFolder)
		if err != nil {
			return err
		}
	}
	return nil
}

// Download the given image (e.g. "File:foo.jpg") to the given folder
func downloadImage(fileDescriptor string, outputFolder string) error {
	filename := strings.Split(fileDescriptor, ":")[1]
	md5sum := fmt.Sprintf("%x", md5.Sum([]byte(filename)))
	sigolo.Debug(filename)
	sigolo.Debug(md5sum)

	url := fmt.Sprintf("https://upload.wikimedia.org/wikipedia/commons/%c/%c%c/%s", md5sum[0], md5sum[0], md5sum[1], filename)
	sigolo.Debug(url)

	// Create the output folder
	err := os.Mkdir(outputFolder, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, fmt.Sprintf("Unable to create output folder %s", outputFolder))
	}

	// If file exists -> ignore
	outputFilepath := filepath.Join(outputFolder, "/", filename)
	if _, err := os.Stat(outputFilepath); err == nil {
		return nil
	}

	// Create the output file
	outputFile, err := os.Create(outputFilepath)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to create output file for image %s", fileDescriptor))
	}
	defer outputFile.Close()

	// Get the data
	response, err := http.Get(url)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to get image %s", fileDescriptor))
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Downloading image %s failed with status code %d", filename, response.StatusCode))
	}

	// Write the body to file
	_, err = io.Copy(outputFile, response.Body)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable copy downloaded content to output file %s", outputFilepath))
	}

	sigolo.Info("Saved image to %s", outputFilepath)
	return nil
}
