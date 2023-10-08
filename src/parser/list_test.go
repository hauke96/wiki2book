package parser

import (
	"fmt"
	"strings"
	"testing"
	"wiki2book/test"
)

func TestParseList(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `foo
* a
bar
# b
end
`

	newContent := tokenizer.parseLists(content)

	expectedTokenizedContent := fmt.Sprintf(`foo
%s
bar
%s
end
`, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 1), fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_ORDERED_LIST, 3))
	test.AssertEqual(t, expectedTokenizedContent, newContent)
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 1): fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_ORDERED_LIST, 3):   fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 2),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0):      " a",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 2):      " b",
	}, tokenizer.getTokenMap())
}

func TestBelongsToListPrefix(t *testing.T) {
	test.AssertTrue(t, belongsToListPrefix("* foo", "*"))
	test.AssertTrue(t, belongsToListPrefix("# foo", "#"))
	test.AssertTrue(t, belongsToListPrefix(": foo", ":"))
	test.AssertTrue(t, belongsToListPrefix("; foo", ";"))
	test.AssertTrue(t, belongsToListPrefix(": foo", ";"))

	test.AssertFalse(t, belongsToListPrefix("; foo", ":"))
	test.AssertFalse(t, belongsToListPrefix("# foo", "*"))
}

func TestRemoveListPrefix(t *testing.T) {
	test.AssertEqual(t, " foo", removeListPrefix("* foo", "*"))
	test.AssertEqual(t, " foo", removeListPrefix("# foo", "#"))
	test.AssertEqual(t, " foo", removeListPrefix(": foo", ":"))
	test.AssertEqual(t, " foo", removeListPrefix("; foo", ";"))
	test.AssertEqual(t, " foo", removeListPrefix(": foo", ";"))

	test.AssertEqual(t, "* foo", removeListPrefix("* foo", "#"))
	test.AssertEqual(t, "; foo", removeListPrefix("; foo", ":"))
}

func TestParseList_withoutListInContent(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `some
text
without list
`

	newContent := tokenizer.parseLists(content)

	test.AssertEqual(t, content, newContent)
}

func TestTokenizeList(t *testing.T) {
	content := `stuff before list
* foo
* bar
end of list
`
	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, "*", TOKEN_UNORDERED_LIST)
	test.AssertEqual(t, 3, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 2), token)
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 2): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0, TOKEN_LIST_ITEM, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0):      " foo",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 1):      " bar",
	}, tokenizer.getTokenMap())
}

func TestTokenizeList_twoTypesBelowEachOther(t *testing.T) {
	content := `stuff before list
* a
* A
# b
# B
end of list
`
	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, "*", TOKEN_UNORDERED_LIST)
	test.AssertEqual(t, 3, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 2), token)
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 2): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0, TOKEN_LIST_ITEM, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0):      " a",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 1):      " A",
	}, tokenizer.getTokenMap())
}

func TestTokenizeList_withSubList(t *testing.T) {
	content := `stuff before list
* foo
** b''a''r
* bar
end of list
`
	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, "*", TOKEN_UNORDERED_LIST)
	test.AssertEqual(t, 4, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 4), token)
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 4): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 2, TOKEN_LIST_ITEM, 3),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 1): fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 2):      fmt.Sprintf(" foo "+TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0):      fmt.Sprintf(" b%sa%sr", MARKER_ITALIC_OPEN, MARKER_ITALIC_CLOSE),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 3):      " bar",
	}, tokenizer.getTokenMap())
}

