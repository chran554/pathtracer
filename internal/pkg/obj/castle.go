package obj

import (
	"github.com/ungerik/go3d/float64/vec3"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"
)

func NewCastle(scale float64) *scn.FacetStructure {
	object := loadCastle(scale)
	object.PurgeEmptySubStructures()

	stainedGlassMaterial := scn.NewMaterial().N("stained_glass").
		C(color.Color{R: 0.95, G: 0.90, B: 0.60}).
		T(0.8, false, 0.0).
		M(0.95, 0.3)
	stainedGlassObjects := object.GetObjectsByMaterialName("colored_glass")
	for _, stainedGlassObject := range stainedGlassObjects {
		stainedGlassObject.Material = stainedGlassMaterial
	}

	roofMaterial := scn.NewMaterial().N("roof").C(color.Color{R: 0.30, G: 0.25, B: 0.10})
	roofObjects := object.GetObjectsByMaterialName("erroded_cupper")
	for _, roofObject := range roofObjects {
		roofObject.Material = roofMaterial
	}

	glassMaterial := scn.NewMaterial().N("glass").
		C(color.Color{R: 0.93, G: 0.93, B: 0.95}).
		T(0.8, false, 0.0).
		M(0.95, 0.1)
	glassObjects := object.GetObjectsByMaterialName("glass")
	for _, glassObject := range glassObjects {
		glassObject.Material = glassMaterial
	}

	return object
}

func loadCastle(scale float64) *scn.FacetStructure {
	castle := wavefrontobj.ReadOrPanic(filepath.Join(ObjEvaluationFileDir, "castle_02.obj"))

	ymin := castle.Bounds.Ymin
	ymax := castle.Bounds.Ymax
	castle.Translate(&vec3.T{0.0, -ymin, 0.0})       // lamp base touch the ground (xz-plane)
	castle.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize/scale to height == 1.0 units
	castle.ScaleUniform(&vec3.Zero, scale)           // resize requested size

	return castle
}
