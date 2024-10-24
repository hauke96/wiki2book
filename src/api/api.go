package api

import (
	"fmt"
	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var httpClient = GetDefaultHttpClient()

// downloadAndCache fires an GET request to the given url and saving the result in cacheFolder/filename. The return
// value is this resulting filepath and a bool (true = file was (tried to be) downloaded, false = file already exists in
// cache) or an error. If the file already exists, no HTTP request is made.
func downloadAndCache(url string, cacheFolder string, filename string) (string, bool, error) {
	// If file exists -> ignore
	outputFilepath := filepath.Join(cacheFolder, filename)
	sigolo.Debugf("Try to find already cached file '%s'", outputFilepath)
	_, err := os.Stat(outputFilepath)
	if err == nil {
		sigolo.Debugf("File %s does already exist. Skip.", outputFilepath)
		return outputFilepath, false, nil
	}
	sigolo.Debug("File not cached, download fresh one")

	// Get the data
	reader, err := download(url, filename)
	if err != nil {
		return "", true, err
	}
	defer reader.Close()

	err = cacheToFile(cacheFolder, filename, reader)
	if err != nil {
		return "", true, errors.Wrapf(err, "Unable to cache to %s", outputFilepath)
	}

	return outputFilepath, true, nil
}

// download returns the open response body of the GET request for the given URL. The article name is just there for
// logging purposes.
func download(url string, filename string) (io.ReadCloser, error) {
	var response *http.Response
	var err error

	for {
		sigolo.Debugf("Make GET request to %s", url)
		response, err = httpClient.Get(url)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to get file %s with url %s", filename, url))
		}

		sigolo.Tracef("Response: %#v", response)

		// Handle 429 (too many requests): wait a bit and retry
		if response.StatusCode == 429 {
			sigolo.Trace("Got 429 response (too many requests). Wait some time and try again...")
			time.Sleep(2 * time.Second)
			continue
		} else if response.StatusCode != 200 {
			return response.Body, errors.Errorf("Downloading file '%s' failed with status code %d for url %s", filename, response.StatusCode, url)
		} else {
			responseErrorHeader := response.Header.Get("mediawiki-api-error")
			if responseErrorHeader != "" {
				return response.Body, errors.Errorf("Downloading file '%s' failed with error header '%s' for url %s", filename, responseErrorHeader, url)
			}
		}

		break
	}
	return response.Body, nil
}

func cacheToFile(cacheFolder string, filename string, reader io.ReadCloser) error {
	// Create the output folder
	sigolo.Debugf("Ensure cache folder '%s'", cacheFolder)
	err := os.MkdirAll(cacheFolder, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, fmt.Sprintf("Unable to create output folder '%s'", cacheFolder))
	}

	outputFilepath := filepath.Join(cacheFolder, filename)

	// Create the output file
	sigolo.Debugf("Create cached file '%s'", outputFilepath)
	outputFile, err := os.Create(outputFilepath)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to create output file for file '%s'", outputFilepath))
	}
	defer outputFile.Close()

	// Write the body to file
	sigolo.Debug("Copy data to cached file")
	_, err = io.Copy(outputFile, reader)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable copy downloaded content to output file '%s'", outputFilepath))
	}

	sigolo.Debugf("Cached file '%s' to '%s'", filename, outputFilepath)

	return nil
}
