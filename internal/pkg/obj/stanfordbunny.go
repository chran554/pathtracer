package obj

import (
	"github.com/ungerik/go3d/float64/vec3"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"
)

func NewStanfordBunny(scale float64) *scn.FacetStructure {
	bunny := loadStanfordBunny()

	bunny.Scale(&vec3.Zero, &vec3.T{-1, 1, 1})
	bunny.RotateX(&vec3.Zero, util.DegToRad(90))
	bunny.CenterOn(&vec3.Zero)
	bunny.Translate(&vec3.T{0, -bunny.Bounds.Ymin, 0})

	bunny.ScaleUniform(&vec3.Zero, scale/bunny.Bounds.Ymax)

	skinMaterial := scn.NewMaterial().N("stanfordbunny").
		C(color.NewColorHex("#A1887F")).
		M(0.1, 0.75)

	bunny.ReplaceMaterial("stanfordbunny", skinMaterial)

	return bunny
}

func loadStanfordBunny() *scn.FacetStructure {
	return wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "stanfordbunny.obj"))
}
