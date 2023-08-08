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

func TestFindCorrespondingCloseToken_multiLineStrings(t *testing.T) {
	var index int

	index = findCorrespondingCloseToken(`ab
c[def]ghi`, 5, "[", "]")
	test.AssertEqual(t, 8, index)

	index = findCorrespondingCloseToken(`abc[d
ef]ghi`, 5, "[", "]")
	test.AssertEqual(t, 8, index)

	index = findCorrespondingCloseToken(`abc[def]g
hi`, 5, "[", "]")
	test.AssertEqual(t, 7, index)
}

func TestFindCorrespondingCloseToken_specialChars(t *testing.T) {
	var index int

	index = findCorrespondingCloseToken("abc[äöü]ghi", 4, "[", "]")
	test.AssertEqual(t, 10, index)

	index = findCorrespondingCloseToken("abc[dµf]ghi", 4, "[", "]")
	test.AssertEqual(t, 8, index)

	index = findCorrespondingCloseToken(`abc[[:EN:FOO:BAR
$ome+µeird-string]]ghi`, 5, "[", "]")
	test.AssertEqual(t, 35, index)
}
