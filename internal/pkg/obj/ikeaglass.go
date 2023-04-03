package obj

import (
	"github.com/ungerik/go3d/float64/vec3"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"
)

func NewGlassIkeaPokal(scale float64) *scn.FacetStructure {
	glass := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "glass_ikea_pokal.obj"))

	glass.CenterOn(&vec3.Zero)
	glass.Translate(&vec3.T{0, -glass.Bounds.Ymin, 0})

	glass.ScaleUniform(&vec3.Zero, scale/glass.Bounds.Ymax)
	glass.ClearMaterials()

	glass.Material = scn.NewMaterial().
		N("Glass IKEA Pokal").
		C(color.Color{R: 0.95, G: 0.95, B: 0.97}).
		M(0.1, 0.05).
		T(0.98, true, scn.RefractionIndex_Glass)

	return glass
}

func NewGlassIkeaSkoja(scale float64) *scn.FacetStructure {
	glass := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "glass_ikea_skoja.obj"))

	glass.CenterOn(&vec3.Zero)
	glass.Translate(&vec3.T{0, -glass.Bounds.Ymin, 0})

	glass.ScaleUniform(&vec3.Zero, scale/glass.Bounds.Ymax)

	liquidObject := glass.GetFirstObjectByName("liquid")
	glassObject := glass.GetFirstObjectByName("glass")

	glassMaterial := scn.NewMaterial().
		N("Glass IKEA Skoja").
		C(color.Color{R: 0.95, G: 0.95, B: 0.97}).
		M(0.1, 0.05).
		T(0.98, true, scn.RefractionIndex_Glass)

	liquidMaterial := scn.NewMaterial().
		N("Red juice").
		C(color.Color{R: 0.97, G: 0.45, B: 0.47}).
		M(0.2, 0.0).
		T(0.98, true, scn.RefractionIndex_SugarSolution60)

	glass.ClearMaterials()
	glassObject.Material = glassMaterial
	liquidObject.Material = liquidMaterial

	return glass
}
