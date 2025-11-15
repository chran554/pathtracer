package obj

import (
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	scn "pathtracer/internal/pkg/scene"
	"strconv"
)

// NewHeightMap todo
func NewHeightMap(filename string, scale vec3.T) *scn.FacetStructure {
	landscape := loadHeightMap(filename)

	landscape.CenterOn(&vec3.Zero)
	landscape.Scale(&vec3.Zero, &vec3.T{landscape.Bounds.SizeX(), landscape.Bounds.SizeY(), landscape.Bounds.SizeZ()})
	landscape.Translate(&vec3.T{0, -landscape.Bounds.Ymin, 0})
	landscape.Scale(&vec3.Zero, &scale)

	return landscape
}

func loadHeightMap(filename string) *scn.FacetStructure {
	landscape := &scn.FacetStructure{}
	landscape.Material = scn.NewMaterial()

	img := floatimage.GetCachedImage(filename)

	pointMap := map[string]*vec3.T{}

	for y := 0; y < img.Height-1; y++ {
		for x := 0; x < img.Width-1; x++ {
			i00 := averageIntensity(img.GetPixel(x+0, (img.Height-1)-(y+0)))
			i10 := averageIntensity(img.GetPixel(x+1, (img.Height-1)-(y+0)))
			i11 := averageIntensity(img.GetPixel(x+1, (img.Height-1)-(y+1)))
			i01 := averageIntensity(img.GetPixel(x+0, (img.Height-1)-(y+1)))

			dx0 := float64(x+0) / float64(img.Width-1)
			dx1 := float64(x+1) / float64(img.Width-1)
			dy0 := float64(y+0) / float64(img.Height-1)
			dy1 := float64(y+1) / float64(img.Height-1)

			p1Index := strconv.Itoa(x+0) + ":" + strconv.Itoa(y+0)
			p2Index := strconv.Itoa(x+1) + ":" + strconv.Itoa(y+0)
			p3Index := strconv.Itoa(x+1) + ":" + strconv.Itoa(y+1)
			p4Index := strconv.Itoa(x+0) + ":" + strconv.Itoa(y+1)

			var ok bool
			var p1 *vec3.T
			var p2 *vec3.T
			var p3 *vec3.T
			var p4 *vec3.T

			if p1, ok = pointMap[p1Index]; !ok {
				point := &vec3.T{dx0, i00, dy0}
				pointMap[p1Index] = point
				p1 = point
			}

			if p2, ok = pointMap[p2Index]; !ok {
				point := &vec3.T{dx1, i10, dy0}
				pointMap[p2Index] = point
				p2 = point
			}

			if p3, ok = pointMap[p3Index]; !ok {
				point := &vec3.T{dx1, i11, dy1}
				pointMap[p3Index] = point
				p3 = point
			}

			if p4, ok = pointMap[p4Index]; !ok {
				point := &vec3.T{dx0, i01, dy1}
				pointMap[p4Index] = point
				p4 = point
			}

			var facets []*scn.Facet
			// Split the "square" facet into two triangles depending on the distance (in height) between opposite corners.
			if math.Abs(p1[1]-p3[1]) > math.Abs(p2[1]-p4[1]) {
				// If height difference of p1 and p3 is greater than the difference between p2 and p4
				// Then produce two triangles: (p1,p2,p4) and (p4,p2,p3)
				facets = getRectangleFacets(p1, p2, p3, p4)
			} else {
				// If height difference of p1 and p3 is less or equal than the difference between p2 and p4
				// Then produce two triangles: (p2,p3,p1) and (p1,p3,p4)
				facets = getRectangleFacets(p2, p3, p4, p1)
			}

			landscape.Facets = append(landscape.Facets, facets...)
		}
	}

	landscape.UpdateBounds()

	return landscape
}

func averageIntensity(c *color.Color) float64 {
	return float64(c.R+c.G+c.B) / 3.0
}

// GetRectangleFacets creates a "four corner facet" using four points (p1,p2,p3,p4).
// The result is two triangles side by side (p1,p2,p4) and (p4,p2,p3).
// Normal direction is calculated as pointing towards observer if the points are listed in counter-clockwise order.
// No test nor calculation is made that the points are exactly in the same plane.
func getRectangleFacets(p1, p2, p3, p4 *vec3.T) []*scn.Facet {
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
