package util

import (
	"testing"
	"wiki2book/test"
)

func TestReadSvg(t *testing.T) {
	svg, err := ReadSimpleAvgAttributes("../test/image.svg")
	if err != nil {
		t.Errorf("%+v", err)
		t.Fail()
	}

	test.AssertEqual(t, svg.Width, "16.156ex")
	test.AssertEqual(t, svg.Height, "9.176ex")
	test.AssertEqual(t, svg.Style, "vertical-align: -4.005ex;")
}

func TestReadSvg_imageNotFound(t *testing.T) {
	filename := "../test/image-that-does-not-exist.svg"
	_, err := ReadSimpleAvgAttributes(filename)
	test.AssertError(t, "Error reading SVG file "+filename+": open ../test/image-that-does-not-exist.svg: no such file or directory", err)
}

func TestReadSvg_brokenImage(t *testing.T) {
	filename := "../test/image-broken.svg"
	_, err := ReadSimpleAvgAttributes(filename)
	test.AssertError(t, "Error parsing SVG file ../test/image-broken.svg: Unable to unmarshal XML of SVG document "+filename+": EOF", err)
}

func TestMakeSvgSizeAbsolute(t *testing.T) {
	fileBytes := []byte(`<svg width="50%" height="100%" viewBox="0 0 123 234">
  <rect width="100%" height="100%" />
</svg>`)
	attributedBefore, err := parseSimpleSvgAttributes(fileBytes, "foo.svg")
	if err != nil {
		t.Errorf("%+v", err)
		t.Fail()
	}
	test.AssertEqual(t, attributedBefore.Width, "50%")
	test.AssertEqual(t, attributedBefore.Height, "100%")

	updatedFileContent, err := replaceRelativeSizeByViewboxSize(string(fileBytes), "../test/image_relative-width-height.svg", attributedBefore)
	if err != nil {
		t.Errorf("%+v", err)
		t.Fail()
	}

	attributedAfter, err := parseSimpleSvgAttributes([]byte(updatedFileContent), "foo.svg")
	if err != nil {
		t.Errorf("%+v", err)
		t.Fail()
	}
	test.AssertEqual(t, attributedAfter.Width, "123pt")
	test.AssertEqual(t, attributedAfter.Height, "234pt")
}
