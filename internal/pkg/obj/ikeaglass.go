package obj

import (
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/fileformat/wavefront"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

func NewGlassIkeaPokal(scale float64) *scene.FacetStructure {
	glass := loadIkeaGlassPokal()

	glass.CenterOn(&vec3.Zero)
	glass.Translate(&vec3.T{0, -glass.Bounds.Ymin, 0})

	glass.ScaleUniform(&vec3.Zero, scale/glass.Bounds.Ymax)

	glass.Material = createGlassMaterial(scale)

	return glass
}

func NewGlassIkeaSkoja(scale float64, includeLiquid bool) *scene.FacetStructure {
	glass := loadIkeaGlassSkoja()

	glass.CenterOn(&vec3.Zero)
	glass.Translate(&vec3.T{0, -glass.Bounds.Ymin, 0})

	glass.ScaleUniform(&vec3.Zero, scale/glass.Bounds.Ymax)

	liquidObject := glass.GetFirstObjectBySubstructureName("liquid")
	glassObject := glass.GetFirstObjectBySubstructureName("glass")

	glassMaterial := createGlassMaterial(scale / 2)

	liquidMaterial := scene.NewMaterial().
		N("red juice").
		C(color.White).
		M(0.0, 0.0).
		T(0.98, true, scene.RefractionIndex_SugarSolution60).
		Attenuate(color.NewColor(0.90, 0.35, 0.37), scale/3)

	glassObject.Material = glassMaterial
	liquidObject.Material = liquidMaterial

	if !includeLiquid {
		glass.RemoveObjectsBySubstructureName("liquid")
	}

	return glass
}

func createGlassMaterial(scale float64) *scene.Material {
	return scene.NewMaterial().
		N("glass").
		C(color.NewColor(0.95, 0.95, 0.97)).
		M(0.0, 0.0).
		T(0.98, true, scene.RefractionIndex_Glass).
		Attenuate(color.NewColor(0.90, 0.95, 0.97), scale)
}

func loadIkeaGlassPokal() *scene.FacetStructure {
	glass := wavefront.ReadFacetStructureOrPanic(filepath.Join(ObjFileDir, "glass_ikea_pokal.obj"))
	return glass
}

func loadIkeaGlassSkoja() *scene.FacetStructure {
	glass := wavefront.ReadFacetStructureOrPanic(filepath.Join(ObjFileDir, "glass_ikea_skoja2.obj"))
	return glass
}
