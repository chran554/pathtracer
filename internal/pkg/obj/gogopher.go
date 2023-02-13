package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"os"
	scn "pathtracer/internal/pkg/scene"
)

func NewGopher(scale *vec3.T) *scn.FacetStructure {
	return loadGopher(scale)
}

func loadGopher(scale *vec3.T) *scn.FacetStructure {
	var objFilename = "go_gopher_color.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objFilename

	objFile, err := os.Open(objFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", objFilenamePath, err.Error())
	}
	defer objFile.Close()

	obj, err := Read(objFile)
	ymin := obj.Bounds.Ymin
	ymax := obj.Bounds.Ymax
	obj.Translate(&vec3.T{0.0, -ymin, 0.0})       // feet touch the ground (xz-plane)
	obj.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize to height == 1.0

	obj.Scale(&vec3.Zero, scale)

	obj.UpdateBounds()
	//fmt.Printf("Gopher bounds: %+v\n", obj.Bounds)

	return obj
}
