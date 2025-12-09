package obj

import (
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

func NewGlassIkeaPokal(scale float64) *scene.FacetStructure {
	glass := loadIkeaGlassPokal()

	glass.CenterOn(&vec3.Zero)
	glass.Translate(&vec3.T{0, -glass.Bounds.Ymin, 0})

	glass.ScaleUniform(&vec3.Zero, scale/glass.Bounds.Ymax)

	glass.Material = scene.NewMaterial().
		N("glass").
		C(color.NewColor(0.90, 0.92, 0.95)).
		M(0.270, 0.030).
		T(0.700, true, scene.RefractionIndex_Glass)

	return glass
}

func NewGlassIkeaSkoja(scale float64, includeLiquid bool) *scene.FacetStructure {
	glass := loadIkeaGlassSkoja()

	glass.CenterOn(&vec3.Zero)
	glass.Translate(&vec3.T{0, -glass.Bounds.Ymin, 0})

	glass.ScaleUniform(&vec3.Zero, scale/glass.Bounds.Ymax)

	liquidObject := glass.GetFirstObjectBySubstructureName("liquid")
	glassObject := glass.GetFirstObjectBySubstructureName("glass")

	glassMaterial := scene.NewMaterial().
		N("glass").
		C(color.NewColor(0.95, 0.95, 0.97)).
		M(0.1, 0.05).
		T(0.98, true, scene.RefractionIndex_Glass)

	liquidMaterial := scene.NewMaterial().
		N("red juice").
		C(color.NewColor(0.97, 0.45, 0.47)).
		M(0.2, 0.0).
		T(0.98, true, scene.RefractionIndex_SugarSolution60)

	glassObject.Material = glassMaterial
	liquidObject.Material = liquidMaterial

	if !includeLiquid {
		glass.RemoveObjectsBySubstructureName("liquid")
	}

	return glass
}

func loadIkeaGlassPokal() *scene.FacetStructure {
	glass := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "glass_ikea_pokal.obj"))
	return glass
}

func loadIkeaGlassSkoja() *scene.FacetStructure {
	glass := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "glass_ikea_skoja.obj"))
	return glass
}
