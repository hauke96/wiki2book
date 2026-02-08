package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"wiki2book/cache"
	"wiki2book/config"
	"wiki2book/generator"
	"wiki2book/util"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

const (
	pathVarArticleName = "articleName"
	pathVarResultToken = "resultToken"

	ResultStatusInProgress = "IN_PROGRESS"
	ResultStatusSuccess    = "SUCCESS"
	ResultStatusFailed     = "FAILED"
)

var (
	// Map from result-token to the filename of the epub file.
	resultStates = map[string]*ResultState{}
)

type ResultState struct {
	Status      string `json:"status"`
	ArticleName string `json:"article-name"`
	ResultToken string `json:"result-token"`
	resultPath  string // No JSON mapping. This should not be visible to API users.
}

func Start() {
	mux := http.NewServeMux()

	mux.HandleFunc(fmt.Sprintf("GET /article/{%s}", pathVarArticleName), handleArticleRequest)
	mux.HandleFunc(fmt.Sprintf("GET /states/{%s}", pathVarResultToken), handleGetStateRequest)
	mux.HandleFunc(fmt.Sprintf("GET /results/{%s}", pathVarResultToken), handleGetResultRequest)

	sigolo.Infof("Start HTTP server on port %d", config.Current.ServerPort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", config.Current.ServerPort), mux)
	sigolo.FatalCheck(errors.Wrapf(err, "Error starting HTTP server on port %d", config.Current.ServerPort))
}

func handleArticleRequest(resp http.ResponseWriter, req *http.Request) {
	articleName := req.PathValue(pathVarArticleName)
	sigolo.Debugf("Received request %s %s for article %s", req.Method, req.URL, articleName)

	resultState := createNewResultState(articleName)

	// Ensure output folder exists
	outputFolderPath := cache.GetDirPathInCache(cache.TempDirName)
	sigolo.Tracef("Ensure cache folder '%s'", outputFolderPath)
	err := util.CurrentFilesystem.MkdirAll(outputFolderPath)
	if err != nil && !os.IsExist(err) {
		resultState.Status = ResultStatusFailed
		sigolo.Errorf("%+v", errors.Wrapf(err, "Error folder for temporary files"))
		returnInternalServerError(resp, "Error creating folder for temporary files")
		return
	}

	// Create the output file
	sanitizedFilename := util.SanitizeFilename(articleName)
	tempFile, err := util.CurrentFilesystem.CreateTemp(cache.GetTempPath(), sanitizedFilename)
	if err != nil {
		resultState.Status = ResultStatusFailed
		sigolo.Errorf("%+v", errors.Wrapf(err, "Error creating temporary file for article '%s'", articleName))
		returnInternalServerError(resp, fmt.Sprintf("Error creating temporary file for article '%s'", articleName))
		return
	}
	defer tempFile.Close()
	tempFilepath := tempFile.Name()
	defer util.CurrentFilesystem.Remove(tempFilepath)
	sigolo.Tracef("Create temp file '%s'", tempFilepath)

	go func() {
		generator.GenerateArticleEbook(articleName, tempFilepath)
		resultState.Status = ResultStatusFailed
	}()

	returnState(resp, resultState)
}

func handleGetStateRequest(resp http.ResponseWriter, req *http.Request) {
	resultToken := req.PathValue(pathVarResultToken)
	sigolo.Debugf("Received request %s %s for token %s", req.Method, req.URL, resultToken)

	resultState, ok := resultStates[resultToken]
	if !ok {
		returnNotFound(resp, fmt.Sprintf("Result state for token '%s' not found", resultToken))
		return
	}

	returnState(resp, resultState)
}

func handleGetResultRequest(resp http.ResponseWriter, req *http.Request) {
	resultToken := req.PathValue(pathVarResultToken)
	sigolo.Debugf("Received request %s %s for token %s", req.Method, req.URL, resultToken)

	resultState, ok := resultStates[resultToken]
	if !ok {
		returnNotFound(resp, fmt.Sprintf("Result state for token '%s' not found", resultToken))
		return
	}
	if resultState.Status != ResultStatusSuccess {
		returnNotFound(resp, fmt.Sprintf("Result for token '%s' is not ready yet or has failed", resultToken))
		return
	}

	returnFile(resp, resultState.resultPath, resultState.ArticleName)
}

func createNewResultState(articleName string) *ResultState {
	resultToken := util.Hash(fmt.Sprintf("%s%d", articleName, time.Now().UnixNano()))
	resultState := &ResultState{
		Status:      ResultStatusInProgress,
		ArticleName: articleName,
		ResultToken: resultToken,
		resultPath:  "",
	}
	resultStates[resultToken] = resultState
	return resultState
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

func returnState(resp http.ResponseWriter, state *ResultState) {
	content, err := json.Marshal(state)
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrapf(err, "Error marshalling state to JSON: %#v", state))
		returnInternalServerError(resp, fmt.Sprintf("An error occurred while creating the status response for article '%s'", state.ArticleName))
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)

	_, err = resp.Write(content)
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrapf(err, "Could not write response for result state with token '%s': %+v", state.ResultToken, err))
		return
	}
}

func returnInternalServerError(resp http.ResponseWriter, errorMessage string) {
	resp.Header().Set("Content-Type", "application/text")
	resp.WriteHeader(http.StatusInternalServerError)
	_, err := resp.Write([]byte(fmt.Sprintf("Internal server error: %s", errorMessage)))
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrap(err, "Could not write internal server error response"))
		return
	}
}

func returnNotFound(resp http.ResponseWriter, errorMessage string) {
	resp.Header().Set("Content-Type", "application/text")
	resp.WriteHeader(http.StatusNotFound)
	_, err := resp.Write([]byte(fmt.Sprintf("Not found: %s", errorMessage)))
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrap(err, "Could not write not found response"))
		return
	}
}
