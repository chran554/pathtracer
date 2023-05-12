package wavefrontobj

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/ungerik/go3d/float64/vec3"
	"path/filepath"
	scn "pathtracer/internal/pkg/scene"
	"testing"
)

func Test_LoadFile(t *testing.T) {
	t.Run("loading of obj-file (cube - vanilla)", func(t *testing.T) {
		cube := loadTestCube("vanilla")
		assert.NotNil(t, cube)
	})
}

func Test_CubeVanilla(t *testing.T) {
	t.Run("obj file: cube - vanilla", func(t *testing.T) {
		cube := loadTestCube("vanilla")
		fmt.Printf("%+v\n", cube)
		assertFacetStructure(t, cube, "UnitCube", "", 6*2, 8, 0, "", scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1})
	})
}

func Test_CubeColors(t *testing.T) {
	t.Run("obj file: cube - colors", func(t *testing.T) {
		cube := loadTestCube("colors")
		fmt.Printf("%+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "UnitCube", "", 0, 8, 3, "", expectedFullCubeBounds)

		assertFacetStructure(t, getSubstructureBySubstructureName(t, "cube_colors::red", cube), "", "cube_colors::red", 2*2, 8, 0, "red", expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructureBySubstructureName(t, "cube_colors::green", cube), "", "cube_colors::green", 2*2, 8, 0, "green", expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructureBySubstructureName(t, "cube_colors::blue", cube), "", "cube_colors::blue", 2*2, 8, 0, "blue", expectedFullCubeBounds)
	})
}

func Test_CubeGroups(t *testing.T) {
	t.Run("obj file: cube - groups", func(t *testing.T) {
		cube := loadTestCube("groups")
		fmt.Printf("%+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "UnitCube", "", 0, 8, 3, "", expectedFullCubeBounds)

		assertFacetStructure(t, getSubstructureBySubstructureName(t, "x", cube), "", "x", 2*2, 8, 0, "", expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructureBySubstructureName(t, "y", cube), "", "y", 2*2, 8, 0, "", expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructureBySubstructureName(t, "z", cube), "", "z", 2*2, 8, 0, "", expectedFullCubeBounds)
	})
}

func Test_CubeGroups_2(t *testing.T) {
	t.Run("obj file: cube - groups_2", func(t *testing.T) {
		cube := loadTestCube("groups_2")
		fmt.Printf("%+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "UnitCube", "", 8, 8, 1, "", expectedFullCubeBounds)

		assertFacetStructure(t, getSubstructureBySubstructureName(t, "z", cube), "", "z", 2*2, 8, 0, "", expectedFullCubeBounds)
	})
}

func getSubstructureBySubstructureName(t *testing.T, name string, structure *scn.FacetStructure) *scn.FacetStructure {
	var substructure *scn.FacetStructure

	for _, facetSubStructure := range structure.FacetStructures {
		if name == facetSubStructure.SubstructureName {
			substructure = facetSubStructure
		}
	}

	assert.NotNil(t, substructure)

	return substructure
}

func assertFacetStructure(t *testing.T, f *scn.FacetStructure, name, substructureName string, amountFacets, amountFacetVertices, amountSubStructures int, materialName string, expectedBounds scn.Bounds) {
	assert.NotNil(t, f)

	assert.Equal(t, name, f.Name)
	assert.Equal(t, substructureName, f.SubstructureName)

	if amountFacets == 0 {
		assert.Nil(t, f.Facets)
	} else {
		assert.NotNil(t, f.Facets)
		assert.Equal(t, amountFacets, f.GetAmountFacets())
	}

	if amountSubStructures == 0 {
		assert.Nil(t, f.FacetStructures)
	} else {
		assert.NotNil(t, f.FacetStructures)
		assert.Equal(t, amountSubStructures, len(f.FacetStructures))
	}

	if materialName == "" {
		assert.Nil(t, f.Material)
	} else {
		assert.NotNil(t, f.Material)
		assert.Equal(t, materialName, f.Material.Name)
	}

	assertBounds(t, expectedBounds, f.Bounds)

	if amountFacets > 0 {
		vertices := make(map[*vec3.T]bool)
		for _, facet := range f.Facets {
			for _, vertex := range facet.Vertices {
				vertices[vertex] = true
			}
		}

		assert.Equal(t, amountFacetVertices, len(vertices))
	}
}

func assertBounds(t *testing.T, expectedBounds scn.Bounds, actualBounds *scn.Bounds) {
	assert.NotNil(t, actualBounds)

	assert.Equal(t, expectedBounds.Xmin, actualBounds.Xmin)
	assert.Equal(t, expectedBounds.Ymin, actualBounds.Ymin)
	assert.Equal(t, expectedBounds.Zmin, actualBounds.Zmin)
	assert.Equal(t, expectedBounds.Xmax, actualBounds.Xmax)
	assert.Equal(t, expectedBounds.Ymax, actualBounds.Ymax)
	assert.Equal(t, expectedBounds.Zmax, actualBounds.Zmax)
}

func loadTestCube(flavour string) *scn.FacetStructure {
	var objFilename = "cube_" + flavour + ".obj"
	var objFilenamePath = filepath.Join("../../../../objects/obj/", "test", objFilename)

	testCube := ReadOrPanic(objFilenamePath)

	return testCube
}
