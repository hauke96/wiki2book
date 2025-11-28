package image

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
	"golang.org/x/net/html/charset"
)

type SimpleSvgAttributes struct {
	Width  string `xml:"width,attr"`
	Height string `xml:"height,attr"`
	Style  string `xml:"style,attr"`
}

var (
	xmlNamespaceValueRegex = regexp.MustCompile(`&ns.*?;`) // Attributes like  xmlns="&ns_svg;"  cause error during Unmarshalling.
)

func ReadSimpleAvgAttributes(filename string) (*SimpleSvgAttributes, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading SVG file '%s'", filename)
	}

	attributes, err := parseSimpleSvgAttributes(file, filename)
	if err != nil {
		return nil, errors.Wrapf(err, "Error parsing SVG file '%s'", filename)
	}
	sigolo.Tracef("Read simple SVG attributes %#v from file '%s'", attributes, filename)

	return attributes, nil
}

// MakeSvgSizeAbsolute turns relative width and height attributes of the given SVG file into absolute values based on
// the "viewBox" attribute. Only if both attributes (width and height) are already absolute values, nothing will be
// changed.
func MakeSvgSizeAbsolute(filename string) error {
	sigolo.Debugf("Make SVG size absolute for image file '%s'", filename)

	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return errors.Wrapf(err, "Error reading SVG file '%s'", filename)
	}

	attributes, err := parseSimpleSvgAttributes(fileBytes, filename)
	if err != nil {
		return err
	}
	sigolo.Tracef("Found SVG attributes: %#v", attributes)

	if !strings.HasSuffix(attributes.Width, "%") && strings.HasSuffix(attributes.Height, "%") {
		// Width and height are already absolute values, nothing to do here.
		sigolo.Debugf("SVG file '%s' does not relative width and height attributes. Found width=%s and height=%s. File stays unchanged.", filename, attributes.Width, attributes.Height)
		return nil
	}

	updatedSvgContent := replaceRelativeSizeByViewboxSize(string(fileBytes), filename, attributes)
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, []byte(updatedSvgContent), 0644)
	if err != nil {
		return errors.Wrap(err, "Error writing SVG file "+filename)
	}

	return nil
}

func replaceRelativeSizeByViewboxSize(fileString string, filename string, oldAttributes *SimpleSvgAttributes) string {
	// Find index of "viewbox" attribute
	viewboxIndex := strings.Index(fileString, "viewBox=\"")
	if viewboxIndex == -1 {
		// No "viewbox" attribute specified, so we can't change the width/height.
		sigolo.Debugf("SVG file '%s' does not contain a 'viewBox' attribute. File stays unchanged.", filename)
		return fileString
	}

	viewboxAttributeContentSlice := strings.SplitN(fileString[viewboxIndex:], "\"", 3)
	if len(viewboxAttributeContentSlice) == 1 {
		// SVG file probably broken, at least we're not able to find a correct value for the viewbox attribute.
		sigolo.Debugf("Unable to find 'viewBox' attribute values in file '%s'. File stays unchanged.", filename)
		return fileString
	}

	var viewboxAttributeValues []string
	viewboxAttributeString := viewboxAttributeContentSlice[1]
	sigolo.Tracef("Found viewBox=%s", viewboxAttributeString)
	if strings.Contains(viewboxAttributeString, ",") {
		viewboxAttributeValues = strings.Split(viewboxAttributeString, ",")
	} else if strings.Contains(viewboxAttributeString, " ") {
		viewboxAttributeValues = strings.Split(viewboxAttributeString, " ")
	} else {
		// No supported separator found
		sigolo.Debugf("Unsupported separator for 'viewBox' attribute values in file '%s', file stays unchanged. Expected comma or space in attribute 'viewbox=\"%s\"'", filename, viewboxAttributeString)
		return fileString
	}

	if len(viewboxAttributeValues) != 4 {
		// Wrong number of elements in viewbox
		sigolo.Debugf("Wrong number of 'viewBox' attribute values in file '%s': Expected 4 but got %d. File stays unchanged.", filename, len(viewboxAttributeValues))
		return fileString
	}

	sigolo.Tracef("Replace width=%s -> width=%s and height=%s -> height=%s", oldAttributes.Width, viewboxAttributeValues[2], oldAttributes.Height, viewboxAttributeValues[3])
	fileString = strings.Replace(fileString, "width=\""+oldAttributes.Width+"\"", "width=\""+viewboxAttributeValues[2]+"pt\"", 1)
	fileString = strings.Replace(fileString, "height=\""+oldAttributes.Height+"\"", "height=\""+viewboxAttributeValues[3]+"pt\"", 1)
	return fileString
}

func parseSimpleSvgAttributes(fileContent []byte, filename string) (*SimpleSvgAttributes, error) {
	// Remove all namespace strings that are in a format not supported by go
	fileContentString := string(fileContent)
	fileContentString = xmlNamespaceValueRegex.ReplaceAllString(fileContentString, "")
	fileContent = []byte(fileContentString)

	// Use the NewReaderLabel to support non-UTF-8 encodings
	var svg = &SimpleSvgAttributes{}
	reader := bytes.NewReader(fileContent)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	err := decoder.Decode(&svg)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Unable to unmarshal XML of SVG document '%s'", filename))
	}

	return svg, nil
}
