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
	item0 := ListItemToken{Content: " a"}
	item2 := ListItemToken{Content: " b"}
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0):      item0,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 1): UnorderedListToken{Items: []ListToken{item0}},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 2):      item2,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_ORDERED_LIST, 3):   OrderedListToken{Items: []ListToken{item2}},
	}, tokenizer.getTokenMap())
}
func TestParseList_withIndentation(t *testing.T) {
	tokenizer := NewTokenizer("foo", "bar")
	content := `   foo
  * a
   bar
 # b
     end
`

	newContent := tokenizer.parseLists(content)

	expectedTokenizedContent := fmt.Sprintf(`   foo
%s
   bar
%s
     end
`, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 1), fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_ORDERED_LIST, 3))
	test.AssertEqual(t, expectedTokenizedContent, newContent)
	item0 := ListItemToken{Content: " a"}
	item2 := ListItemToken{Content: " b"}
	test.AssertMapEqual(t, map[string]interface{}{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0):      item0,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 1): UnorderedListToken{Items: []ListToken{item0}},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 2):      item2,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_ORDERED_LIST, 3):   OrderedListToken{Items: []ListToken{item2}},
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
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, "*")
	test.AssertEqual(t, 3, i)
	item0 := ListItemToken{Content: " foo"}
	item1 := ListItemToken{Content: " bar"}
	list2 := UnorderedListToken{Items: []ListToken{item0, item1}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 2)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, list2, token)
	test.AssertMapEqual(t, map[string]interface{}{
		expectedTokenKey: list2,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0): item0,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 1): item1,
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
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, "*")
	test.AssertEqual(t, 3, i)
	item0 := ListItemToken{Content: " a"}
	item1 := ListItemToken{Content: " A"}
	list2 := UnorderedListToken{Items: []ListToken{item0, item1}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 2)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, list2, token)
	test.AssertMapEqual(t, map[string]interface{}{
		expectedTokenKey: list2,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0): item0,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 1): item1,
	}, tokenizer.getTokenMap())
}

func TestTokenizeList_withSubList(t *testing.T) {
	content := `stuff before list
* foo
* bar1
** b''a''r
** f''o''o
* bar2
end of list
`
	tokenizer := NewTokenizer("foo", "bar")
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, "*")
	test.AssertEqual(t, 6, i)
	item0 := ListItemToken{Content: " foo"}
	item1 := ListItemToken{Content: " bar1"}
	item2 := ListItemToken{Content: fmt.Sprintf(" b%sa%sr", MARKER_ITALIC_OPEN, MARKER_ITALIC_CLOSE)}
	item3 := ListItemToken{Content: fmt.Sprintf(" f%so%so", MARKER_ITALIC_OPEN, MARKER_ITALIC_CLOSE)}
	list4 := UnorderedListToken{Items: []ListToken{item2, item3}}
	item5 := ListItemToken{Content: " bar2"}
	list6 := UnorderedListToken{Items: []ListToken{item0, item1, list4, item5}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 6)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, list6, token)
	test.AssertMapEqual(t, map[string]interface{}{
		expectedTokenKey: list6,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0):      item0,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 1):      item1,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 2):      item2,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 3):      item3,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 4): list4,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 5):      item5,
	}, tokenizer.getTokenMap())
}

