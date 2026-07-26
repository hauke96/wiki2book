package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	Title       string `json:"title"`
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
	mux.HandleFunc("POST /project", s.handleProjectPostRequest)
	mux.HandleFunc("POST /standalone", s.handleStandalonePostRequest)
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
	req.Body = http.MaxBytesReader(resp, req.Body, s.configService.Get().ServerMaxRequestBodySize)

	articleName := req.PathValue(pathVarArticleName)
	sigolo.Debugf("Received request %s %s for article %s", req.Method, req.URL, articleName)

	resultState := s.createNewResultState(articleName)

	currentConfig := config.NewDefaultConfig()
	currentConfig.MergeNonDefaultValues(s.configService.Get())

	// Read body to current config. Fields not set by the given request-config stay unchanged, so only the fields that
	// are present in the request-config will be set here.
	err := json.NewDecoder(req.Body).Decode(currentConfig)
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			sigolo.Errorf("%+v", errors.Wrapf(err, "Request too large, only %d bytes allowed", s.configService.Get().ServerMaxRequestBodySize))
			s.returnInternalServerError(resp, resultState, fmt.Sprintf("Request too large, only %d bytes allowed", s.configService.Get().ServerMaxRequestBodySize))
			return
		}

		sigolo.Errorf("%+v", errors.Wrapf(err, "Error reading request body"))
		s.returnInternalServerError(resp, resultState, "Error reading request body")
		return
	}

	// Restore certain config entries that should not be set by users of the API:
	s.resetNonUploadableConfigProperties(currentConfig)

	configServiceForRequest := config.NewConfigServiceForConfig(currentConfig)

	s.handleArticleRequest(resp, resultState, configServiceForRequest)
}

func (s *Server) handleArticleRequest(resp http.ResponseWriter, resultState *ResultState, configService *config.ConfigService) {
	outputFilename := resultState.Title

	resultFilepath, err := s.initFilePaths(resp, resultState, outputFilename)
	if err != nil {
		// Logging and setting error states already happened in initHandleRequest
		return
	}

	go func() {
		ebookGeneratorService := generator.NewEbookGenerator(configService, s.fileCache)
		ebookGeneratorService.GenerateArticleEbook(resultState.Title, resultFilepath)
		resultState.Status = ResultStatusSuccess
		resultState.resultPath = resultFilepath
	}()

	s.returnState(resp, resultState)
}

func (s *Server) handleProjectPostRequest(resp http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(resp, req.Body, s.configService.Get().ServerMaxRequestBodySize)

	sigolo.Debugf("Received request %s %s for project", req.Method, req.URL)

	// Set dummy-title and later fill the title in the result state
	resultState := s.createNewResultState("project")

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			sigolo.Errorf("%+v", errors.Wrapf(err, "Request too large, only %d bytes allowed", s.configService.Get().ServerMaxRequestBodySize))
			s.returnInternalServerError(resp, resultState, fmt.Sprintf("Request too large, only %d bytes allowed", s.configService.Get().ServerMaxRequestBodySize))
			return
		}

		sigolo.Errorf("%+v", errors.Wrapf(err, "Error reading request body"))
		s.returnInternalServerError(resp, resultState, "Error reading request body")
		return
	}

	// Read body to current config. Fields not set by the given request-config stay unchanged, so only the fields that
	// are present in the request-project will be set here.
	project, err := config.LoadProjectFromBytes(bodyBytes)
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrapf(err, "Error turning request body into project instance"))
		s.returnInternalServerError(resp, resultState, "Error turning request body into project instance")
		return
	}

	cgf := config.NewDefaultConfig()
	cgf.MergeNonDefaultValues(&project.Configuration)
	project.Configuration = *cgf

	resultState.Title = project.Metadata.Title

	// Restore certain config entries that should not be set by users of the API:
	s.resetNonUploadableProjectProperties(project)

	configServiceForRequest := config.NewConfigServiceForConfig(&project.Configuration)

	resultFilepath, err := s.initFilePaths(resp, resultState, resultState.ResultToken)
	if err != nil {
		// Logging and setting error states already happened in initHandleRequest
		return
	}

	project.OutputFile = resultFilepath

	go func() {
		ebookGeneratorService := generator.NewEbookGenerator(configServiceForRequest, s.fileCache)
		ebookGeneratorService.GenerateBookFromProject(project)
		resultState.Status = ResultStatusSuccess
		resultState.resultPath = project.OutputFile
	}()

	s.returnState(resp, resultState)
}

