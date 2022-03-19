package color

var (
	Black = Color{R: 0, G: 0, B: 0}
	White = Color{R: 1, G: 1, B: 1}
)

type Color struct{ R, G, B float32 }

func (c *Color) Copy() Color {
	color := *c
	return color
}

func (c *Color) Add(color Color) *Color {
	c.R += color.R
	c.G += color.G
	c.B += color.B
	return c
}

func (c *Color) Divide(divider float32) *Color {
	c.R /= divider
	c.G /= divider
	c.B /= divider
	return c
}

func (c *Color) Multiply(factor float32) *Color {
	c.R *= factor
	c.G *= factor
	c.B *= factor
	return c
}
