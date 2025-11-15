package obj

import (
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec2"
	"github.com/ungerik/go3d/float64/vec3"
)

type BoxType int

const (
	BoxCentered          BoxType = iota // BoxCentered is a box centered on origin (0,0,0) on axes. Each side is 2 unit in length from [-1, 1].
	BoxPositive                         // BoxPositive is a box with one corner in origin (0,0,0) and one corner in (1,1,1). Each side is 1 unit in length from [0, 1].
	BoxCenteredYPositive                // BoxCenteredYPositive is a box centered on origin (0,0,0) on the x and z axes (but not y-axis). Each side is 1 unit in length from [-0.5, 0.5], and [0, 1] for y axis.
)

// NewBoxWithEmission return a box which sides all have unit length 1.
// The box has a material called "light" on each side and a color emission.
// The sides of the box have prepared texture coordinates at each vertex: [0,0], [1,0], [1,1], [0,1].
func NewBoxWithEmission(boxType BoxType, c color.Color, scaleEmission float64, texture *floatimage.FloatImage) (*scn.FacetStructure, *scn.Material) {
	box := NewBox(boxType)

	box.Material = scn.NewMaterial().E(c, scaleEmission, true)
	if texture != nil {
		box.Material.TP(texture)
	}

	return box, box.Material
}

// NewBox return a box which sides all have the unit length 1.
func NewBox(boxType BoxType) *scn.FacetStructure {
	p1 := vec3.T{1, 1, 0} // Top right close            3----------2
	p2 := vec3.T{1, 1, 1} // Top right away            /          /|
	p3 := vec3.T{0, 1, 1} // Top left away            /          / |
	p4 := vec3.T{0, 1, 0} // Top left close          4----------1  |
	p5 := vec3.T{1, 0, 0} // Bottom right close      | (7)      |  6
	p6 := vec3.T{1, 0, 1} // Bottom right away       |          | /
	p7 := vec3.T{0, 0, 1} // Bottom left away        |          |/
	p8 := vec3.T{0, 0, 0} // Bottom left close       8----------5

	box := scn.FacetStructure{
		FacetStructures: []*scn.FacetStructure{
			{Facets: GetRectangleFacets(&p1, &p2, &p3, &p4), SubstructureName: "ymax"},
			{Facets: GetRectangleFacets(&p8, &p7, &p6, &p5), SubstructureName: "ymin"},
			{Facets: GetRectangleFacets(&p4, &p8, &p5, &p1), SubstructureName: "zmin"},
			{Facets: GetRectangleFacets(&p6, &p7, &p3, &p2), SubstructureName: "zmax"},
			{Facets: GetRectangleFacets(&p6, &p2, &p1, &p5), SubstructureName: "xmax"},
			{Facets: GetRectangleFacets(&p7, &p8, &p4, &p3), SubstructureName: "xmin"},
		},
	}

	if boxType == BoxCentered {
		box.Translate(&vec3.T{-0.5, -0.5, -0.5})
		box.Scale(&vec3.Zero, &vec3.T{2, 2, 2})
	} else if boxType == BoxCenteredYPositive {
		box.Translate(&vec3.T{-0.5, 0.0, -0.5})
	}

	box.UpdateBounds()
	box.UpdateNormals()

	return &box
}

// GetRectangleFacets creates a "four corner facet" using four points (p1,p2,p3,p4).
// The result is two triangles side by side (p1,p2,p4) and (p4,p2,p3).
// Normal direction is calculated as pointing towards observer if the points are listed in counter-clockwise order.
// No test nor calculation is made that the points are exactly in the same plane.
func GetRectangleFacets(p1, p2, p3, p4 *vec3.T) []*scn.Facet {
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
		{Vertices: []*vec3.T{p1, p2, p4}, Normal: &normal1, TextureCoordinates: []*vec2.T{{1, 0}, {1, 1}, {0, 0}}},
		{Vertices: []*vec3.T{p4, p2, p3}, Normal: &normal2, TextureCoordinates: []*vec2.T{{0, 0}, {1, 1}, {0, 1}}},
	}
}
