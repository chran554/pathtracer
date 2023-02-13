package scene

import (
	"github.com/stretchr/testify/assert"
	"github.com/ungerik/go3d/float64/vec3"
	"testing"
)

func TestFacet_ChangeWindingOrder_triangle(t *testing.T) {
	t.Run("winding order of three point (triangle) facet", func(t *testing.T) {
		p1 := &vec3.T{0, 0, 0}
		p2 := &vec3.T{1, 0, 0}
		p3 := &vec3.T{0, 1, 0}
		facet := Facet{Vertices: []*vec3.T{p1, p2, p3}}

		facet.ChangeWindingOrder()

		assert.Equal(t, facet.Vertices[0], p3)
		assert.Equal(t, facet.Vertices[1], p2)
		assert.Equal(t, facet.Vertices[2], p1)
	})
}

func TestFacet_ChangeWindingOrder_4p(t *testing.T) {
	t.Run("winding order of 4 point facet", func(t *testing.T) {
		p0 := &vec3.T{0, 0, 0}
		p1 := &vec3.T{1, 0, 0}
		p2 := &vec3.T{1, 0, 1}
		p3 := &vec3.T{0, 0, 1}
		facet := Facet{Vertices: []*vec3.T{p0, p1, p2, p3}} // Square in xz-plane

		facet.ChangeWindingOrder()

		assert.Equal(t, facet.Vertices[0], p3)
		assert.Equal(t, facet.Vertices[1], p2)
		assert.Equal(t, facet.Vertices[2], p1)
		assert.Equal(t, facet.Vertices[3], p0)
	})
}

func TestFacet_ChangeWindingOrder_5p(t *testing.T) {
	t.Run("winding order of 4 point facet", func(t *testing.T) {
		p0 := &vec3.T{0, 0, 0}
		p1 := &vec3.T{1, 0, 0}
		p2 := &vec3.T{1, 0, 1}
		p3 := &vec3.T{0, 0, 2}
		p4 := &vec3.T{-1, 0, 1}
		facet := Facet{Vertices: []*vec3.T{p0, p1, p2, p3, p4}} // 5 point facet in xz-plane

		facet.ChangeWindingOrder()

		assert.Equal(t, facet.Vertices[0], p4)
		assert.Equal(t, facet.Vertices[1], p3)
		assert.Equal(t, facet.Vertices[2], p2)
		assert.Equal(t, facet.Vertices[3], p1)
		assert.Equal(t, facet.Vertices[4], p0)
	})
}
