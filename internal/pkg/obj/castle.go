package obj

import (
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"
)

func NewCastle(scale float64, lightColor color.Color, lightEmission float64) *scn.FacetStructure {
	object := loadCastle(scale)
	object.PurgeEmptySubStructures()

	stainedGlassMaterial := scn.NewMaterial().N("stained_glass").
		C(color.NewColor(0.95, 0.90, 0.60)).
		T(0.8, false, 0.0).
		M(0.2, 0.3)
	stainedGlassObjects := object.GetObjectsByMaterialName("colored_glass")
	for _, stainedGlassObject := range stainedGlassObjects {
		stainedGlassObject.Material = stainedGlassMaterial
	}

	roofMaterial := scn.NewMaterial().N("roof").C(color.NewColor(0.30, 0.25, 0.10))
	roofObjects := object.GetObjectsByMaterialName("erroded_cupper")
	for _, roofObject := range roofObjects {
		roofObject.Material = roofMaterial
	}

	glassMaterial := scn.NewMaterial().N("glass").
		C(color.NewColor(0.93, 0.93, 0.95)).
		T(0.8, false, 0.0).
		M(0.2, 0.1)
	glassObjects := object.GetObjectsByMaterialName("glass")
	for _, glassObject := range glassObjects {
		glassObject.Material = glassMaterial
	}

	lightMaterial := scn.NewMaterial().N("light").E(lightColor, lightEmission, true)
	object.ReplaceMaterial("chapel_light", lightMaterial)
	object.ReplaceMaterial("hall_light", lightMaterial)
	object.ReplaceMaterial("hall_tower_left_light", lightMaterial)
	object.ReplaceMaterial("hall_tower_right_light", lightMaterial)
	object.ReplaceMaterial("house_tower_light", lightMaterial)
	object.ReplaceMaterial("tower_back_right_light", lightMaterial)
	object.ReplaceMaterial("tower_middle_short_light", lightMaterial)
	object.ReplaceMaterial("tower_middle_tall_light", lightMaterial)

	object.RotateY(&vec3.Zero, math.Pi)

	return object
}

func loadCastle(scale float64) *scn.FacetStructure {
	castle := wavefrontobj.ReadOrPanic(filepath.Join(ObjEvaluationFileDir, "castle_03.obj"))
	// castle.Scale(&vec3.Zero, &vec3.T{-1, 1, 1}) // Flip along x-axis
	castle.CenterOn(&vec3.Zero)

	ymin := castle.Bounds.Ymin
	ymax := castle.Bounds.Ymax
	castle.Translate(&vec3.T{0.0, -ymin, 0.0})       // lamp base touch the ground (xz-plane)
	castle.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize/scale to height == 1.0 units
	castle.ScaleUniform(&vec3.Zero, scale)           // resize requested size

	return castle
}
