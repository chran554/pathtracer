package obj

import (
	"math"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

func NewDragon01(scale float64) *scene.FacetStructure {
	dragon := loadDragon01()

	dragon.CenterOn(&vec3.Zero)
	dragon.RotateZ(&vec3.Zero, math.Pi/2)
	dragon.UpdateBounds()
	dragon.Translate(&vec3.T{0, -dragon.Bounds.Ymin, 0})
	dragon.UpdateBounds()

	dragon.ScaleUniform(&vec3.Zero, scale/dragon.Bounds.Ymax)
	dragon.UpdateBounds()
	dragon.ClearMaterials()

	/*	dragon.Material = scene.NewMaterial().
		N("Dragon").
		C(color.NewColor(0.95, 0.95, 0.97), 1.0).
		M(0.2, 0.05).
		T(1.0, true, scene.RefractionIndex_Glass)
	*/
	dragon.Material = scene.NewMaterial().N("dragon").
		C(color.NewColor(0.7, 0.6, 0.3)).
		M(0.4, 0.5)
	dragon.RotateY(&vec3.Zero, math.Pi/20)
	dragon.UpdateBounds()

	return dragon
}

func loadDragon01() *scene.FacetStructure {
	return wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "dragon_01.obj"))
}
