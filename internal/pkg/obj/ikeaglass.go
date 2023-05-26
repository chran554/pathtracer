package obj

import (
	"github.com/ungerik/go3d/float64/vec3"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"
)

func NewGlassIkeaPokal(scale float64) *scn.FacetStructure {
	glass := loadIkeaGlassPokal()

	glass.CenterOn(&vec3.Zero)
	glass.Translate(&vec3.T{0, -glass.Bounds.Ymin, 0})

	glass.ScaleUniform(&vec3.Zero, scale/glass.Bounds.Ymax)

	glass.Material = scn.NewMaterial().
		N("glass").
		C(color.NewColor(0.95, 0.95, 0.97)).
		M(0.1, 0.05).
		T(0.98, true, scn.RefractionIndex_Glass)

	return glass
}

func NewGlassIkeaSkoja(scale float64, includeLiquid bool) *scn.FacetStructure {
	glass := loadIkeaGlassSkoja()

	glass.CenterOn(&vec3.Zero)
	glass.Translate(&vec3.T{0, -glass.Bounds.Ymin, 0})

	glass.ScaleUniform(&vec3.Zero, scale/glass.Bounds.Ymax)

	liquidObject := glass.GetFirstObjectBySubstructureName("liquid")
	glassObject := glass.GetFirstObjectBySubstructureName("glass")

	glassMaterial := scn.NewMaterial().
		N("glass").
		C(color.NewColor(0.95, 0.95, 0.97)).
		M(0.1, 0.05).
		T(0.98, true, scn.RefractionIndex_Glass)

	liquidMaterial := scn.NewMaterial().
		N("red juice").
		C(color.NewColor(0.97, 0.45, 0.47)).
		M(0.2, 0.0).
		T(0.98, true, scn.RefractionIndex_SugarSolution60)

	glassObject.Material = glassMaterial
	liquidObject.Material = liquidMaterial

	if !includeLiquid {
		glass.RemoveObjectsBySubstructureName("liquid")
	}

	return glass
}

func loadIkeaGlassPokal() *scn.FacetStructure {
	glass := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "glass_ikea_pokal.obj"))
	return glass
}

func loadIkeaGlassSkoja() *scn.FacetStructure {
	glass := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "glass_ikea_skoja.obj"))
	return glass
}
