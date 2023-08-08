package util

import (
	"encoding/xml"
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/pkg/errors"
	"os"
	"strings"
)

type SimpleSvgAttributes struct {
	Width  string `xml:"width,attr"`
	Height string `xml:"height,attr"`
	Style  string `xml:"style,attr"`
}

func ReadSimpleAvgAttributes(filename string) (*SimpleSvgAttributes, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading SVG file "+filename)
	}

	attributes, err := parseSimpleSvgAttributes(file, filename)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing SVG file "+filename)
	}

	return attributes, nil
}

// MakeSvgSizeAbsolute turns relative width and height attributes of the given SVG file into absolute values based on
// the "viewBox" attribute. Only if both attributes (width and height) are already absolute values, nothing will be
// changed.
func MakeSvgSizeAbsolute(filename string) error {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "Error reading SVG file "+filename)
	}

	attributes, err := parseSimpleSvgAttributes(fileBytes, filename)
	if err != nil {
		return err
	}

	if !strings.HasSuffix(attributes.Width, "%") && strings.HasSuffix(attributes.Height, "%") {
		// Width and height are already absolute values, nothing to do here.
		sigolo.Debug("SVG file %s does not relative width and height attributes. Found width=%s and height=%s. File stays unchanged.", filename, attributes.Width, attributes.Height)
		return nil
	}

	updatedSvgContent, err := replaceRelativeSizeByViewboxSize(string(fileBytes), filename, attributes)
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, []byte(updatedSvgContent), 0644)
	if err != nil {
		return errors.Wrap(err, "Error writing SVG file "+filename)
	}

	return nil
}

func replaceRelativeSizeByViewboxSize(fileString string, filename string, oldAttributes *SimpleSvgAttributes) (string, error) {
	// Find index of "viewbox" attribute
	viewboxIndex := strings.Index(fileString, "viewBox=\"")
	if viewboxIndex == -1 {
		// No "viewbox" attribute specified, so we can't change the width/height.
		sigolo.Debug("SVG file %s does not contain a 'viewBox' attribute. File stays unchanged.", filename)
		return fileString, nil
	}

	viewboxAttributeContentSlice := strings.SplitN(fileString[viewboxIndex:], "\"", 3)
	if len(viewboxAttributeContentSlice) == 1 {
		// SVG file probably broken, at least we're not able to find a correct value for the viewbox attribute.
		sigolo.Debug("Unable to find 'viewBox' attribute values in file %s. File stays unchanged.", filename)
		return fileString, nil
	}

	var viewboxAttributeValues []string
	viewboxAttributeString := viewboxAttributeContentSlice[1]
	if strings.Contains(viewboxAttributeString, ",") {
		viewboxAttributeValues = strings.Split(viewboxAttributeString, ",")
	} else if strings.Contains(viewboxAttributeString, " ") {
		viewboxAttributeValues = strings.Split(viewboxAttributeString, " ")
	} else {
		// No supportes separator found
		sigolo.Debug("Unsupported separator for 'viewBox' attribute values in file %s, file stays unchanged. Expected comma or space in attribute 'viewbox=\"%s\"'", filename, viewboxAttributeString)
		return fileString, nil
	}

	if len(viewboxAttributeValues) != 4 {
		// Wrong number of elements in viewbox
		sigolo.Debug("Wrong number of 'viewBox' attribute values in file %s: Expected 4 but got %d. File stays unchanged.", filename, len(viewboxAttributeValues))
		return fileString, nil
	}

	fileString = strings.Replace(fileString, "width=\""+oldAttributes.Width+"\"", "width=\""+viewboxAttributeValues[2]+"pt\"", 1)
	fileString = strings.Replace(fileString, "height=\""+oldAttributes.Height+"\"", "height=\""+viewboxAttributeValues[3]+"pt\"", 1)
	return fileString, nil
}

func parseSimpleSvgAttributes(file []byte, filename string) (*SimpleSvgAttributes, error) {
	var svg = &SimpleSvgAttributes{}
	err := xml.Unmarshal(file, &svg)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Unable to unmarshal XML of SVG document %s", filename))
	}

	return svg, nil
}