func TestTokenizeList_higherLevelStart(t *testing.T) {
	content := `stuff before list
*** foo
*** bar
end of list
`
	tokenizer := NewTokenizer("foo", "bar")
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, "*")
	test.AssertEqual(t, 3, i)
	item0 := ListItemToken{Content: " foo"}
	item1 := ListItemToken{Content: " bar"}
	list2 := UnorderedListToken{Items: []ListToken{item0, item1}}
	list3 := UnorderedListToken{Items: []ListToken{list2}}
	list4 := UnorderedListToken{Items: []ListToken{list3}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 4)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, list4, token)
	test.AssertMapEqual(t, map[string]interface{}{
		expectedTokenKey: list4,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 3): list3,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 2): list2,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0):      item0,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 1):      item1,
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList(t *testing.T) {
	content := `bla
; foo : bar
; blubb
end`

	tokenizer := NewTokenizer("foo", "bar")
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ";")

	test.AssertEqual(t, 3, i)
	item0 := DescriptionListHeadToken{ListItemToken{Content: " foo "}}
	item1 := DescriptionListItemToken{ListItemToken{Content: " bar"}}
	item2 := DescriptionListHeadToken{ListItemToken{Content: " blubb"}}
	list3 := DescriptionListToken{Items: []ListToken{item0, item1, item2}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 3)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, list3, token)
	test.AssertMapEqual(t, map[string]interface{}{
		expectedTokenKey: list3,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 0): item0,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 1): item1,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 2): item2,
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList_headingIsSecondItem(t *testing.T) {
	content := `bla
: foo
; bar
blubb`

	tokenizer := NewTokenizer("foo", "bar")
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ";")

	test.AssertEqual(t, 3, i)
	item0 := DescriptionListItemToken{ListItemToken{Content: " foo"}}
	item1 := DescriptionListHeadToken{ListItemToken: ListItemToken{Content: " bar"}}
	list2 := DescriptionListToken{Items: []ListToken{item0, item1}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, list2, token)
	test.AssertMapEqual(t, map[string]interface{}{
		expectedTokenKey: list2,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 0): item0,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 1): item1,
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList_withoutHeading(t *testing.T) {
	content := `bla
: foo
: bar
blubb`

	tokenizer := NewTokenizer("foo", "bar")
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ";")

	test.AssertEqual(t, 3, i)
	item0 := DescriptionListItemToken{ListItemToken{Content: " foo"}}
	item1 := DescriptionListItemToken{ListItemToken{Content: " bar"}}
	list2 := DescriptionListToken{Items: []ListToken{item0, item1}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, list2, token)
	test.AssertMapEqual(t, map[string]interface{}{
		expectedTokenKey: list2,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 0): item0,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 1): item1,
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList_deepBeginning(t *testing.T) {
	content := `bla
::: foo
::: bar
blubb`

	tokenizer := NewTokenizer("foo", "bar")
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ":")

	test.AssertEqual(t, 3, i)
	item0 := DescriptionListItemToken{ListItemToken{Content: " foo"}}
	item1 := DescriptionListItemToken{ListItemToken{Content: " bar"}}
	list2 := DescriptionListToken{Items: []ListToken{item0, item1}}
	list3 := DescriptionListToken{Items: []ListToken{list2}}
	list4 := DescriptionListToken{Items: []ListToken{list3}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 4)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, list4, token)
	test.AssertMapEqual(t, map[string]interface{}{
		expectedTokenKey: list4,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 3):      list3,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2):      list2,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 0): item0,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 1): item1,
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList_withEmptyItem(t *testing.T) {
	content := `foo
; list
:
: bar
blubb`

	tokenizer := NewTokenizer("foo", "bar")
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ";")

	test.AssertEqual(t, 4, i)
	head0 := DescriptionListHeadToken{ListItemToken{Content: " list"}}
	item1 := DescriptionListItemToken{ListItemToken{Content: ""}}
	item2 := DescriptionListItemToken{ListItemToken{Content: " bar"}}
	list4 := DescriptionListToken{Items: []ListToken{head0, item1, item2}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 3)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, list4, token)
	test.AssertMapEqual(t, map[string]interface{}{
		expectedTokenKey: list4,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 0): head0,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 1): item1,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 2): item2,
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
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ";")

	test.AssertEqual(t, 6, i)
	head0 := DescriptionListHeadToken{ListItemToken{Content: " list"}}
	item1 := DescriptionListItemToken{ListItemToken{Content: " foo1"}}
	item2 := ListItemToken{Content: " bar1"}
	item3 := ListItemToken{Content: " bar2"}
	list4 := UnorderedListToken{Items: []ListToken{item2, item3}}
	item5 := DescriptionListItemToken{ListItemToken{Content: " foo2"}}
	list6 := DescriptionListToken{Items: []ListToken{head0, item1, list4, item5}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 6)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, list6, token)
	test.AssertMapEqual(t, map[string]interface{}{
		expectedTokenKey: list6,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 0): head0,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 1): item1,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 2):             item2,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 3):             item3,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 4):        list4,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 5): item5,
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
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, "*")
	test.AssertEqual(t, 5, i)
	item0 := ListItemToken{Content: " foo"}
	head1 := DescriptionListHeadToken{ListItemToken{Content: " descr"}}
	item2 := DescriptionListItemToken{ListItemToken{Content: " descr-item"}}
	list3 := DescriptionListToken{Items: []ListToken{head1, item2}}
	item4 := ListItemToken{Content: " bar"}
	list5 := UnorderedListToken{Items: []ListToken{item0, list3, item4}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 5)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, list5, token)
	test.AssertMapEqual(t, map[string]interface{}{
		expectedTokenKey: list5,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 0):             item0,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_HEAD, 1): head1,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST_ITEM, 2): item2,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 3):      list3,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_LIST_ITEM, 4):             item4,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 5):        list5,
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
