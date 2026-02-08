package server

import (
	"fmt"
	"net/http"
	"wiki2book/config"

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
	sigolo.FatalCheck(errors.Wrapf(err, "Unable to start HTTP server on port %d", config.Current.ServerPort))
}

func handleArticleRequest(resp http.ResponseWriter, req *http.Request) {
	article := req.PathValue(pathVarArticleName)

	sigolo.Debugf("Received request %s %s for article %s", req.Method, req.URL, article)

}
