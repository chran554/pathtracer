package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"pathtracer/internal/pkg/scene"
)

// NewTessellatedSphere creates a tessellated sphere (tessellated by triangles).
// Level is the tesselation level (triangle subdivision level).
//
// Level 0 is an icosahedron with 20 triangles.
//
// Level 1 is a tessellated sphere with 20*4=80 triangles.
//
// Level 2 is a tessellated sphere with 20*4*4=320 triangles.
//
// Level 3 is a tessellated sphere with 20*4*4*4=1280 triangles.
//
// Amount triangles at each level is: <triangle count> = <start_amount_sides == 4> * (4^level)
func NewTessellatedSphere(level int, useVertexNormals bool) *scene.FacetStructure {
	tessellatedSphere := NewIcosahedron()
	tessellatedSphere.SubstructureName = fmt.Sprintf("tessellated sphere level %d from icosahedron", level)

	for i := 0; i < level; i++ {
		tessellateSpherical(tessellatedSphere)
	}

	tessellatedSphere.UpdateBounds()
	if useVertexNormals {
		tessellatedSphere.UpdateVertexNormals(false)
	}

	return tessellatedSphere
}

func NewOctahedron() *scene.FacetStructure {
	var facets []*scene.Facet
	pt := &vec3.T{0, 1, 0}  // pt is point at top
	pb := &vec3.T{0, -1, 0} // pb is point at bottom

	for i := 0; i < 4; i++ {
		angle1 := float64(i) * math.Pi / 2.0
		angle2 := float64(i+1) * math.Pi / 2.0
		p1 := &vec3.T{math.Cos(angle1), 0, math.Sin(angle1)}
		p2 := &vec3.T{math.Cos(angle2), 0, math.Sin(angle2)}
		facets = append(facets, &scene.Facet{Vertices: []*vec3.T{pt, p1, p2}})
		facets = append(facets, &scene.Facet{Vertices: []*vec3.T{pb, p2, p1}})
	}

	octahedron := &scene.FacetStructure{SubstructureName: "octahedron", Facets: facets}

	return octahedron
}

func NewIcosahedron() *scene.FacetStructure {
	var facets []*scene.Facet

	pt := &vec3.T{0, 1, 0}   // pt is point at top
	pb := &vec3.T{0, -1, 0}  // pb is point at bottom
	tp := make([]*vec3.T, 5) // points on the upper (top) ring
	bp := make([]*vec3.T, 5) // points on the lower (bottom) ring

	h := math.Sin(math.Pi / 6)
	r := math.Cos(math.Pi / 6)
	a := math.Pi * 2 / 5
	for i := 0; i < 5; i++ {
		angleTop := float64(i) * a
		angleBottom := float64(i)*a + (a / 2)
		tp[i] = &vec3.T{r * math.Cos(angleTop), h, r * math.Sin(angleTop)}
		bp[i] = &vec3.T{r * math.Cos(angleBottom), -h, r * math.Sin(angleBottom)}
	}

	// Top facets
	for i := 0; i < len(tp); i++ {
		facets = append(facets, &scene.Facet{Vertices: []*vec3.T{pt, tp[i%len(tp)], tp[(i+1)%len(tp)]}})
	}

	// Middle facets
	for i := 0; i < len(tp); i++ {
		facets = append(facets, &scene.Facet{Vertices: []*vec3.T{tp[i%len(tp)], bp[i%len(bp)], tp[(i+1)%len(tp)]}})
		facets = append(facets, &scene.Facet{Vertices: []*vec3.T{bp[i%len(bp)], bp[(i+1)%len(bp)], tp[(i+1)%len(tp)]}})
	}

	// Bottom facets
	for i := 0; i < len(bp); i++ {
		facets = append(facets, &scene.Facet{Vertices: []*vec3.T{pb, bp[(i+1)%len(bp)], bp[i%len(bp)]}})
	}

	icosahedron := &scene.FacetStructure{SubstructureName: "icosahedron", Facets: facets}

	return icosahedron
}

func tessellateSpherical(sphere *scene.FacetStructure) {
	type pointPair struct {
		p1, p2 *vec3.T
	}

	// Create new points on every point edge, and store them in a map for easy lookup
	newPoints := make(map[pointPair]*vec3.T)
	for _, facet := range sphere.Facets {
		amountVertices := len(facet.Vertices)

		if amountVertices != 3 {
			continue
		}

		for i := range facet.Vertices {
			p1 := facet.Vertices[i%amountVertices]
			p2 := facet.Vertices[(i+1)%amountVertices]

			pair12 := pointPair{p1, p2}
			pair21 := pointPair{p2, p1}
			_, newPointExists := newPoints[pair12]
			if !newPointExists {
				newPoint := vec3.Interpolate(p1, p2, 0.5)
				newPoint = projectOnUnitSphere(newPoint)
				newPoints[pair12] = &newPoint
				newPoints[pair21] = &newPoint
			}
		}
	}

	// Create 4 new triangles for each original triangle
	var newFacets []*scene.Facet = nil
	for _, facet := range sphere.Facets {
		if len(facet.Vertices) != 3 {
			continue
		}

		p1 := facet.Vertices[0]
		p2 := facet.Vertices[1]
		p3 := facet.Vertices[2]

		p12 := newPoints[pointPair{p1, p2}]
		p23 := newPoints[pointPair{p2, p3}]
		p31 := newPoints[pointPair{p3, p1}]

		newFacets = append(newFacets, &scene.Facet{Vertices: []*vec3.T{p1, p12, p31}})
		newFacets = append(newFacets, &scene.Facet{Vertices: []*vec3.T{p12, p2, p23}})
		newFacets = append(newFacets, &scene.Facet{Vertices: []*vec3.T{p12, p23, p31}})
		newFacets = append(newFacets, &scene.Facet{Vertices: []*vec3.T{p31, p23, p3}})
	}

	sphere.Facets = newFacets // replace old facets with new subdividing facets
}

func projectOnUnitSphere(point vec3.T) vec3.T {
	ray := &scene.Ray{Origin: &vec3.Zero, Heading: &point}
	sphere := &scene.Sphere{Origin: &vec3.Zero, Radius: 1.0}

	intersection, hit := scene.SphereIntersection(ray, sphere)
	if hit {
		return *intersection
	} else {
		return point
	}
}
