package obj

import (
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
)

func NewDragon02(scale float64, includeDragon bool, includePillar bool) *scn.FacetStructure {
	var objectFilename = "dragon_02.obj"
	var objectFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objectFilename

	dragon := ReadOrPanic(objectFilenamePath)
	if !includeDragon {
		dragon.RemoveObjectsByMaterialName("skin")
	}
	if !includePillar {
		dragon.RemoveObjectsByMaterialName("pillar")
	}

	dragon.CenterOn(&vec3.Zero)
	dragon.RotateX(&vec3.Zero, math.Pi/2)
	dragon.RotateY(&vec3.Zero, math.Pi)
	dragon.CenterOn(&vec3.Zero)
	dragon.Translate(&vec3.T{0, -dragon.Bounds.Ymin, 0})
	dragon.UpdateBounds()

	dragon.ScaleUniform(&vec3.Zero, scale/dragon.Bounds.Ymax)
	dragon.UpdateBounds()

	skinMaterial := scn.NewMaterial().N("skin").
		C(color.Color{R: 0.6, G: 0.5, B: 0.2}).
		M(0.3, 0.2)

	pillarMaterial := scn.NewMaterial().N("pillar").
		C(color.Color{R: 0.8, G: 0.85, B: 0.7}).
		M(0.2, 0.6)

	dragon.ReplaceMaterial("skin", skinMaterial)
	dragon.ReplaceMaterial("pillar", pillarMaterial)

	return dragon
}