func (s *Server) handleStandalonePostRequest(resp http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(resp, req.Body, s.configService.Get().ServerMaxRequestBodySize)

	sigolo.Debugf("Received request %s %s for standalone eBook", req.Method, req.URL)

	// Set dummy-title and later fill the title in the result state
	resultState := s.createNewResultState("standalone")

	var configBytes []byte
	var contentBytes []byte
	var err error

	if strings.HasPrefix(req.Header.Get("content-type"), "multipart/form-data") {
		err = req.ParseMultipartForm(10 << 20) // Allow 10MB of the body to stay in memory
		if err != nil {
			var maxBytesError *http.MaxBytesError
			if errors.As(err, &maxBytesError) {
				sigolo.Errorf("%+v", errors.Wrapf(err, "Request too large, only %d bytes allowed", s.configService.Get().ServerMaxRequestBodySize))
				s.returnInternalServerError(resp, resultState, fmt.Sprintf("Request too large, only %d bytes allowed", s.configService.Get().ServerMaxRequestBodySize))
				return
			}

			sigolo.Errorf("%+v", errors.Wrapf(err, "Error parsing form-data"))
			s.returnInternalServerError(resp, resultState, "Error parsing form-data")
			return
		}
	}

	if req.MultipartForm == nil {
		sigolo.Debug("Process non-multi-part request and treat body as content")
		contentBytes, err = io.ReadAll(req.Body)
		if err != nil {
			var maxBytesError *http.MaxBytesError
			if errors.As(err, &maxBytesError) {
				sigolo.Errorf("%+v", errors.Wrapf(err, "Request too large, only %d bytes allowed", s.configService.Get().ServerMaxRequestBodySize))
				s.returnInternalServerError(resp, resultState, fmt.Sprintf("Request too large, only %d bytes allowed", s.configService.Get().ServerMaxRequestBodySize))
				return
			}

			sigolo.Errorf("%+v", errors.Wrapf(err, "Error reading request body"))
			s.returnInternalServerError(resp, resultState, "Error reading request body")
			return
		}
	} else {
		sigolo.Debug("Process multi-part request")
		configBytes = []byte(req.FormValue("config"))
		contentBytes = []byte(req.FormValue("content"))
	}

	currentConfig := config.NewDefaultConfig()
	currentConfig.MergeNonDefaultValues(s.configService.Get())

	// Read body to current config. Fields not set by the given request-config stay unchanged, so only the fields that
	// are present in the request-config will be set here.
	err = json.Unmarshal(configBytes, currentConfig)
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrapf(err, "Error reading request body"))
		s.returnInternalServerError(resp, resultState, "Error reading request body")
		return
	}

	// Restore certain config entries that should not be set by users of the API:
	s.resetNonUploadableConfigProperties(currentConfig)

	configServiceForRequest := config.NewConfigServiceForConfig(currentConfig)

	outputFilename := resultState.ResultToken

	resultFilepath, err := s.initFilePaths(resp, resultState, outputFilename)
	if err != nil {
		// Logging and setting error states already happened in initHandleRequest
		return
	}

	go func() {
		ebookGeneratorService := generator.NewEbookGenerator(configServiceForRequest, s.fileCache)
		ebookGeneratorService.GenerateStandaloneEbookFromString(contentBytes, resultFilepath, resultState.Title)
		resultState.Status = ResultStatusSuccess
		resultState.resultPath = resultFilepath
	}()

	s.returnState(resp, resultState)
}

// initFilePaths initializes the processing of the input, i.e. creating the output file. In case an error is
// returned, the response is already an internal server error and the result status is already set.
func (s *Server) initFilePaths(resp http.ResponseWriter, resultState *ResultState, outputFilename string) (string, error) {
	// Ensure output folder exists
	outputFolderPath := s.fileCache.GetDirPathInCache(cache.TempDirName)
	sigolo.Tracef("Ensure temp directory in cache folder '%s' exists", outputFolderPath)
	err := util.CurrentFilesystem.MkdirAll(outputFolderPath)
	if err != nil && !os.IsExist(err) {
		sigolo.Errorf("%+v", errors.Wrapf(err, "Error folder for temporary files"))
		s.returnInternalServerError(resp, resultState, "Error creating folder for temporary files")
		return "", err
	}

	// Create the output file
	sanitizedFilename := util.SanitizeFilename(outputFilename)
	tempFile, err := util.CurrentFilesystem.CreateTemp(s.fileCache.GetTempPath(), sanitizedFilename)
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrapf(err, "Error creating temporary file for article '%s'", outputFilename))
		s.returnInternalServerError(resp, resultState, fmt.Sprintf("Error creating temporary file for article '%s'", outputFilename))
		return "", err
	}
	defer tempFile.Close()
	tempFilepath := tempFile.Name()
	defer util.CurrentFilesystem.Remove(tempFilepath)
	sigolo.Tracef("Create temp file '%s'", tempFilepath)

	return tempFilepath, nil
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

	s.returnFile(resp, resultState)
}

// resetNonUploadableProjectProperties resets properties of the project to the configured values of the application. Not
// all properties are allowed to be set by users and this function takes care of that.
func (s *Server) resetNonUploadableProjectProperties(project *config.Project) {
	s.resetNonUploadableConfigProperties(&project.Configuration)

	project.OutputFile = ""
}

