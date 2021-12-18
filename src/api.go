package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type WikiPageDto struct {
	Parse WikiParsePageDto `json:"parse"`
}

type WikiParsePageDto struct {
	Title    string              `json:"title"`
	Wikitext WikiWildcardTextDto `json:"wikitext"`
}

type WikiWildcardTextDto struct {
	Content string `json:"*"`
}

type WikiExpandedTemplateDto struct {
	ExpandTemplate WikitextDto `json:"expandtemplates"`
}

type WikitextDto struct {
	Content string `json:"wikitext"`
}

var imageSources = []string{"commons", "de"}

func downloadPage(language string, title string) (*WikiPageDto, error) {
	escapedTitle := strings.ReplaceAll(title, " ", "_")
	escapedTitle = url.QueryEscape(escapedTitle)
	urlString := fmt.Sprintf("https://%s.wikipedia.org/w/api.php?action=parse&prop=wikitext&format=json&page=%s", language, escapedTitle)
	response, err := http.Get(urlString)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to download article content of article "+title)
	}
	if response.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Downloading article %s failed with status code %d für url %s", title, response.StatusCode, urlString))
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read body bytes")
	}

	wikiPageDto := &WikiPageDto{}
	json.Unmarshal(bodyBytes, wikiPageDto)

	return wikiPageDto, nil
}

func downloadImages(images []string, outputFolder string) error {
	for _, image := range images {
		var err error = nil
		var outputFilepath string

		for _, source := range imageSources {
			outputFilepath, err = downloadImage(image, outputFolder, source)
			if err != nil {
				sigolo.Error("Error downloading image %s from source %s: %s. Try next source.", image, source, err.Error())
				continue
			}

			// If the file is new, rescale it using ImageMagick.
			if outputFilepath != "" {
				const imgSize = 600
				cmd := exec.Command("convert", outputFilepath, "-colorspace", "gray", "-separate", "-average", "-resize", fmt.Sprintf("%dx%d>", imgSize, imgSize), "-quality", "75",
					"-define", "PNG:compression-level=9", "-define", "PNG:compression-filter=0", outputFilepath)
				err = cmd.Run()
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("Error rescaling image %s", outputFilepath))
				}

				//err = os.Rename(outputFilepath+".tmp", outputFilepath)
				//if err != nil {
				//	return errors.Wrap(err, fmt.Sprintf("Unable to rename rescaled temporary image %s.tmp", outputFilepath))
				//}
			}

			break
		}

		if err != nil {
			return err
		}
	}
	return nil
}

// downloadImage downloads the given image (e.g. "File:foo.jpg") to the given folder.
// When the file already exists, nothing is done and "", nil will be returned.
// When the file has been downloaded "filename", nil will be returned.
func downloadImage(fileDescriptor string, outputFolder string, source string) (string, error) {
	filename := strings.Split(fileDescriptor, ":")[1]
	md5sum := fmt.Sprintf("%x", md5.Sum([]byte(filename)))
	sigolo.Debug(filename)
	sigolo.Debug(md5sum)

	url := fmt.Sprintf("https://upload.wikimedia.org/wikipedia/%s/%c/%c%c/%s", source, md5sum[0], md5sum[0], md5sum[1], filename)
	sigolo.Debug(url)

	// Create the output folder
	err := os.Mkdir(outputFolder, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return "", errors.Wrap(err, fmt.Sprintf("Unable to create output folder %s", outputFolder))
	}

	// If file exists -> ignore
	outputFilepath := filepath.Join(outputFolder, "/", filename)
	if _, err := os.Stat(outputFilepath); err == nil {
		sigolo.Info("Image file %s does already exist. Skip.", outputFilepath)
		return "", nil
	}

	// Get the data
	var response *http.Response
	for {
		response, err = http.Get(url)
		if err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("Unable to get image %s with url %s", fileDescriptor, url))
		}
		defer response.Body.Close()

		// Handle 429 (too many requests): wait a bit and retry
		if response.StatusCode == 429 {
			time.Sleep(2 * time.Second)
			continue
		} else if response.StatusCode != 200 {
			return "", errors.New(fmt.Sprintf("Downloading image %s failed with status code %d for url %s", filename, response.StatusCode, url))
		}

		break
	}

	// Create the output file
	outputFile, err := os.Create(outputFilepath)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Unable to create output file for image %s", fileDescriptor))
	}
	defer outputFile.Close()

	// Write the body to file
	_, err = io.Copy(outputFile, response.Body)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Unable copy downloaded content to output file %s", outputFilepath))
	}

	sigolo.Info("Saved image to %s", outputFilepath)

	return outputFilepath, nil
}

func evaluateTemplate(template string) (string, error) {
	sigolo.Info("Evaluate template %s", template)

	urlString := "https://de.wikipedia.org/w/api.php?action=expandtemplates&format=json&prop=wikitext&text=" + url.QueryEscape(template)
	sigolo.Debug("Call %s", urlString)

	response, err := http.Get(urlString)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Unable to call evaluation URL for template %s", template))
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("Response returned with status code %d", response.StatusCode))
	}

	evaluatedTemplateString, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", errors.Wrap(err, "Reading response body failed")
	}

	evaluatedTemplate := &WikiExpandedTemplateDto{}
	json.Unmarshal(evaluatedTemplateString, evaluatedTemplate)

	return evaluatedTemplate.ExpandTemplate.Content, nil
}
