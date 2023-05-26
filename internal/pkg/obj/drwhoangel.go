package obj

import (
	"github.com/ungerik/go3d/float64/vec3"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"
)

func NewDrWhoAngel(scale float64, includeAngel bool, includePillar bool) *scn.FacetStructure {
	angel := loadDrWhoAngel()

	if !includeAngel {
		angel.RemoveObjectsBySubstructureName("drwho_angel")
	}
	if !includePillar {
		angel.RemoveObjectsBySubstructureName("drwho_angel::pillar")
	}

	angel.RotateX(&vec3.Zero, util.DegToRad(90))
	angel.RotateY(&vec3.Zero, util.DegToRad(180))
	angel.CenterOn(&vec3.Zero)
	angel.Translate(&vec3.T{0, -angel.Bounds.Ymin, 0})

	angel.ScaleUniform(&vec3.Zero, scale/angel.Bounds.Ymax)

	statueMaterial := scn.NewMaterial().N("angel").
		C(color.NewColor(0.9, 0.9, 0.9)).
		M(0.3, 0.6)

	pillarMaterial := scn.NewMaterial().N("pillar").
		C(color.NewColor(0.8, 0.8, 0.8)).
		M(0.1, 0.8)

	angel.ReplaceMaterial("angel", statueMaterial)
	angel.ReplaceMaterial("pillar", pillarMaterial)

	return angel
}

func loadDrWhoAngel() *scn.FacetStructure {
	angel := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "drwhoangel.obj"))
	return angel
}
