package wavefrontobj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	scn "pathtracer/internal/pkg/scene"
	"testing"
)

func TestLoadFile_TestFile(t *testing.T) {
	t.Run("loading of obj-file", func(t *testing.T) {
		testCube := loadTestCube(&vec3.T{1, 1, 1})

		fmt.Printf("%+v\n", testCube)
	})
}

func loadTestCube(scale *vec3.T) *scn.FacetStructure {
	var objFilename = "test_cube.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/" + objFilename

	testCube := ReadOrPanic(objFilenamePath)

	testCube.Scale(&vec3.Zero, scale)

	fmt.Printf("Test cube bounds: %+v\n", testCube.Bounds)

	return testCube
}
