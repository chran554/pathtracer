package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"os"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
)

func NewDragon02(scale float64, includeDragon bool, includePillar bool) *scn.FacetStructure {
	var objectFilename = "dragon_02.obj"
	var objectFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objectFilename

	objectFile, err := os.Open(objectFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", objectFilenamePath, err.Error())
	}
	defer objectFile.Close()

	object, err := Read(objectFile)
	if !includeDragon {
		object.RemoveObjectsByMaterialName("skin")
	}
	if !includePillar {
		object.RemoveObjectsByMaterialName("pillar")
	}

	object.CenterOn(&vec3.Zero)
	object.RotateX(&vec3.Zero, math.Pi/2)
	object.RotateY(&vec3.Zero, math.Pi)
	object.CenterOn(&vec3.Zero)
	object.Translate(&vec3.T{0, -object.Bounds.Ymin, 0})
	object.UpdateBounds()

	object.ScaleUniform(&vec3.Zero, scale/object.Bounds.Ymax)
	object.UpdateBounds()

	skinMaterial := scn.NewMaterial().N("skin").
		C(color.Color{R: 0.6, G: 0.5, B: 0.2}, 1.0).
		M(0.3, 0.2)

	pillarMaterial := scn.NewMaterial().N("pillar").
		C(color.Color{R: 0.8, G: 0.85, B: 0.7}, 1.0).
		M(0.2, 0.6)

	object.ReplaceMaterial("skin", skinMaterial)
	object.ReplaceMaterial("pillar", pillarMaterial)

	return object
}
