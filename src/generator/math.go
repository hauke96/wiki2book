package generator

import (
	"strings"
	"wiki2book/cache"
	"wiki2book/config"
	"wiki2book/util"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

type MathGenerator interface {
	// RenderMathToSvg turns the given math expression into an SVG image. The filename of the file in the cache and an
	// error are returned. The path is empty in case there is an error.
	RenderMathToSvg(mathText string) (string, error)
}

type MathGeneratorImpl struct {
}

func NewMathGenerator() *MathGeneratorImpl {
	return &MathGeneratorImpl{}
}

func (m *MathGeneratorImpl) RenderMathToSvg(mathText string) (string, error) {
	filename := util.Hash(mathText) + ".svg"

	outputFile, fileIsCached, err := cache.GetFile(cache.ImageCacheDirName, filename)
	if err == nil && fileIsCached {
		sigolo.Debugf("Math SVG '%s' already cached -> use this cached file", outputFile)
		return filename, nil
	}
	if err != nil {
		return "", errors.Wrapf(err, "Unable to check whether math SVG '%s' is already cached or not", outputFile)
	}
	sigolo.Debugf("Math SVG '%s' not cached -> render math", outputFile)

	sigolo.Tracef("Convert math string '%s' to SVG '%s'", util.TruncString(mathText), outputFile)

	commandString := strings.ReplaceAll(config.Current.CommandTemplateMathToSvg, config.InputPlaceholder, mathText)
	commandString = strings.ReplaceAll(commandString, config.OutputPlaceholder, outputFile)

	err = util.ExecuteCommandWithArgs(commandString, ".")

	return filename, errors.Wrapf(err, "Converting math string '%s' to SVG '%s' failed", util.TruncString(mathText), outputFile)
}
