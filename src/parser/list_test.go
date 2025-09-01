package parser

import (
	"fmt"
	"strings"
	"testing"
	"wiki2book/test"
	"wiki2book/wikipedia"
)

func TestParseList(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})
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
`, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 0), fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_ORDERED_LIST, 1))
	test.AssertEqual(t, expectedTokenizedContent, newContent)
	item1 := ListItemToken{Type: NORMAL_ITEM, Content: " a"}
	item2 := ListItemToken{Type: NORMAL_ITEM, Content: " b"}
	test.AssertMapEqual(t, map[string]Token{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 0): UnorderedListToken{Items: []ListItemToken{item1}},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_ORDERED_LIST, 1):   OrderedListToken{Items: []ListItemToken{item2}},
	}, tokenizer.getTokenMap())
}
func TestParseList_withIndentation(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})
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
`, fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 0), fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_ORDERED_LIST, 1))
	test.AssertEqual(t, expectedTokenizedContent, newContent)
	item1 := ListItemToken{Type: NORMAL_ITEM, Content: " a"}
	item2 := ListItemToken{Type: NORMAL_ITEM, Content: " b"}
	test.AssertMapEqual(t, map[string]Token{
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 0): UnorderedListToken{Items: []ListItemToken{item1}},
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_ORDERED_LIST, 1):   OrderedListToken{Items: []ListItemToken{item2}},
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
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})
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
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, "*")
	test.AssertEqual(t, 3, i)
	item1 := ListItemToken{Type: NORMAL_ITEM, Content: " foo"}
	item2 := ListItemToken{Type: NORMAL_ITEM, Content: " bar"}
	list := UnorderedListToken{Items: []ListItemToken{item1, item2}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 0)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, list, token)
	test.AssertMapEqual(t, map[string]Token{
		expectedTokenKey: list,
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
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, "*")
	test.AssertEqual(t, 3, i)
	item1 := ListItemToken{Type: NORMAL_ITEM, Content: " a"}
	item2 := ListItemToken{Type: NORMAL_ITEM, Content: " A"}
	list := UnorderedListToken{Items: []ListItemToken{item1, item2}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 0)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, list, token)
	test.AssertMapEqual(t, map[string]Token{
		expectedTokenKey: list,
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
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, "*")
	test.AssertEqual(t, 6, i)
	itemInner1 := ListItemToken{Type: NORMAL_ITEM, Content: fmt.Sprintf(" b%sa%sr", MARKER_ITALIC_OPEN, MARKER_ITALIC_CLOSE)}
	itemInner2 := ListItemToken{Type: NORMAL_ITEM, Content: fmt.Sprintf(" f%so%so", MARKER_ITALIC_OPEN, MARKER_ITALIC_CLOSE)}
	innerList := UnorderedListToken{Items: []ListItemToken{itemInner1, itemInner2}}
	itemOuter1 := ListItemToken{Type: NORMAL_ITEM, Content: " foo"}
	itemOuter2 := ListItemToken{Type: NORMAL_ITEM, Content: " bar1", SubLists: []ListToken{innerList}}
	itemOuter3 := ListItemToken{Type: NORMAL_ITEM, Content: " bar2"}
	list := UnorderedListToken{Items: []ListItemToken{itemOuter1, itemOuter2, itemOuter3}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 1)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, list, token)
	test.AssertMapEqual(t, map[string]Token{
		expectedTokenKey: list,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 0): innerList,
	}, tokenizer.getTokenMap())
}

func TestTokenizeList_higherLevelStart(t *testing.T) {
	content := `stuff before list
*** foo
*** bar
end of list
`
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, "*")
	test.AssertEqual(t, 3, i)
	item1 := ListItemToken{Type: NORMAL_ITEM, Content: " foo"}
	item2 := ListItemToken{Type: NORMAL_ITEM, Content: " bar"}
	listInner := UnorderedListToken{Items: []ListItemToken{item1, item2}}
	itemMiddle := ListItemToken{Type: NORMAL_ITEM, Content: "", SubLists: []ListToken{listInner}}
	listMiddle := UnorderedListToken{Items: []ListItemToken{itemMiddle}}
	itemOuter := ListItemToken{Type: NORMAL_ITEM, Content: "", SubLists: []ListToken{listMiddle}}
	listOuter := UnorderedListToken{Items: []ListItemToken{itemOuter}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 2)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, listOuter, token)
	test.AssertMapEqual(t, map[string]Token{
		expectedTokenKey: listOuter,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 1): listMiddle,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 0): listInner,
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList(t *testing.T) {
	content := `bla
; foo : bar
; blubb
end`

	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ";")

	test.AssertEqual(t, 3, i)
	item1 := ListItemToken{Type: DESCRIPTION_HEAD, Content: " foo "}
	item2 := ListItemToken{Type: DESCRIPTION_ITEM, Content: " bar"}
	item3 := ListItemToken{Type: DESCRIPTION_HEAD, Content: " blubb"}
	list := DescriptionListToken{Items: []ListItemToken{item1, item2, item3}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 0)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, list, token)
	test.AssertMapEqual(t, map[string]Token{
		expectedTokenKey: list,
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList_headingIsSecondItem(t *testing.T) {
	content := `bla
: foo
; bar
blubb`

	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ";")

	test.AssertEqual(t, 3, i)
	item1 := ListItemToken{Type: DESCRIPTION_ITEM, Content: " foo"}
	item2 := ListItemToken{Type: DESCRIPTION_HEAD, Content: " bar"}
	list := DescriptionListToken{Items: []ListItemToken{item1, item2}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 0)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, list, token)
	test.AssertMapEqual(t, map[string]Token{
		expectedTokenKey: list,
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList_withoutHeading(t *testing.T) {
	content := `bla
: foo
: bar
blubb`

	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ";")

	test.AssertEqual(t, 3, i)
	item1 := ListItemToken{Type: DESCRIPTION_ITEM, Content: " foo"}
	item2 := ListItemToken{Type: DESCRIPTION_ITEM, Content: " bar"}
	list := DescriptionListToken{Items: []ListItemToken{item1, item2}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 0)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, list, token)
	test.AssertMapEqual(t, map[string]Token{
		expectedTokenKey: list,
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList_deepBeginning(t *testing.T) {
	content := `bla
::: foo
::: bar
blubb`

	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ":")

	test.AssertEqual(t, 3, i)
	item1 := ListItemToken{Type: DESCRIPTION_ITEM, Content: " foo"}
	item2 := ListItemToken{Type: DESCRIPTION_ITEM, Content: " bar"}
	listInner := DescriptionListToken{Items: []ListItemToken{item1, item2}}
	itemMiddle := ListItemToken{Type: DESCRIPTION_ITEM, Content: "", SubLists: []ListToken{listInner}}
	listMiddle := DescriptionListToken{Items: []ListItemToken{itemMiddle}}
	itemOuter := ListItemToken{Type: DESCRIPTION_ITEM, Content: "", SubLists: []ListToken{listMiddle}}
	listOuter := DescriptionListToken{Items: []ListItemToken{itemOuter}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 2)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, listOuter, token)
	test.AssertMapEqual(t, map[string]Token{
		expectedTokenKey: listOuter,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 1): listMiddle,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 0): listInner,
	}, tokenizer.getTokenMap())
}

