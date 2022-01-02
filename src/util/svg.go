package util

import (
	"encoding/xml"
	"github.com/pkg/errors"
	"io/ioutil"
)

type SVG struct {
	Width  string `xml:"width,attr"`
	Height string `xml:"height,attr"`
	Style  string `xml:"style,attr"`
}

func ReadSvg(filename string) (*SVG, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading svg file "+filename)
	}

	svg := &SVG{}
	err = xml.Unmarshal(file, svg)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing svg file "+filename)
	}

	return svg, nil
}
