package obj

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"os"
	"pathtracer/internal/pkg/scene"
	"testing"
)

func Test_tessellateSphere(t *testing.T) {
	t.Run("tessellation of facet", func(t *testing.T) {
		facet := &scene.FacetStructure{Facets: []*scene.Facet{{Vertices: []*vec3.T{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}}}}}
		tessellateSpherical(facet)

		assert.Equal(t, 4, len(facet.Facets))
	})
}

func Test_NewTessellatedSphere(t *testing.T) {
	amountStartFacets := 20 // icosahedron

	t.Run("tessellated sphere level 0", func(t *testing.T) {
		level := 0
		sphere := NewTessellatedSphere(level, true)

		expectedAmountFacets := int(math.Round(float64(amountStartFacets) * math.Pow(4, float64(level))))
		assert.Equal(t, expectedAmountFacets, len(sphere.Facets))
	})

	t.Run("tessellated sphere level 1", func(t *testing.T) {
		level := 1
		sphere := NewTessellatedSphere(level, true)

		expectedAmountFacets := int(math.Round(float64(amountStartFacets) * math.Pow(4, float64(level))))
		assert.Equal(t, expectedAmountFacets, len(sphere.Facets))
	})

	t.Run("tessellated sphere level 2", func(t *testing.T) {
		level := 2
		sphere := NewTessellatedSphere(level, true)

		expectedAmountFacets := int(math.Round(float64(amountStartFacets) * math.Pow(4, float64(level))))
		assert.Equal(t, expectedAmountFacets, len(sphere.Facets))
	})
}

func Test_TessellatedSphereObject(t *testing.T) {
	t.Run("saved tessellated sphere", func(t *testing.T) {
		level := 4
		sphere := NewTessellatedSphere(level, true)

		objFile := createFile("tessellated_sphere.obj")
		defer objFile.Close()
		mtlFile := createFile("tessellated_sphere.mtl")
		defer mtlFile.Close()

		WriteObjFile(objFile, mtlFile, sphere, nil)

		defer os.Remove(objFile.Name())
		defer os.Remove(mtlFile.Name())
	})
}

func createFile(name string) *os.File {
	objFile, err := os.Create(name)
	if err != nil {
		fmt.Printf("could not create file: '%s'\n%s\n", objFile.Name(), err.Error())
		os.Exit(1)
	}
	return objFile
}
