package wavefrontobj

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// Expected structure:
//
// {on: UnitCube, sn: -, mn: -, f: 12}
func Test_CubeVanilla(t *testing.T) {
	t.Run("obj file: cube - vanilla", func(t *testing.T) {
		cube := loadTestCube("vanilla")
		fmt.Printf("Facet structure to be tested: %+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "UnitCube", "", "", 6*2, 6*2, 8, 0, expectedFullCubeBounds)
	})
}

// Expected structure:
//
// {on: UnitCube, sn: -, mn: -, f: 4}
//
//	{on: -, sn: -, mn: red, f: 4}
//	{on: -, sn: -, mn: green, f: 4}
//	{on: -, sn: -, mn: blue, f: 4}
func Test_CubeColors(t *testing.T) {
	t.Run("obj file: cube - colors", func(t *testing.T) {
		cube := loadTestCube("colors")
		fmt.Printf("Facet structure to be tested: %+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "UnitCube", "", "", 0, 12, 8, 3, expectedFullCubeBounds)

		assertFacetStructure(t, getSubstructure(t, cube, "", "", "red"), "", "", "red", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "", "", "green"), "", "", "green", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "", "", "blue"), "", "", "blue", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
	})
}

// Expected structure:
//
// {on: UnitCube, sn: -, mn: -, f: 4}
//
//	{on: -, sn: -, mn: green, f: 4}
//	{on: -, sn: -, mn: blue, f: 4}
func Test_CubeColors2(t *testing.T) {
	t.Run("obj file: cube - colors_2", func(t *testing.T) {
		cube := loadTestCube("colors_2")
		fmt.Printf("Facet structure to be tested: %+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "UnitCube", "", "", 4, 12, 8, 2, expectedFullCubeBounds)

		assertFacetStructure(t, getSubstructure(t, cube, "", "", "green"), "", "", "green", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "", "", "blue"), "", "", "blue", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
	})
}

// Expected structure:
//
// {on: UnitCube, sn: -, mn: -, f: 0}
//
//	{on: -, sn: x, mn: -, f: 4}
//	{on: -, sn: y, mn: -, f: 4}
//	{on: -, sn: z, mn: -, f: 4}
func Test_CubeGroups(t *testing.T) {
	t.Run("obj file: cube - groups", func(t *testing.T) {
		cube := loadTestCube("groups")
		fmt.Printf("Facet structure to be tested: %+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "UnitCube", "", "", 0, 12, 8, 3, expectedFullCubeBounds)

		assertFacetStructure(t, getSubstructure(t, cube, "", "x", ""), "", "x", "", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "", "y", ""), "", "y", "", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "", "z", ""), "", "z", "", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
	})
}

// Expected structure:
//
// {on: UnitCube, sn: -, mn: -, f: 4}
//
//	{on: y, sn: -, mn: -, f: 4}
//	{on: z, sn: -, mn: -, f: 4}
func Test_CubeGroups_2(t *testing.T) {
	t.Run("obj file: cube - groups_2", func(t *testing.T) {
		cube := loadTestCube("groups_2")
		fmt.Printf("Facet structure to be tested: %+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "UnitCube", "", "", 2*2, 12, 8, 2, expectedFullCubeBounds)

		assertFacetStructure(t, getSubstructure(t, cube, "", "y", ""), "", "y", "", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "", "z", ""), "", "z", "", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
	})
}

// Expected structure:
//
// {on: cube_groups_3, sn: -, mn: -, f: 4}
//
//	{on: -, sn: x, mn: -, f: 4}
//	{on: -, sn: y, mn: -, f: 4}
//	{on: -, sn: z, mn: -, f: 4}
func Test_CubeGroups_3(t *testing.T) {
	t.Run("obj file: cube - groups_3", func(t *testing.T) {
		cube := loadTestCube("groups_3")
		fmt.Printf("Facet structure to be tested: %+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "cube_groups_3", "", "", 0, 12, 8, 3, expectedFullCubeBounds)

		assertFacetStructure(t, getSubstructure(t, cube, "", "x", ""), "", "x", "", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "", "y", ""), "", "y", "", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "", "z", ""), "", "z", "", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
	})
}

// Expected structure:
//
// {on: UnitCube, sn: -, mn: -, f: 0}
//
//	{on: x, sn: -, mn: red, f: 4}
//	{on: y, sn: -, mn: green, f: 4}
//	{on: z, sn: -, mn: blue, f: 4}
func Test_CubeGroupsColors(t *testing.T) {
	t.Run("obj file: cube - groups colors", func(t *testing.T) {
		cube := loadTestCube("groups_colors")
		fmt.Printf("Facet structure to be tested: %+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "UnitCube", "", "", 0, 12, 8, 3, expectedFullCubeBounds)

		assertFacetStructure(t, getSubstructure(t, cube, "", "x", "red"), "", "x", "red", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "", "y", "green"), "", "y", "green", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "", "z", "blue"), "", "z", "blue", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
	})
}

// Expected structure:
//
// {on: UnitCube, sn: -, mn: -, f: 0}
//
//	{on:  , sn: -, mn: red,   f: 4}
//	{on: y, sn: -, mn: green, f: 4}
//	{on: z, sn: -, mn: blue,  f: 4}
func Test_CubeGroupsColors2(t *testing.T) {
	t.Run("obj file: cube - groups colors_2", func(t *testing.T) {
		cube := loadTestCube("groups_colors_2")
		fmt.Printf("Facet structure to be tested: %+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "UnitCube", "", "", 0, 12, 8, 3, expectedFullCubeBounds)

		assertFacetStructure(t, getSubstructure(t, cube, "", "", "red"), "", "", "red", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "", "y", "green"), "", "y", "green", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "", "z", "blue"), "", "z", "blue", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
	})
}

// Expected structure:
//
// {on: UnitCube, sn: -, mn: -, f: 0}
//
//	{on:  , sn: -, mn: red,   f: 4}
//	{on:  , sn: y, mn: -,     f: 0}
//	 {on:  , sn: -, mn: green, f: 4}
//	 {on:  , sn: -, mn: blue,  f: 4}
func Test_CubeGroupsColors3(t *testing.T) {
	t.Run("obj file: cube - groups colors_3", func(t *testing.T) {
		cube := loadTestCube("groups_colors_3")
		fmt.Printf("Facet structure to be tested: %+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "UnitCube", "", "", 0, 12, 8, 2, expectedFullCubeBounds)

		materialRed := getSubstructure(t, cube, "", "", "red")
		groupY := getSubstructure(t, cube, "", "y", "")
		groupYMaterialGreen := getSubstructure(t, groupY, "", "", "green")
		groupYMaterialBlue := getSubstructure(t, groupY, "", "", "blue")

		assertFacetStructure(t, materialRed, "", "", "red", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, groupY, "", "y", "", 0, 2*2*2, 8, 2, expectedFullCubeBounds)
		assertFacetStructure(t, groupYMaterialGreen, "", "", "green", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, groupYMaterialBlue, "", "", "blue", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
	})
}

// Expected structure:
//
// {on: UnitCube, sn: -, mn: -, f: 0}
//
//	{on:  , sn: -, mn: red,   f: 4}
//	{on: y, sn: -, mn: green, f: 4}
//	{on: z, sn: -, mn: green, f: 4}
func Test_CubeGroupsColors4(t *testing.T) {
	t.Run("obj file: cube - groups colors_4", func(t *testing.T) {
		cube := loadTestCube("groups_colors_4")
		fmt.Printf("Facet structure to be tested: %+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "UnitCube", "", "", 0, 12, 8, 3, expectedFullCubeBounds)

		assertFacetStructure(t, getSubstructure(t, cube, "", "", "red"), "", "", "red", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "", "y", "green"), "", "y", "green", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "", "z", "green"), "", "z", "green", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
	})
}

// Expected structure:
//
// {on: UnitCube, sn: -, mn: -, f: 0}
//
//	{on:  , sn: -, mn: red,   f: 4}
//	{on:  , sn: -, mn: green, f: 4}
//	{on: z, sn: -, mn: blue,  f: 4}
func Test_CubeGroupsColors5(t *testing.T) {
	t.Run("obj file: cube - groups colors_5", func(t *testing.T) {
		cube := loadTestCube("groups_colors_5")
		fmt.Printf("Facet structure to be tested: %+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "UnitCube", "", "", 0, 12, 8, 3, expectedFullCubeBounds)

		assertFacetStructure(t, getSubstructure(t, cube, "", "", "red"), "", "", "red", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "", "", "green"), "", "", "green", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "", "z", "blue"), "", "z", "blue", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
	})
}

// Expected structure:
//
// {on: UnitCube, sn: -, mn: -, f: 4}
//
//	{on:  , sn: -, mn: green, f: 4}
//	{on: z, sn: -, mn: blue,  f: 4}
func Test_CubeGroupsColors6(t *testing.T) {
	t.Run("obj file: cube - groups colors_6", func(t *testing.T) {
		cube := loadTestCube("groups_colors_6")
		fmt.Printf("Facet structure to be tested: %+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "UnitCube", "", "", 2*2, 12, 8, 2, expectedFullCubeBounds)

		assertFacetStructure(t, getSubstructure(t, cube, "", "", "green"), "", "", "green", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "", "z", "blue"), "", "z", "blue", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
	})
}

// Expected structure:
//
// {on: UnitCube, sn: -, mn: -, f: 0}
//
//	{on: -, sn: x, mn: -,    f: 0}
//	 {on: -, sn: -, mn: red,   f: 4}
//	 {on: -, sn: -, mn: green, f: 4}
//	{on: -, sn: z, mn: blue,  f: 4}
func Test_CubeGroupsColors7(t *testing.T) {
	t.Run("obj file: cube - groups colors_7", func(t *testing.T) {
		cube := loadTestCube("groups_colors_7")
		fmt.Printf("Facet structure to be tested: %+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "UnitCube", "", "", 0, 12, 8, 2, expectedFullCubeBounds)

		groupX := getSubstructure(t, cube, "", "x", "")
		groupXMaterialRed := getSubstructure(t, groupX, "", "", "red")
		groupXMaterialGreen := getSubstructure(t, groupX, "", "", "green")
		assertFacetStructure(t, groupX, "", "x", "", 0, 2*2*2, 8, 2, expectedFullCubeBounds)
		assertFacetStructure(t, groupXMaterialRed, "", "", "red", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, groupXMaterialGreen, "", "", "green", 2*2, 2*2, 8, 0, expectedFullCubeBounds)

		groupZ := getSubstructure(t, cube, "", "z", "blue")
		assertFacetStructure(t, groupZ, "", "z", "blue", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
	})
}

// Expected structure:
//
// {on: UnitCube, sn: -, mn: -, f: 0}
//
//	{on: -, sn: x, mn: -,    f: 0}
//	 {on: -, sn: -, mn: red,   f: 4}
//	 {on: -, sn: -, mn: green, f: 4}
//	{on: -, sn: z, mn: green,  f: 4}
func Test_CubeGroupsColors8(t *testing.T) {
	t.Run("obj file: cube - groups colors_8", func(t *testing.T) {
		cube := loadTestCube("groups_colors_8")
		fmt.Printf("Facet structure to be tested: %+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "UnitCube", "", "", 0, 12, 8, 2, expectedFullCubeBounds)

		groupX := getSubstructure(t, cube, "", "x", "")
		groupXMaterialRed := getSubstructure(t, groupX, "", "", "red")
		groupXMaterialGreen := getSubstructure(t, groupX, "", "", "green")
		assertFacetStructure(t, groupX, "", "x", "", 0, 2*2*2, 8, 2, expectedFullCubeBounds)
		assertFacetStructure(t, groupXMaterialRed, "", "", "red", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, groupXMaterialGreen, "", "", "green", 2*2, 2*2, 8, 0, expectedFullCubeBounds)

		groupZ := getSubstructure(t, cube, "", "z", "green")
		assertFacetStructure(t, groupZ, "", "z", "green", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
	})
}

// Expected structure:
//
// {on: UnitCube, sn: -, mn: -, f: 0}
//
//	{on: -, sn: x, mn: red,    f: 4}
//	{on: -, sn: y, mn: -, f: 0}
//	 {on: -, sn: -, mn: green, f: 4}
//	 {on: -, sn: -, mn: blue,  f: 4}
func Test_CubeGroupsColors9(t *testing.T) {
	t.Run("obj file: cube - groups colors_9", func(t *testing.T) {
		cube := loadTestCube("groups_colors_9")
		fmt.Printf("Facet structure to be tested: %+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "UnitCube", "", "", 0, 12, 8, 2, expectedFullCubeBounds)

		groupXMaterialRed := getSubstructure(t, cube, "", "x", "red")
		groupY := getSubstructure(t, cube, "", "y", "")
		groupYMaterialGreen := getSubstructure(t, groupY, "", "", "green")
		groupYMaterialBlue := getSubstructure(t, groupY, "", "", "blue")
		assertFacetStructure(t, groupXMaterialRed, "", "x", "red", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, groupY, "", "y", "", 0, 2*2*2, 8, 2, expectedFullCubeBounds)
		assertFacetStructure(t, groupYMaterialGreen, "", "", "green", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, groupYMaterialBlue, "", "", "blue", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
	})
}

// Expected structure:
//
// {on: cube_objects_colors, sn: -, mn: -, f: 0}
//
//	{on: x, sn: -, mn: red, f: 4}
//	{on: y, sn: -, mn: green, f: 4}
//	{on: z, sn: -, mn: blue, f: 4}
func Test_CubeObjects(t *testing.T) {
	t.Run("obj file: cube - objects", func(t *testing.T) {
		cube := loadTestCube("objects")
		fmt.Printf("Facet structure to be tested: %+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "cube_objects", "", "", 0, 12, 8, 3, expectedFullCubeBounds)

		assertFacetStructure(t, getSubstructure(t, cube, "x", "", ""), "x", "", "", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "y", "", ""), "y", "", "", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "z", "", ""), "z", "", "", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
	})
}

// Expected structure:
//
// {on: cube_objects_colors, sn: -, mn: -, f: 0}
//
//	{on: x, sn: -, mn: red, f: 4}
//	{on: y, sn: -, mn: green, f: 4}
//	{on: z, sn: -, mn: blue, f: 4}
func Test_CubeObjects2(t *testing.T) {
	t.Run("obj file: cube - objects_2", func(t *testing.T) {
		cube := loadTestCube("objects_2")
		fmt.Printf("Facet structure to be tested: %+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "cube_objects_2", "", "", 0, 12, 8, 3, expectedFullCubeBounds)

		assertFacetStructure(t, getSubstructure(t, cube, "UnitCube", "", ""), "UnitCube", "", "", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "y", "", ""), "y", "", "", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "z", "", ""), "z", "", "", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
	})
}

// Expected structure:
//
// {on: cube_objects_colors, sn: -, mn: -, f: 0}
//
//	{on: x, sn: -, mn: red, f: 4}
//	{on: y, sn: -, mn: green, f: 4}
//	{on: z, sn: -, mn: blue, f: 4}
func Test_CubeObjectsColors(t *testing.T) {
	t.Run("obj file: cube - objects_colors", func(t *testing.T) {
		cube := loadTestCube("objects_colors")
		fmt.Printf("Facet structure to be tested: %+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "cube_objects_colors", "", "", 0, 12, 8, 3, expectedFullCubeBounds)

		assertFacetStructure(t, getSubstructure(t, cube, "x", "", "red"), "x", "", "red", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "y", "", "green"), "y", "", "green", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "z", "", "blue"), "z", "", "blue", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
	})
}

// Expected structure:
//
// {on: UnitCube, sn: -, mn: -, f: 0}
//
//	{on: -, sn: x, mn: -, f: 4}
//	{on: -, sn: y, mn: -, f: 4}
//	{on: -, sn: z, mn: -, f: 4}
func Test_CubeObjectsMaterials2(t *testing.T) {
	t.Run("obj file: cube - objects_materials_2", func(t *testing.T) {
		cube := loadTestCube("objects_materials_2")
		fmt.Printf("Facet structure to be tested: %+v\n", cube)
		expectedFullCubeBounds := scn.Bounds{Xmin: -1, Xmax: 1, Ymin: -1, Ymax: 1, Zmin: -1, Zmax: 1}
		assertFacetStructure(t, cube, "UnitCube", "", "", 0, 12, 8, 3, expectedFullCubeBounds)

		assertFacetStructure(t, getSubstructure(t, cube, "", "x", "red"), "", "x", "red", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "", "y", "red"), "", "y", "red", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
		assertFacetStructure(t, getSubstructure(t, cube, "", "z", "red"), "", "z", "red", 2*2, 2*2, 8, 0, expectedFullCubeBounds)
	})
}

func getSubstructure(t *testing.T, structure *scn.FacetStructure, objectName, groupName, materialName string) *scn.FacetStructure {
	var substructure *scn.FacetStructure

	for _, facetSubStructure := range structure.FacetStructures {
		if (objectName == facetSubStructure.Name) &&
			(groupName == facetSubStructure.SubstructureName) &&
			((facetSubStructure.Material == nil && materialName == "") || (facetSubStructure.Material != nil && facetSubStructure.Material.Name == materialName)) {
			substructure = facetSubStructure
		}
	}

	require.NotNil(t, substructure, "Could not find expected substructure \"%s::%s::%s\" of facet structure \"%s\".", objectName, groupName, materialName, structure.StructureNames())

	return substructure
}

func assertFacetStructure(t *testing.T, f *scn.FacetStructure, name, substructureName, materialName string, amountFacets, amountRecursiveFacets, amountFacetVertices, amountSubStructures int, expectedBounds scn.Bounds) {
	require.NotNil(t, f)

	assert.Equal(t, name, f.Name, "Facet structure \"%s\" do not have expected name.", f.StructureNames())
	assert.Equal(t, substructureName, f.SubstructureName, "Facet structure \"%s\" do not have expected sub structure name.", f.StructureNames())
	if f.Material == nil {
		assert.Equal(t, materialName, "")
	} else {
		assert.Equal(t, materialName, f.Material.Name)
	}

	if amountFacets == 0 {
		assert.Equal(t, 0, len(f.Facets), "The amount of immediate facets on structure \"%s\" is not 0 as expected.", f.StructureNames())
		assert.Nil(t, f.Facets, "Facet slice is expected to be nil (not slice of size 0) when there are no facets on structure \"%s\".", f.StructureNames())
	}
	assert.Equal(t, amountFacets, len(f.Facets), "The amount of immediate facets on structure \"%s\" do not match expected amount.", f.StructureNames())

	if amountRecursiveFacets == 0 {
		assert.Nil(t, f.Facets, "Facet slice is expected to be nil (not slice of size 0) when there are no recursive facets on structure \"%s\".", f.StructureNames())
	}
	assert.Equal(t, amountRecursiveFacets, f.GetAmountFacets(), "The amount of recursive facets on structure \"%s\" do not match expected amount.", f.StructureNames())

	if amountSubStructures == 0 {
		assert.Nil(t, f.FacetStructures)
	} else {
		assert.NotNil(t, f.FacetStructures)
		assert.Equal(t, amountSubStructures, len(f.FacetStructures), "Amount of substructures of facet structure \"%s\" differ from expected.", f.StructureNames())
	}

	if materialName == "" {
		assert.Nil(t, f.Material)
	} else {
		assert.NotNil(t, f.Material)
		assert.Equal(t, materialName, f.Material.Name)
	}

	assertBounds(t, expectedBounds, f.Bounds)

	// Get the recursive amount of unique vertices for the facet structure f
	vertices := make(map[*vec3.T]bool)
	facetStructures := []*scn.FacetStructure{f}
	for len(facetStructures) > 0 {
		facetStructure := facetStructures[0]
		facetStructures = facetStructures[1:]

		for _, facet := range facetStructure.Facets {
			for _, vertex := range facet.Vertices {
				vertices[vertex] = true
			}
		}

		facetStructures = append(facetStructures, facetStructure.FacetStructures...)
	}

	assert.Equal(t, amountFacetVertices, len(vertices))
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
