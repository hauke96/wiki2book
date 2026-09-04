package generator

type mockMathGenerator struct {
	RenderMathToSvgFunc func(mathText string) (string, error)
}

func newMockMathGenerator() *mockMathGenerator {
	return &mockMathGenerator{
		RenderMathToSvgFunc: func(mathText string) (string, error) { return "", nil },
	}
}

func (m *mockMathGenerator) RenderMathToSvg(mathText string) (string, error) {
	return m.RenderMathToSvgFunc(mathText)
}
