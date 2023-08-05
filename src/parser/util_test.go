package parser

import (
	"testing"
	"wiki2book/test"
)

func TestFindCorrespondingCloseToken(t *testing.T) {
	var index int

	index = findCorrespondingCloseToken("abc[def]ghi", 0, "[", "]")
	test.AssertEqual(t, -1, index)

	index = findCorrespondingCloseToken("abc[def]ghi", 1, "[", "]")
	test.AssertEqual(t, -1, index)

	index = findCorrespondingCloseToken("abc[def]ghi", 2, "[", "]")
	test.AssertEqual(t, -1, index)

	index = findCorrespondingCloseToken("abc[def]ghi", 3, "[", "]")
	test.AssertEqual(t, -1, index)

	index = findCorrespondingCloseToken("abc[def]ghi", 4, "[", "]")
	test.AssertEqual(t, 7, index)

	index = findCorrespondingCloseToken("abc[def]ghi", 5, "[", "]")
	test.AssertEqual(t, 7, index)

	index = findCorrespondingCloseToken("abc[def]ghi", 6, "[", "]")
	test.AssertEqual(t, 7, index)

	index = findCorrespondingCloseToken("abc[def]ghi", 7, "[", "]")
	test.AssertEqual(t, 7, index)

	index = findCorrespondingCloseToken("abc[def]ghi", 8, "[", "]")
	test.AssertEqual(t, -1, index)

	index = findCorrespondingCloseToken("abc[def]ghi", 9, "[", "]")
	test.AssertEqual(t, -1, index)

	index = findCorrespondingCloseToken("abc[def]ghi", 10, "[", "]")
	test.AssertEqual(t, -1, index)
}
