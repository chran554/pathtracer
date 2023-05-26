package obj

import (
	"github.com/ungerik/go3d/float64/vec3"
	scn "pathtracer/internal/pkg/scene"
)

type SquareType int

const (
	XYPlane SquareType = iota // Square lower left corner is located at (0,0,0) on axes. Each side is 1 unit in length from [0, 1]. Normal parallel to z-axis.
	YZPlane                   // Square lower left corner is located at (0,0,0) on axes. Each side is 1 unit in length from [0, 1]. Normal parallel to x-axis.
	XZPlane                   // Square lower left corner is located at (0,0,0) on axes. Each side is 1 unit in length from [0, 1]. Normal anti-parallel to y-axis (parallel to inverted direction of y-axis).
)

func NewSquare(squareType SquareType) []*scn.Facet {
	p1 := vec3.T{1, 1, 0} // Top right close            3----------2
	// p2 := vec3.T{1, 1, 1} // Top right away         /          /|         y
	p3 := vec3.T{0, 1, 1} // Top left away            /          / |         ^
	p4 := vec3.T{0, 1, 0} // Top left close          4----------1  |         |   z
	p5 := vec3.T{1, 0, 0} // Bottom right close      | (7)      |  6         |  7
	p6 := vec3.T{1, 0, 1} // Bottom right away       |          | /          | /
	p7 := vec3.T{0, 0, 1} // Bottom left away        |          |/           |/
	p8 := vec3.T{0, 0, 0} // Bottom left close       8----------5            *----> x

	switch squareType {
	case XYPlane:
		return GetRectangleFacets(&p4, &p8, &p5, &p1)
	case YZPlane:
		return GetRectangleFacets(&p7, &p8, &p4, &p3)
	case XZPlane:
		fallthrough
	default:
		return GetRectangleFacets(&p8, &p7, &p6, &p5)
	}
}

// getSquareFacets creates a "four corner facet" using four points (p1,p2,p3,p4).
// The result is two triangles side by side (p1,p2,p4) and (p4,p2,p3).
// Normal direction is calculated as pointing towards observer if the points are listed in counter-clockwise order.
// No test nor calculation is made that the points are exactly in the same plane.
func getSquareFacets(p1, p2, p3, p4 *vec3.T) []*scn.Facet {
	//       p1
	//       *
	//      / \
	//     /   \
	// p2 *-----* p4
	//     \   /
	//      \ /
	//       *
	//      p3
	//
	// (Normal calculated for each sub-triangle and s aimed towards observer.)

	n1v1 := p2.Subed(p1)
	n1v2 := p4.Subed(p1)
	normal1 := vec3.Cross(&n1v1, &n1v2)
	normal1.Normalize()

	n2v1 := p4.Subed(p3)
	n2v2 := p2.Subed(p3)
	normal2 := vec3.Cross(&n2v1, &n2v2)
	normal2.Normalize()

	return []*scn.Facet{
		{Vertices: []*vec3.T{p1, p2, p4}, Normal: &normal1},
		{Vertices: []*vec3.T{p4, p2, p3}, Normal: &normal2},
	}
}
