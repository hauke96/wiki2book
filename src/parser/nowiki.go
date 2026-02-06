package parser

type NowikiToken struct {
	Token
	Content string
}

func (t *Tokenizer) parseNowiki(content string) string {
	// The following steps are performed:
	//   1. Split by the end token "</nowiki>" of nowiki areas
	//   2. For each element in that slice, split by start token "<nowiki>" of comments
	//   3. Only append the non-comment parts of the splits to the result segments

	nowikiStart := "<nowiki>"
	nowikiEnd := "</nowiki>"
	nowikiStartLen := len(nowikiStart)
	nowikiEndLen := len(nowikiEnd)

	for i := 0; i < len(content)-nowikiEndLen; i++ {
		cursor := content[i : i+nowikiStartLen]

		if cursor == nowikiStart {
			endIndex := FindCorrespondingCloseToken(content, i+nowikiStartLen, nowikiStart, nowikiEnd)

			token := NowikiToken{
				Content: content[i+nowikiStartLen : endIndex],
			}
			tokenKey := t.getToken(TOKEN_NOWIKI)
			t.setRawToken(tokenKey, token)

			content = content[0:i] + tokenKey + content[endIndex+nowikiEndLen:]
		}
	}

	return content
}
