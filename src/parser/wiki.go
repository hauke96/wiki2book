package parser

type Article struct {
	Title    string
	Content  string
	TokenMap map[string]string
	Images   []string
}
