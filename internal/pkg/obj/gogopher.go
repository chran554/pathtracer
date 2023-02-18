package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"os"
	scn "pathtracer/internal/pkg/scene"
)

func NewGopher(scale float64) *scn.FacetStructure {
	return loadGopher(scale)
}

func loadGopher(scale float64) *scn.FacetStructure {
	var objFilename = "go_gopher_color.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objFilename

	objFile, err := os.Open(objFilenamePath)
	if err != nil {
		message := fmt.Sprintf("ouupps, something went wrong loading file: '%s'\n%s\n", objFilenamePath, err.Error())
		panic(message)
	}
	defer objFile.Close()

	obj, err := Read(objFile)
	ymin := obj.Bounds.Ymin
	ymax := obj.Bounds.Ymax
	obj.Translate(&vec3.T{0.0, -ymin, 0.0})       // feet touch the ground (xz-plane)
	obj.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize to height == 1.0

	obj.ScaleUniform(&vec3.Zero, scale)

	obj.UpdateBounds()

	return obj
}