// resetNonUploadableConfigProperties resets properties of "config" to the configured values of the application. Not
// all properties are allowed to be set by users and this function takes care of that.
func (s *Server) resetNonUploadableConfigProperties(config *config.Configuration) {
	config.ForceRegenerateHtml = s.configService.Get().ForceRegenerateHtml
	// Should not be set by user: currentConfig.SvgSizeToViewbox
	// Should not be set by user: currentConfig.OutputType
	config.OutputDriver = s.configService.Get().OutputDriver
	config.CacheDir = s.configService.Get().CacheDir
	config.CacheMaxSize = s.configService.Get().CacheMaxSize
	config.CacheMaxAge = s.configService.Get().CacheMaxAge
	config.CacheEvictionStrategy = s.configService.Get().CacheEvictionStrategy
	config.StyleFile = s.configService.Get().StyleFile
	config.CoverImage = s.configService.Get().CoverImage
	config.CommandTemplateSvgToPng = s.configService.Get().CommandTemplateSvgToPng
	config.CommandTemplateMathSvgToPng = s.configService.Get().CommandTemplateMathSvgToPng
	config.CommandTemplateImageProcessing = s.configService.Get().CommandTemplateImageProcessing
	config.CommandTemplatePdfToPng = s.configService.Get().CommandTemplatePdfToPng
	config.CommandTemplateWebpToPng = s.configService.Get().CommandTemplateWebpToPng
	config.PandocExecutable = s.configService.Get().PandocExecutable
	config.PandocDataDir = s.configService.Get().PandocDataDir
	config.FontFiles = s.configService.Get().FontFiles
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
	config.MathConverter = s.configService.Get().MathConverter
	// Should not be set by user: currentConfig.TocDepth
	config.WorkerThreads = s.configService.Get().WorkerThreads
	config.UserAgentTemplate = s.configService.Get().UserAgentTemplate
	config.ServerPort = s.configService.Get().ServerPort
}

func (s *Server) createNewResultState(title string) *ResultState {
	resultToken := util.Hash(fmt.Sprintf("%s%d", title, time.Now().UnixNano()))
	resultState := &ResultState{
		Status:      ResultStatusInProgress,
		Title:       title,
		ResultToken: resultToken,
		resultPath:  "",
	}
	resultStates[resultToken] = resultState
	return resultState
}

// returnFile writes the result file from the resulState to the given response.
func (s *Server) returnFile(resp http.ResponseWriter, resultState *ResultState) {
	fileContent, err := util.CurrentFilesystem.ReadFile(resultState.resultPath)
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrapf(err, "Error reading file '%s' for '%s'", resultState.resultPath, resultState.Title))
		s.returnInternalServerError(resp, resultState, fmt.Sprintf("An error occurred while creating the response for '%s'", resultState.Title))
		return
	}

	resp.Header().Set("Content-Type", "application/octet-stream")
	resp.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=\"%s\"", filepath.Base(resultState.resultPath)))
	resp.WriteHeader(http.StatusOK)

	_, err = resp.Write(fileContent)
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrapf(err, "Could not write response for file '%s': %+v", resultState.resultPath, err))
		return
	}
}

func (s *Server) returnState(resp http.ResponseWriter, resultState *ResultState) {
	content, err := json.Marshal(resultState)
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrapf(err, "Error marshalling state to JSON: %#v", resultState))
		s.returnInternalServerError(resp, resultState, fmt.Sprintf("An error occurred while creating the status response for '%s'", resultState.Title))
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)

	_, err = resp.Write(content)
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrapf(err, "Could not write response for result state with token '%s': %+v", resultState.ResultToken, err))
		return
	}
}

func (s *Server) returnInternalServerError(resp http.ResponseWriter, resultState *ResultState, errorMessage string) {
	resultState.Status = ResultStatusFailed
	resp.Header().Set("Content-Type", "text/plain")
	resp.WriteHeader(http.StatusInternalServerError)
	_, err := resp.Write([]byte(fmt.Sprintf("Internal server error: %s", errorMessage)))
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrap(err, "Could not write internal server error response"))
		return
	}
}

func (s *Server) returnRequestTooLargeError(resp http.ResponseWriter, resultState *ResultState, errorMessage string) {
	resultState.Status = ResultStatusFailed
	resp.Header().Set("Content-Type", "text/plain")
	resp.WriteHeader(http.StatusRequestEntityTooLarge)
	_, err := resp.Write([]byte(fmt.Sprintf("Request entity too large: %s", errorMessage)))
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrap(err, "Could not write request too larger error response"))
		return
	}
}

func (s *Server) returnNotFound(resp http.ResponseWriter, errorMessage string) {
	resp.Header().Set("Content-Type", "text/plain")
	resp.WriteHeader(http.StatusNotFound)
	_, err := resp.Write([]byte(fmt.Sprintf("Not found: %s", errorMessage)))
	if err != nil {
		sigolo.Errorf("%+v", errors.Wrap(err, "Could not write not found response"))
		return
	}
}
