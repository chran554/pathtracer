package obj

import (
	"github.com/ungerik/go3d/float64/vec3"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"
)

func NewDrWhoAngel(scale float64) *scn.FacetStructure {
	angel := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "drwho_angel.obj"))

	angel.CenterOn(&vec3.Zero)
	angel.Translate(&vec3.T{0, -angel.Bounds.Ymin, 0})
	angel.UpdateBounds()

	angel.ScaleUniform(&vec3.Zero, scale/angel.Bounds.Ymax)
	angel.UpdateBounds()

	statueMaterial := scn.NewMaterial().N("angel").
		C(color.Color{R: 0.9, G: 0.9, B: 0.9}).
		M(0.3, 0.6)

	pillarMaterial := scn.NewMaterial().N("pillar").
		C(color.Color{R: 0.8, G: 0.8, B: 0.8}).
		M(0.1, 0.8)

	angel.ReplaceMaterial("angel", statueMaterial)
	angel.ReplaceMaterial("pillar", pillarMaterial)

	return angel
}
