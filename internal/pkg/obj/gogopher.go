package obj

import (
	"github.com/ungerik/go3d/float64/vec3"
	scn "pathtracer/internal/pkg/scene"
)

func NewGopher(scale float64) *scn.FacetStructure {
	return loadGopher(scale)
}

func loadGopher(scale float64) *scn.FacetStructure {
	var objFilename = "go_gopher_color.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objFilename

	gopher := ReadOrPanic(objFilenamePath)

	ymin := gopher.Bounds.Ymin
	ymax := gopher.Bounds.Ymax
	gopher.Translate(&vec3.T{0.0, -ymin, 0.0})       // feet touch the ground (xz-plane)
	gopher.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize to height == 1.0

	gopher.ScaleUniform(&vec3.Zero, scale)

	return gopher
}
