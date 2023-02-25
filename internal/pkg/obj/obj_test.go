package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"os"
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

	objFile, err := os.Open(objFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", objFilenamePath, err.Error())
		os.Exit(1)
	}
	defer objFile.Close()

	obj, err := Read(objFile)
	obj.Scale(&vec3.Zero, scale)

	obj.UpdateBounds()
	fmt.Printf("Test cube bounds: %+v\n", obj.Bounds)

	return obj
}