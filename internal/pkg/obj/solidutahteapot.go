package obj

import (
	"github.com/ungerik/go3d/float64/vec3"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"
)

func NewSolidUtahTeapot(scale float64) *scn.FacetStructure {
	utahTeaPot := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "utah_teapot_solid_02.obj"))

	utahTeaPot.CenterOn(&vec3.Zero)
	utahTeaPot.Translate(&vec3.T{0, -utahTeaPot.Bounds.Ymin, 0})

	utahTeaPot.ScaleUniform(&vec3.Zero, scale/utahTeaPot.Bounds.Ymax)

	porcelainMaterial := scn.NewMaterial().
		N("Porcelain material").
		C(color.Color{R: 0.88, G: 0.96, B: 0.96}).
		M(0.05, 0.1).
		T(0.0, true, scn.RefractionIndex_Porcelain)

	// glassMaterial := scn.NewMaterial().
	// 	N("Glass material").
	// 	C(color.Color{R: 0.95, G: 0.95, B: 0.97}, 1.0).
	// 	M(0.2, 0.05).
	// 	T(1.0, true, scn.RefractionIndex_Glass)

	utahTeaPot.ClearMaterials()
	utahTeaPot.Material = porcelainMaterial

	return utahTeaPot
}
