Parsing the wikitext is done using different strategies.

# Overall strategy

The general strategy is to parse each aspect of the wikitext separately.
Meaning, there's not one giant loop trying to find all kinds of different token but there are multiple loops (s. `parser.tokenizeContent()`), e.g. one for links, one for images and one for tables.
This slightly decreases performance (even though the number of string manipulations probably is the same) but simplifies the code.

Each parsing loop has different strategies how to find token and replace parts of the input string.
These strategies were chosen based on the complexity of the problem and performance impact.

# Regex usage

The `pattern.go` contains several expressions used to find or replace entire lines or just to find indices of certain strings.

**Note:** The usage of regular expressions is **very slow** and should be avoided where possible.
In fact, more and more parsing steps move away from regular expressions. 

# Splitting by token

This strategy splits the text by a token and is used to get the content between a start and end token.
Each resulting substring then starts or ends with the requested content.

One example for suffix-splitting is the cleanup of comments.
The entire content is split by `-->` which means each substring ends with the comment and (very likely) contains only one comment-start-token `<!--`.

# Sliding cursor/window

This is used to find certain token in the text.
The idea is simple: 
For a string `s`, index `i` and window size `w`, the string `s[i:i+w]` is the window (often also called cursor).

Often the end token is wanted as well (e.g. when finding and evaluating templates), which can be done using the same strategy.