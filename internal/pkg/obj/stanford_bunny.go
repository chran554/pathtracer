package obj

import (
	"github.com/ungerik/go3d/float64/vec3"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"
)

func NewStanfordBunny(scale float64) *scn.FacetStructure {
	bunny := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "stanford_bunny.obj"))

	bunny.CenterOn(&vec3.Zero)
	bunny.Translate(&vec3.T{0, -bunny.Bounds.Ymin, 0})

	bunny.ScaleUniform(&vec3.Zero, scale/bunny.Bounds.Ymax)

	skinMaterial := scn.NewMaterial().N("stanford_bunny").
		C(color.Color{R: 0.9, G: 0.85, B: 0.6}).
		M(0.3, 0.6)

	bunny.ReplaceMaterial("stanford_bunny", skinMaterial)

	return bunny
}
