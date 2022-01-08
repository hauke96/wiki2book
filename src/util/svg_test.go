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

	test.AssertEqual(t, svg.Width, "16.156ex")
	test.AssertEqual(t, svg.Height, "9.176ex")
	test.AssertEqual(t, svg.Style, "vertical-align: -4.005ex;")
}
