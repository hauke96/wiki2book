package generator

type MathGenerator interface {
	// RenderMathToSvg turns the given math expression into an SVG image. The path to the file in the cache and an error
	// are returned. The path is empty in case there is an error.
	RenderMathToSvg(mathText string) (string, error)
}

type MathGeneratorImpl struct {
}

func NewMathGenerator() *MathGeneratorImpl {
	return &MathGeneratorImpl{}
}

func (m *MathGeneratorImpl) RenderMathToSvg(mathText string) (string, error) {
	return "foobar", nil
}
