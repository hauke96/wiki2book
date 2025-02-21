package api

import (
	"fmt"
	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"wiki2book/util"
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
	responseBodyReader, err := download(url, filename)
	if err != nil {
		logResponseBodyAsError(responseBodyReader, url)
		return "", true, err
	}
	defer responseBodyReader.Close()

	err = cacheToFile(cacheFolder, filename, responseBodyReader)
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
			logResponseBodyAsError(response.Body, url)
			return response.Body, errors.Errorf("Downloading file '%s' failed with status code %d for url %s", filename, response.StatusCode, url)
		} else {
			responseErrorHeader := response.Header.Get("mediawiki-api-error")
			if responseErrorHeader != "" {
				logResponseBodyAsError(response.Body, url)
				return response.Body, errors.Errorf("Downloading file '%s' failed with error header '%s' for url %s", filename, responseErrorHeader, url)
			}
		}

		break
	}
	return response.Body, nil
}

func cacheToFile(cacheFolder string, filename string, reader io.ReadCloser) error {
	outputFilepath := filepath.Join(cacheFolder, filename)
	sigolo.Debugf("Write data to cache file '%s'", outputFilepath)

	// Create the output folder
	sigolo.Tracef("Ensure cache folder '%s'", cacheFolder)
	err := os.MkdirAll(cacheFolder, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, fmt.Sprintf("Unable to create output folder '%s'", cacheFolder))
	}

	//
	// 1. Write to temporary file. This prevents broken files on disk in case the application exits during writing.
	//

	// Create the output file
	tempFile, err := os.CreateTemp(util.TempDirName, filename)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to create temporary file '%s'", filepath.Join(util.TempDirName, filename)))
	}
	tempFilepath := tempFile.Name()
	defer os.Remove(tempFilepath)
	sigolo.Tracef("Create temp file '%s'", tempFilepath)

	// Write the body to file
	sigolo.Trace("Copy data to temp file")
	_, err = io.Copy(tempFile, reader)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable copy downloaded content to temp file '%s'", tempFilepath))
	}

	//
	// 2. Move file to actual location
	//

	sigolo.Tracef("Move temp file '%s' to '%s'", tempFilepath, outputFilepath)
	err = os.Rename(tempFilepath, outputFilepath)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error moving temp file '%s' to '%s'", tempFilepath, outputFilepath))
	}

	sigolo.Tracef("Cached file '%s' to '%s'", filename, outputFilepath)
	return nil
}

func logResponseBodyAsError(bodyReader io.Reader, urlString string) {
	if bodyReader != nil {
		buf := new(strings.Builder)
		_, err := io.Copy(buf, bodyReader)
		if err == nil {
			sigolo.Errorf("Response body for url %s:\n%s", urlString, buf.String())
		}
	}
}