func TestTokenizeDescriptionList_withEmptyItem(t *testing.T) {
	content := `foo
; list
:
: bar
blubb`

	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ";")

	test.AssertEqual(t, 4, i)
	head1 := ListItemToken{Type: DESCRIPTION_HEAD, Content: " list"}
	item2 := ListItemToken{Type: DESCRIPTION_ITEM, Content: ""}
	item3 := ListItemToken{Type: DESCRIPTION_ITEM, Content: " bar"}
	list := DescriptionListToken{Items: []ListItemToken{head1, item2, item3}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 0)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, list, token)
	test.AssertMapEqual(t, map[string]Token{
		expectedTokenKey: list,
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

	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, ";")

	test.AssertEqual(t, 6, i)
	innerItem1 := ListItemToken{Type: NORMAL_ITEM, Content: " bar1"}
	innerItem2 := ListItemToken{Type: NORMAL_ITEM, Content: " bar2"}
	innerList := UnorderedListToken{Items: []ListItemToken{innerItem1, innerItem2}}
	head := ListItemToken{Type: DESCRIPTION_HEAD, Content: " list"}
	item1 := ListItemToken{Type: DESCRIPTION_ITEM, Content: " foo1", SubLists: []ListToken{innerList}}
	item2 := ListItemToken{Type: DESCRIPTION_ITEM, Content: " foo2"}
	listOuter := DescriptionListToken{Items: []ListItemToken{head, item1, item2}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 1)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, listOuter, token)
	test.AssertMapEqual(t, map[string]Token{
		expectedTokenKey: listOuter,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 0): innerList,
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
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})
	token, tokenKey, i := tokenizer.tokenizeList(strings.Split(content, "\n"), 1, "*")
	test.AssertEqual(t, 5, i)
	innerHead := ListItemToken{Type: DESCRIPTION_HEAD, Content: " descr"}
	innerItem := ListItemToken{Type: DESCRIPTION_ITEM, Content: " descr-item"}
	innerList := DescriptionListToken{Items: []ListItemToken{innerHead, innerItem}}
	itemOuter1 := ListItemToken{Type: NORMAL_ITEM, Content: " foo", SubLists: []ListToken{innerList}}
	itemOuter2 := ListItemToken{Type: NORMAL_ITEM, Content: " bar"}
	listOuter := UnorderedListToken{Items: []ListItemToken{itemOuter1, itemOuter2}}
	expectedTokenKey := fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_UNORDERED_LIST, 1)
	test.AssertEqual(t, expectedTokenKey, tokenKey)
	test.AssertEqual(t, listOuter, token)
	test.AssertMapEqual(t, map[string]Token{
		expectedTokenKey: listOuter,
		fmt.Sprintf(TOKEN_TEMPLATE, TOKEN_DESCRIPTION_LIST, 0): innerList,
	}, tokenizer.getTokenMap())
}

func TestGetListTokenKey(t *testing.T) {
	tokenizer := NewTokenizer(&wikipedia.DummyWikipediaService{})
	test.AssertEqual(t, TOKEN_UNORDERED_LIST, tokenizer.getListTokenKey("*"))
	test.AssertEqual(t, TOKEN_ORDERED_LIST, tokenizer.getListTokenKey("#"))
	test.AssertEqual(t, TOKEN_DESCRIPTION_LIST, tokenizer.getListTokenKey(";"))
	test.AssertEqual(t, "UNKNOWN_LIST_TYPE_-", tokenizer.getListTokenKey("-"))
	test.AssertEqual(t, "UNKNOWN_LIST_TYPE_~", tokenizer.getListTokenKey("~"))
	test.AssertEqual(t, "UNKNOWN_LIST_TYPE_ ", tokenizer.getListTokenKey(" "))
	test.AssertEqual(t, "UNKNOWN_LIST_TYPE_", tokenizer.getListTokenKey(""))
}
