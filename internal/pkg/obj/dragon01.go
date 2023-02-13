package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"os"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
)

func NewDragon01(scale float64) *scn.FacetStructure {
	var objectFilename = "dragon_01.obj"
	var objectFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objectFilename

	objectFile, err := os.Open(objectFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", objectFilenamePath, err.Error())
	}
	defer objectFile.Close()

	object, err := Read(objectFile)
	object.CenterOn(&vec3.Zero)
	object.RotateZ(&vec3.Zero, math.Pi/2)
	object.UpdateBounds()
	object.Translate(&vec3.T{0, -object.Bounds.Ymin, 0})
	object.UpdateBounds()

	object.ScaleUniform(&vec3.Zero, scale/object.Bounds.Ymax)
	object.UpdateBounds()
	object.ClearMaterials()

	/*	object.Material = scn.NewMaterial().
		N("Dragon").
		C(color.Color{R: 0.95, G: 0.95, B: 0.97}, 1.0).
		M(0.2, 0.05).
		T(1.0, true, scn.RefractionIndex_Glass)
	*/
	object.Material = scn.NewMaterial().N("dragon").
		C(color.Color{R: 0.7, G: 0.6, B: 0.3}, 1.0).
		M(0.4, 0.5)
	object.RotateY(&vec3.Zero, math.Pi/20)
	object.UpdateBounds()

	return object
}