func TestTokenizeList_higherLevelStart(t *testing.T) {
	content := `stuff before list
*** foo
*** bar
end of list
`
	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, "*", TOKEN_UNORDERED_LIST)
	test.AssertEqual(t, 3, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 4), token)
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 4): fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 3),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 3): fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 2),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 2): fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0, TOKEN_LIST_ITEM, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0):      " foo",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 1):      " bar",
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList(t *testing.T) {
	content := `bla
; foo: bar
: blubb
blubb`

	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ";", TOKEN_DESCRIPTION_LIST)

	test.AssertEqual(t, 3, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 3), token)
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 3):      fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 0, TOKEN_DESCRIPTION_LIST_ITEM, 1, TOKEN_DESCRIPTION_LIST_ITEM, 2),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 0): " foo",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 1): " bar",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 2): " blubb",
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList_headingIsSecondItem(t *testing.T) {
	content := `bla
: foo
; bar
blubb`

	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ";", TOKEN_DESCRIPTION_LIST)

	test.AssertEqual(t, 3, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2), token)
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2):      fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 0, TOKEN_DESCRIPTION_LIST_HEAD, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 0): " foo",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 1): " bar",
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList_withoutHeading(t *testing.T) {
	content := `bla
: foo
: bar
blubb`

	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ";", TOKEN_DESCRIPTION_LIST)

	test.AssertEqual(t, 3, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2), token)
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2):      fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 0, TOKEN_DESCRIPTION_LIST_ITEM, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 0): " foo",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 1): " bar",
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList_deepBeginning(t *testing.T) {
	content := `bla
::: foo
::: bar
blubb`

	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ":", TOKEN_DESCRIPTION_LIST)

	test.AssertEqual(t, 3, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 4), token)
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 4):      fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 3),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 3):      fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2):      fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 0, TOKEN_DESCRIPTION_LIST_ITEM, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 0): " foo",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 1): " bar",
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList_withEmptyItem(t *testing.T) {
	content := `foo
; list
:
: bar
blubb`

	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ";", TOKEN_DESCRIPTION_LIST)

	test.AssertEqual(t, 4, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 3), token)
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 3):      fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 0, TOKEN_DESCRIPTION_LIST_ITEM, 1, TOKEN_DESCRIPTION_LIST_ITEM, 2),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 0): " list",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 1): "",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 2): " bar",
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList_withOtherSubList(t *testing.T) {
	content := `foo
; list
: foo1
:* bar1
:* bar2
: foo2
blubb`

	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ";", TOKEN_DESCRIPTION_LIST)

	test.AssertEqual(t, 6, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 9), token)
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 9):      fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 0, TOKEN_DESCRIPTION_LIST_ITEM, 1, TOKEN_DESCRIPTION_LIST_ITEM, 4, TOKEN_DESCRIPTION_LIST_ITEM, 7, TOKEN_DESCRIPTION_LIST_ITEM, 8),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 0): " list",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 1): " foo1",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 4): fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 3),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 3):        fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 2),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 2):             " bar1",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 7): fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 6),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 6):        fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 5),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 5):             " bar2",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 8): " foo2",
	}, tokenizer.getTokenMap())
}

func TestTokenizeList_withDescriptionSubList(t *testing.T) {
	content := `stuff before list
* foo
*; descr
*: descr-item
* bar
end of list
`
	tokenizer := NewTokenizer("foo", "bar")
	token, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, "*", TOKEN_UNORDERED_LIST)
	test.AssertEqual(t, 5, i)
	test.AssertEqual(t, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 5), token)
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 5):        fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 3, TOKEN_LIST_ITEM, 4),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 3):             fmt.Sprintf(" foo "+TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 4):             " bar",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2):      fmt.Sprintf(TOKEN_TEMPLATE+" "+TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 0, TOKEN_DESCRIPTION_LIST_ITEM, 1),
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 0): " descr",
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 1): " descr-item",
	}, tokenizer.getTokenMap())
}

func TestGetListTokenKey(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	test.AssertEqual(t, TOKEN_UNORDERED_LIST, tokenizer.getListTokenKey("*"))
	test.AssertEqual(t, TOKEN_ORDERED_LIST, tokenizer.getListTokenKey("#"))
	test.AssertEqual(t, TOKEN_DESCRIPTION_LIST, tokenizer.getListTokenKey(";"))
	test.AssertEqual(t, "UNKNOWN_LIST_TYPE_-", tokenizer.getListTokenKey("-"))
	test.AssertEqual(t, "UNKNOWN_LIST_TYPE_~", tokenizer.getListTokenKey("~"))
	test.AssertEqual(t, "UNKNOWN_LIST_TYPE_ ", tokenizer.getListTokenKey(" "))
	test.AssertEqual(t, "UNKNOWN_LIST_TYPE_", tokenizer.getListTokenKey(""))
}

func TestGetListItemTokenKey(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	test.AssertEqual(t, TOKEN_LIST_ITEM, tokenizer.getListItemTokenKey("*"))
	test.AssertEqual(t, TOKEN_LIST_ITEM, tokenizer.getListItemTokenKey("#"))
	test.AssertEqual(t, TOKEN_DESCRIPTION_LIST_HEAD, tokenizer.getListItemTokenKey(";"))
	test.AssertEqual(t, TOKEN_DESCRIPTION_LIST_ITEM, tokenizer.getListItemTokenKey(":"))
	test.AssertEqual(t, "UNKNOWN_LIST_ITEM_TYPE_-", tokenizer.getListItemTokenKey("-"))
	test.AssertEqual(t, "UNKNOWN_LIST_ITEM_TYPE_~", tokenizer.getListItemTokenKey("~"))
	test.AssertEqual(t, "UNKNOWN_LIST_ITEM_TYPE_ ", tokenizer.getListItemTokenKey(" "))
	test.AssertEqual(t, "UNKNOWN_LIST_ITEM_TYPE_", tokenizer.getListItemTokenKey(""))
}
