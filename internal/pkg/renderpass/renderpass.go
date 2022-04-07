package renderpass

import (
	"fmt"
	"math"
)

var QuadRenderPasses = RenderPasses{
	MaxPixelWidth:  2,
	MaxPixelHeight: 2,
	RenderPasses: []RenderPass{
		{Dx: 0, Dy: 0, PaintWidth: 2, PaintHeight: 2},
		{Dx: 1, Dy: 1, PaintWidth: -2, PaintHeight: 1},
		{Dx: 1, Dy: 0, PaintWidth: 1, PaintHeight: 1},
		{Dx: 0, Dy: 1, PaintWidth: 1, PaintHeight: 1},
	},
}

var NonaRenderPasses = RenderPasses{
	MaxPixelWidth:  3,
	MaxPixelHeight: 3,
	RenderPasses: []RenderPass{
		{Dx: 0, Dy: 0, PaintWidth: 3, PaintHeight: 3},
		{Dx: 1, Dy: 1, PaintWidth: 2, PaintHeight: 2},
		{Dx: 2, Dy: 0, PaintWidth: 1, PaintHeight: 2},
		{Dx: 0, Dy: 2, PaintWidth: 2, PaintHeight: 1},
		{Dx: 2, Dy: 1, PaintWidth: 1, PaintHeight: 2},

		{Dx: 0, Dy: 1, PaintWidth: 1, PaintHeight: 1},
		{Dx: 2, Dy: 2, PaintWidth: 1, PaintHeight: 1},
		{Dx: 1, Dy: 0, PaintWidth: 1, PaintHeight: 1},
		{Dx: 1, Dy: 2, PaintWidth: 1, PaintHeight: 1},
	},
}

type RenderPasses struct {
	RenderPasses   []RenderPass
	MaxPixelWidth  int
	MaxPixelHeight int
}

type RenderPass struct {
	Dx, Dy                  int
	width, height           int
	PaintWidth, PaintHeight int
}

func CreateRenderPasses(size int) RenderPasses {
	renderPasses := RenderPasses{
		MaxPixelWidth:  size,
		MaxPixelHeight: size,
	}

	var passes []RenderPass
	passes = append(passes, RenderPass{Dx: 0, Dy: 0, width: size, height: size, PaintWidth: size, PaintHeight: size})

	for index := 0; index < len(passes); index++ {
		pass := passes[index]
		x := pass.Dx
		y := pass.Dy
		w := pass.width
		h := pass.height

		if w == 1 && h == 1 {
			continue
		}

		nw := int(math.Max(float64(w/2), 1.0))
		nh := int(math.Max(float64(h/2), 1.0))

		if w >= h {
			passes = appendLegalPass(passes, RenderPass{Dx: x + nw, Dy: y, width: w - nw, height: nh, PaintWidth: w - nw, PaintHeight: h})
			passes = appendLegalPass(passes, RenderPass{Dx: x, Dy: y + nh, width: nw, height: h - nh, PaintWidth: nw, PaintHeight: h - nh})
			passes = appendLegalPass(passes, RenderPass{Dx: x + nw, Dy: y + nh, width: w - nw, height: h - nh, PaintWidth: w - nw, PaintHeight: h - nh})
			passes = appendLegalPass(passes, RenderPass{Dx: x, Dy: y, width: nw, height: nh, PaintWidth: nw, PaintHeight: nh})
		} else {
			passes = appendLegalPass(passes, RenderPass{Dx: x, Dy: y + nh, width: nw, height: h - nh, PaintWidth: w, PaintHeight: h - nh})
			passes = appendLegalPass(passes, RenderPass{Dx: x + nw, Dy: y, width: w - nw, height: nh, PaintWidth: w - nw, PaintHeight: nh})
			passes = appendLegalPass(passes, RenderPass{Dx: x + nw, Dy: y + nh, width: w - nw, height: h - nh, PaintWidth: w - nw, PaintHeight: h - nh})
			passes = appendLegalPass(passes, RenderPass{Dx: x, Dy: y, width: nw, height: nh, PaintWidth: nw, PaintHeight: nh})
		}
	}

	// Filter out duplicate passes. I.e. passes with (dx,dy) already set
	var filteredPasses []RenderPass
	setPositions := make(map[string]bool, size*size)
	for _, pass := range passes {
		positionKey := fmt.Sprintf("%dx%d", pass.Dx, pass.Dy)

		if !setPositions[positionKey] {
			filteredPasses = append(filteredPasses, pass)
			setPositions[positionKey] = true
		}
	}

	renderPasses.RenderPasses = filteredPasses
	return renderPasses
}

func appendLegalPass(passes []RenderPass, pass RenderPass) []RenderPass {
	if pass.PaintWidth > 0 && pass.PaintHeight > 0 {
		passes = append(passes, pass)
	}
	return passes
}
