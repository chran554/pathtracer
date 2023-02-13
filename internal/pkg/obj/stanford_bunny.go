package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"os"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
)

func NewStanfordBunny(scale float64) *scn.FacetStructure {
	var objectFilename = "stanford_bunny.obj"
	var objectFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objectFilename

	objectFile, err := os.Open(objectFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", objectFilenamePath, err.Error())
	}
	defer objectFile.Close()

	object, err := Read(objectFile)

	object.CenterOn(&vec3.Zero)
	object.Translate(&vec3.T{0, -object.Bounds.Ymin, 0})
	object.UpdateBounds()

	object.ScaleUniform(&vec3.Zero, scale/object.Bounds.Ymax)
	object.UpdateBounds()

	skinMaterial := scn.NewMaterial().N("stanford_bunny").
		C(color.Color{R: 0.9, G: 0.85, B: 0.6}, 1.0).
		M(0.3, 0.6)

	object.ReplaceMaterial("stanford_bunny", skinMaterial)

	return object
}
