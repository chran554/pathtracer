package color

var (
	Black = Color{R: 0, G: 0, B: 0}
	White = Color{R: 1, G: 1, B: 1}
)

type Color struct{ R, G, B float64 }

func (c *Color) Add(color Color) *Color {
	c.R += color.R
	c.G += color.G
	c.B += color.B
	return c
}

func (c *Color) Divide(divider float64) *Color {
	c.R /= divider
	c.G /= divider
	c.B /= divider
	return c
}
