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

type Server struct {
	configService *config.ConfigService
	fileCache     *cache.Cache
}

func NewServer(configService *config.ConfigService, fileCache *cache.Cache) *Server {
	return &Server{
		configService: configService,
		fileCache:     fileCache,
	}
}

func (s *Server) Start() {
	mux := http.NewServeMux()

	mux.HandleFunc(fmt.Sprintf("GET /article/{%s}", pathVarArticleName), s.handleArticleGetRequest)
	mux.HandleFunc(fmt.Sprintf("POST /article/{%s}", pathVarArticleName), s.handleArticlePostRequest)
	// TODO POST handler for projects
	// TODO POST handle for standalone
	mux.HandleFunc(fmt.Sprintf("GET /states/{%s}", pathVarResultToken), s.handleGetStateRequest)
	mux.HandleFunc(fmt.Sprintf("GET /results/{%s}", pathVarResultToken), s.handleGetResultRequest)

	sigolo.Infof("Start HTTP server on port %d", s.configService.Get().ServerPort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.configService.Get().ServerPort), mux)
	sigolo.FatalCheck(errors.Wrapf(err, "Error starting HTTP server on port %d", s.configService.Get().ServerPort))
}

func (s *Server) handleArticleGetRequest(resp http.ResponseWriter, req *http.Request) {
	articleName := req.PathValue(pathVarArticleName)
	sigolo.Debugf("Received request %s %s for article %s", req.Method, req.URL, articleName)

	resultState := s.createNewResultState(articleName)

	s.handleArticleRequest(resp, resultState, s.configService)
}

func (s *Server) handleArticlePostRequest(resp http.ResponseWriter, req *http.Request) {
	articleName := req.PathValue(pathVarArticleName)
	sigolo.Debugf("Received request %s %s for article %s", req.Method, req.URL, articleName)

	resultState := s.createNewResultState(articleName)

	currentConfig := config.NewDefaultConfig()
	currentConfig.MergeNonDefaultValues(s.configService.Get())

	// Read body to current config. Fields not set by the given request-config stay unchanged, so only the fields that
	// are present in the request-config will be set here.
	err := json.NewDecoder(req.Body).Decode(currentConfig)
	if err != nil {
		resultState.Status = ResultStatusFailed
		sigolo.Errorf("%+v", errors.Wrapf(err, "Error reading request body"))
		s.returnInternalServerError(resp, "Error reading request body")
		return
	}

	// Restore certain config entries that should not be set by users of the API:
	// TODO test this logic
	currentConfig.ForceRegenerateHtml = s.configService.Get().ForceRegenerateHtml
	// Should not be set by user: currentConfig.SvgSizeToViewbox
	// Should not be set by user: currentConfig.OutputType
	currentConfig.OutputDriver = s.configService.Get().OutputDriver
	currentConfig.CacheDir = s.configService.Get().CacheDir
	currentConfig.CacheMaxSize = s.configService.Get().CacheMaxSize
	currentConfig.CacheMaxAge = s.configService.Get().CacheMaxAge
	currentConfig.CacheEvictionStrategy = s.configService.Get().CacheEvictionStrategy
	currentConfig.StyleFile = s.configService.Get().StyleFile
	currentConfig.CoverImage = s.configService.Get().CoverImage
	currentConfig.CommandTemplateSvgToPng = s.configService.Get().CommandTemplateSvgToPng
	currentConfig.CommandTemplateMathSvgToPng = s.configService.Get().CommandTemplateMathSvgToPng
	currentConfig.CommandTemplateImageProcessing = s.configService.Get().CommandTemplateImageProcessing
	currentConfig.CommandTemplatePdfToPng = s.configService.Get().CommandTemplatePdfToPng
	currentConfig.CommandTemplateWebpToPng = s.configService.Get().CommandTemplateWebpToPng
	currentConfig.PandocExecutable = s.configService.Get().PandocExecutable
	currentConfig.PandocDataDir = s.configService.Get().PandocDataDir
	currentConfig.FontFiles = s.configService.Get().FontFiles
	// Should not be set by user: currentConfig.IgnoredTemplates
	// Should not be set by user: currentConfig.TrailingTemplates
	// Should not be set by user: currentConfig.IgnoredImageParams
	// Should not be set by user: currentConfig.IgnoredMediaTypes
	// Should not be set by user: currentConfig.WikipediaInstance
	// Should not be set by user: currentConfig.WikipediaHost
	// Should not be set by user: currentConfig.WikipediaImageHost
	// Should not be set by user: currentConfig.WikipediaImageArticleHosts
	// Should not be set by user: currentConfig.WikipediaMathRestApi
	// Should not be set by user: currentConfig.FilePrefixes
	// Should not be set by user: currentConfig.AllowedLinkPrefixes
	// Should not be set by user: currentConfig.CategoryPrefixes
	currentConfig.MathConverter = s.configService.Get().MathConverter
	// Should not be set by user: currentConfig.TocDepth
	currentConfig.WorkerThreads = s.configService.Get().WorkerThreads
	currentConfig.UserAgentTemplate = s.configService.Get().UserAgentTemplate
	currentConfig.ServerPort = s.configService.Get().ServerPort

	configServiceForRequest := config.NewConfigServiceForConfig(currentConfig)

	s.handleArticleRequest(resp, resultState, configServiceForRequest)
}

