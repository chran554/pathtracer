package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"os"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
)

func NewDrWhoAngel(scale float64) *scn.FacetStructure {
	var objectFilename = "drwho_angel.obj"
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

	statueMaterial := scn.NewMaterial().N("angel").
		C(color.Color{R: 0.9, G: 0.9, B: 0.9}).
		M(0.3, 0.6)

	pillarMaterial := scn.NewMaterial().N("pillar").
		C(color.Color{R: 0.8, G: 0.8, B: 0.8}).
		M(0.1, 0.8)

	object.ReplaceMaterial("angel", statueMaterial)
	object.ReplaceMaterial("pillar", pillarMaterial)

	return object
}
