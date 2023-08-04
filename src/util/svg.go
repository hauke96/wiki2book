package util

import (
	"encoding/xml"
	"github.com/pkg/errors"
	"os"
)

type SimpleSvgAttributes struct {
	Width  string `xml:"width,attr"`
	Height string `xml:"height,attr"`
	Style  string `xml:"style,attr"`
}

func ReadSvg(filename string) (*SimpleSvgAttributes, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading svg file "+filename)
	}

	var svg = &SimpleSvgAttributes{}
	err = xml.Unmarshal(file, &svg)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing svg file "+filename)
	}

	return svg, nil
}
