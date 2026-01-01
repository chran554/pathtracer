package obj

import (
	"math"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

func NewDragon02(scale float64, includeDragon bool, includePillar bool) *scene.FacetStructure {
	dragon := loadDragon02()

	if !includeDragon {
		dragon.RemoveObjectsByMaterialName("skin")
	}

	if !includePillar {
		dragon.RemoveObjectsByMaterialName("pillar")
	}

	dragon.CenterOn(&vec3.Zero)
	dragon.RotateX(&vec3.Zero, math.Pi/2)
	dragon.RotateY(&vec3.Zero, math.Pi)
	dragon.Translate(&vec3.T{0, -dragon.Bounds.Ymin, 0})

	dragon.ScaleUniform(&vec3.Zero, scale/dragon.Bounds.Ymax)

	goldColor := color.NewColor(1.0, 0.85, 0.58).Multiply(0.9)
	skinMaterial := scene.NewMaterial().N("skin").
		C(goldColor).
		M(0.30, 0.15).
		T(0.0, true, scene.RefractionIndex_Gold)
	skinMaterial.ColorizeReflection = true

	pillarMaterial := scene.NewMaterial().N("pillar").
		C(color.NewColor(0.8, 0.85, 0.7)).
		M(0.2, 0.6)

	dragon.ReplaceMaterial("skin", skinMaterial)
	dragon.ReplaceMaterial("pillar", pillarMaterial)

	return dragon
}

func loadDragon02() *scene.FacetStructure {
	dragon := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "dragon_02.obj"))
	return dragon
}
