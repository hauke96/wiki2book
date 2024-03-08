package generator

import (
	"github.com/hauke96/sigolo"
	"github.com/pkg/errors"
	"strings"
	"wiki2book/util"
)

const (
	OutputTypeEpub2 = "epub2"
	OutputTypeEpub3 = "epub3"

	OutputDriverPandoc   = "pandoc"
	OutputDriverInternal = "internal"
)

var (
	AllOutputTypes = []string{
		OutputTypeEpub2,
		OutputTypeEpub3,
	}

	AllOutputDrivers = []string{
		OutputDriverPandoc,
		OutputDriverInternal,
	}
)

// VerifyOutputAndDriver returns an error if the output type and driver are not compatible and returns nil if they are.
func VerifyOutputAndDriver(outputType string, outputDriver string) error {
	sigolo.Debug("Verify compatibility of outputType '%s' and outputDriver '%s'", outputType, outputDriver)

	if !util.Contains(AllOutputDrivers, outputDriver) {
		return errors.Errorf("Unknown output driver '%s'. Known driver: '%s'", outputDriver, strings.Join(AllOutputDrivers, ", "))
	}

	switch outputType {
	case OutputTypeEpub2:
		if outputDriver == OutputDriverPandoc {
			return nil
		}
		return errors.Errorf("Incompatible output type '%s' with output driver '%s'", outputType, outputDriver)
	case OutputTypeEpub3:
		if outputDriver == OutputDriverPandoc ||
			outputDriver == OutputDriverInternal {
			return nil
		}
		return errors.Errorf("Incompatible output type '%s' with output driver '%s'", outputType, outputDriver)
	}

	return errors.Errorf("Unknown output type '%s'. Known types: '%s'", outputType, strings.Join(AllOutputTypes, ", "))
}
