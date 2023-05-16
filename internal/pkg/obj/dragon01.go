package obj

import (
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj2"
	scn "pathtracer/internal/pkg/scene"
)

func NewDragon01(scale float64) *scn.FacetStructure {
	dragon := loadDragon01()

	dragon.CenterOn(&vec3.Zero)
	dragon.RotateZ(&vec3.Zero, math.Pi/2)
	dragon.UpdateBounds()
	dragon.Translate(&vec3.T{0, -dragon.Bounds.Ymin, 0})
	dragon.UpdateBounds()

	dragon.ScaleUniform(&vec3.Zero, scale/dragon.Bounds.Ymax)
	dragon.UpdateBounds()
	dragon.ClearMaterials()

	/*	dragon.Material = scn.NewMaterial().
		N("Dragon").
		C(color.Color{R: 0.95, G: 0.95, B: 0.97}, 1.0).
		M(0.2, 0.05).
		T(1.0, true, scn.RefractionIndex_Glass)
	*/
	dragon.Material = scn.NewMaterial().N("dragon").
		C(color.Color{R: 0.7, G: 0.6, B: 0.3}).
		M(0.4, 0.5)
	dragon.RotateY(&vec3.Zero, math.Pi/20)
	dragon.UpdateBounds()

	return dragon
}

func loadDragon01() *scn.FacetStructure {
	return wavefrontobj2.ReadOrPanic(filepath.Join(ObjFileDir, "dragon_01.obj"))
}
