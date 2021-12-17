package wiki

type Article struct {
	Title   string
	Content string
	Images  []Image
}

type Image struct {
	Filename string
	Caption string
}