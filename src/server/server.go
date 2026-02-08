package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"wiki2book/cache"
	"wiki2book/config"
	"wiki2book/generator"
	"wiki2book/util"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

const (
	pathVarArticleName = "articleName"
)

func Start() {
	mux := http.NewServeMux()

	mux.HandleFunc(fmt.Sprintf("GET /article/{%s}", pathVarArticleName), handleArticleRequest)

	sigolo.Infof("Start HTTP server on port %d", config.Current.ServerPort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", config.Current.ServerPort), mux)
	sigolo.FatalCheck(errors.Wrapf(err, "Error starting HTTP server on port %d", config.Current.ServerPort))
}

func handleArticleRequest(resp http.ResponseWriter, req *http.Request) {
	articleName := req.PathValue(pathVarArticleName)

	sigolo.Debugf("Received request %s %s for article %s", req.Method, req.URL, articleName)

	// Ensure output folder exists
	outputFolderPath := cache.GetDirPathInCache(cache.TempDirName)
	sigolo.Tracef("Ensure cache folder '%s'", outputFolderPath)
	err := util.CurrentFilesystem.MkdirAll(outputFolderPath)
	if err != nil && !os.IsExist(err) {
		sigolo.Errorf("%+v", errors.Wrapf(err, "Error folder for temporary files"))
		returnInternalServerError(resp, "Error creating folder for temporary files")
		return
	}

	// Create the output file
	sanitizedFilename := util.SanitizeFilename(articleName)
	tempFile, err := util.CurrentFilesystem.CreateTemp(cache.GetTempPath(), sanitizedFilename)
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrapf(err, "Error creating temporary file for article '%s'", articleName))
		returnInternalServerError(resp, fmt.Sprintf("Error creating temporary file for article '%s'", articleName))
		return
	}
	defer tempFile.Close()
	tempFilepath := tempFile.Name()
	defer util.CurrentFilesystem.Remove(tempFilepath)
	sigolo.Tracef("Create temp file '%s'", tempFilepath)

	generator.GenerateArticleEbook(articleName, tempFilepath)

	returnFile(resp, tempFilepath, articleName)
}

func returnFile(resp http.ResponseWriter, filePath string, articleName string) {
	fileContent, err := util.CurrentFilesystem.ReadFile(filePath)
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrapf(err, "Error reading file '%s' for article '%s'", filePath, articleName))
		returnInternalServerError(resp, fmt.Sprintf("An error occurred while creating the response for article '%s'", articleName))
		return
	}

	resp.Header().Set("Content-Type", "application/octet-stream")
	resp.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=\"%s\"", filepath.Base(filePath)))
	resp.WriteHeader(http.StatusOK)

	_, err = resp.Write(fileContent)
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrapf(err, "Could not write response for file '%s': %+v", filePath, err))
		return
	}
}

func returnInternalServerError(resp http.ResponseWriter, errorMessage string) {
	resp.WriteHeader(http.StatusInternalServerError)
	_, err := resp.Write([]byte(fmt.Sprintf("Internal server error: %s", errorMessage)))
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrap(err, "Could not write internal server error response"))
		return
	}
}
