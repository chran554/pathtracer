package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	scn "pathtracer/internal/pkg/scene"
)

// NewDiamond creates a new diamond with radius set to scale value.
func NewDiamond(scale float64) *scn.FacetStructure {
	return loadDiamond(scale)
}

func loadDiamond(scale float64) *scn.FacetStructure {
	var objFilename = "diamond.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objFilename

	diamond := ReadOrPanic(objFilenamePath)

	xmin := diamond.Bounds.Xmin
	xmax := diamond.Bounds.Xmax
	ymin := diamond.Bounds.Ymin

	diamond.Translate(&vec3.T{0.0, -ymin, 0.0})       // feet touch the ground (xz-plane)
	diamond.ScaleUniform(&vec3.Zero, 1.0/(xmax-xmin)) // resize to girdle radius == 0.5 (diameter == 1.0)
	diamond.ScaleUniform(&vec3.Zero, 2*scale)         // resize to girdle diameter of scale value

	fmt.Printf("Diamond bounds: %+v\n", diamond.Bounds)

	return diamond
}