package util

import (
	"github.com/hauke96/wiki2book/src/test"
	"testing"
)

func TestReadSvg(t *testing.T) {
	svg, err := ReadSvg("../test/image.svg")
	if err != nil {
		t.Errorf("%+v", err)
		t.Fail()
	}

	test.AssertEqualString(t, svg.Width, "16.156ex")
	test.AssertEqualString(t, svg.Height, "9.176ex")
	test.AssertEqualString(t, svg.Style, "vertical-align: -4.005ex;")
}

func TestReadSvg_imageNotFound(t *testing.T) {
	filename := "../test/image-that-does-not-exist.svg"
	_, err := ReadSvg(filename)
	test.AssertError(t, "Error reading svg file "+filename+": open ../test/image-that-does-not-exist.svg: no such file or directory", err)
}

func TestReadSvg_brokenImage(t *testing.T) {
	filename := "../test/image-broken.svg"
	_, err := ReadSvg(filename)
	test.AssertError(t, "Error parsing svg file "+filename+": EOF", err)
}