func (s *Server) handleArticleRequest(resp http.ResponseWriter, resultState *ResultState, configService *config.ConfigService) {
	// Ensure output folder exists
	outputFolderPath := s.fileCache.GetDirPathInCache(cache.TempDirName)
	sigolo.Tracef("Ensure cache folder '%s'", outputFolderPath)
	err := util.CurrentFilesystem.MkdirAll(outputFolderPath)
	if err != nil && !os.IsExist(err) {
		resultState.Status = ResultStatusFailed
		sigolo.Errorf("%+v", errors.Wrapf(err, "Error folder for temporary files"))
		s.returnInternalServerError(resp, "Error creating folder for temporary files")
		return
	}

	// Create the output file
	sanitizedFilename := util.SanitizeFilename(resultState.ArticleName)
	tempFile, err := util.CurrentFilesystem.CreateTemp(s.fileCache.GetTempPath(), sanitizedFilename)
	if err != nil {
		resultState.Status = ResultStatusFailed
		sigolo.Errorf("%+v", errors.Wrapf(err, "Error creating temporary file for article '%s'", resultState.ArticleName))
		s.returnInternalServerError(resp, fmt.Sprintf("Error creating temporary file for article '%s'", resultState.ArticleName))
		return
	}
	defer tempFile.Close()
	tempFilepath := tempFile.Name()
	defer util.CurrentFilesystem.Remove(tempFilepath)
	sigolo.Tracef("Create temp file '%s'", tempFilepath)

	go func() {
		ebookGeneratorService := generator.NewEbookGenerator(configService, s.fileCache)
		ebookGeneratorService.GenerateArticleEbook(resultState.ArticleName, tempFilepath)
		resultState.Status = ResultStatusSuccess
		resultState.resultPath = tempFilepath
	}()

	s.returnState(resp, resultState)
}

func (s *Server) handleGetStateRequest(resp http.ResponseWriter, req *http.Request) {
	resultToken := req.PathValue(pathVarResultToken)
	sigolo.Debugf("Received request %s %s for token %s", req.Method, req.URL, resultToken)

	resultState, ok := resultStates[resultToken]
	if !ok {
		s.returnNotFound(resp, fmt.Sprintf("Result state for token '%s' not found", resultToken))
		return
	}

	s.returnState(resp, resultState)
}

func (s *Server) handleGetResultRequest(resp http.ResponseWriter, req *http.Request) {
	resultToken := req.PathValue(pathVarResultToken)
	sigolo.Debugf("Received request %s %s for token %s", req.Method, req.URL, resultToken)

	resultState, ok := resultStates[resultToken]
	if !ok {
		s.returnNotFound(resp, fmt.Sprintf("Result state for token '%s' not found", resultToken))
		return
	}
	if resultState.Status != ResultStatusSuccess {
		s.returnNotFound(resp, fmt.Sprintf("Result for token '%s' is not ready yet or has failed", resultToken))
		return
	}

	s.returnFile(resp, resultState.resultPath, resultState.ArticleName)
}

func (s *Server) createNewResultState(articleName string) *ResultState {
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

func (s *Server) returnFile(resp http.ResponseWriter, filePath string, articleName string) {
	fileContent, err := util.CurrentFilesystem.ReadFile(filePath)
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrapf(err, "Error reading file '%s' for article '%s'", filePath, articleName))
		s.returnInternalServerError(resp, fmt.Sprintf("An error occurred while creating the response for article '%s'", articleName))
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

func (s *Server) returnState(resp http.ResponseWriter, state *ResultState) {
	content, err := json.Marshal(state)
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrapf(err, "Error marshalling state to JSON: %#v", state))
		s.returnInternalServerError(resp, fmt.Sprintf("An error occurred while creating the status response for article '%s'", state.ArticleName))
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

func (s *Server) returnInternalServerError(resp http.ResponseWriter, errorMessage string) {
	resp.Header().Set("Content-Type", "application/text")
	resp.WriteHeader(http.StatusInternalServerError)
	_, err := resp.Write([]byte(fmt.Sprintf("Internal server error: %s", errorMessage)))
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrap(err, "Could not write internal server error response"))
		return
	}
}

func (s *Server) returnNotFound(resp http.ResponseWriter, errorMessage string) {
	resp.Header().Set("Content-Type", "application/text")
	resp.WriteHeader(http.StatusNotFound)
	_, err := resp.Write([]byte(fmt.Sprintf("Not found: %s", errorMessage)))
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrap(err, "Could not write not found response"))
		return
	}
}
