Parsing the wikitext is done using different strategies.

# Overall strategy

The general strategy is to parse each aspect of the wikitext separately.
Meaning, there's not one giant loop trying to find all kinds of different token but there are multiple loops (s. `parser.tokenizeContent()`), e.g. one for links, one for images and one for tables.
This slightly decreases performance (even though the number of string manipulations probably is the same) but simplifies the code.

Each parsing function has a different strategy how to find tokens and may replace parts of the input string with high-level tokens.
The actual strategy for a parsing function is chosen based on the complexity of the problem and performance impact.

In the following, the specific parsing strategies are described in more detail.

# Regex usage

The `pattern.go` contains several expressions used to find or replace things in the input data.

**Note:** The usage of regular expressions is **very slow** and should be avoided where possible.
In fact, more and more parsing steps move away from regular expressions but sometimes they are just very convenient. 

# Splitting by string

This strategy splits the entire input text by a certain string (e.g. `<!--` for comments) and is used to get the content between a start and end token.
Each resulting substring then starts/ends with the requested content.

One example for suffix-splitting is the cleanup of comments.
The entire content is split by `-->` which means each resulting substring ends with a comment and (very likely) contains only one comment-start-token `<!--`.
Splitting by that start token `<!--` results, for each substring of the first splitting, in two new substrings: The content before the comment and the content within the comment.

After processing the separate parts (e.g. removing them, creating high-level token, etc.), the content needs to be re-assembled by simple joining.

# Sliding cursor/window

This is used to find certain token in the text and is probably known from real compilers.

The idea is simple: 
For a string `s`, index `i` and window size `w`, the string `s[i:i+w]` is the window (also called cursor).
It slides, via a for-loop, over the content to easily find a wanted string or start token of size `w`.

Often the end token differs from the start token but is wanted as well (e.g. when finding and evaluating templates `{{some-template}}`), which can also be easily done using this strategy.
