package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"os"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
)

func NewSolidUtahTeapot(scale float64) *scn.FacetStructure {
	var objectFilename = "utah_teapot_solid_01.obj"
	var objectFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objectFilename

	objectFile, err := os.Open(objectFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", objectFilenamePath, err.Error())
	}
	defer objectFile.Close()

	object, err := Read(objectFile)
	//object.CenterOn(&vec3.Zero)
	object.Translate(&vec3.T{0, -object.Bounds.Ymin, 0})

	object.ScaleUniform(&vec3.Zero, scale/object.Bounds.Ymax)
	object.UpdateBounds()

	porcelainMaterial := scn.NewMaterial().
		N("Porcelain material").
		C(color.Color{R: 0.88, G: 0.96, B: 0.96}, 1.0).
		M(0.1, 0.1)

	// glassMaterial := scn.NewMaterial().
	// 	N("Glass material").
	// 	C(color.Color{R: 0.95, G: 0.95, B: 0.97}, 1.0).
	// 	M(0.2, 0.05).
	// 	T(1.0, true, scn.RefractionIndex_Glass)

	object.ClearMaterials()
	object.Material = porcelainMaterial

	return object
}
