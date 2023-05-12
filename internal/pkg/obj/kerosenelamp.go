package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj2"
	scn "pathtracer/internal/pkg/scene"
)

func NewKeroseneLamp(scale float64, emission float64) *scn.FacetStructure {
	keroseneLamp := loadKeroseneLamp(scale)

	brassMaterial := scn.NewMaterial().N("brass").
		C(color.NewColor(0.8/2, 0.60/2, 0.25/2)).
		M(0.8, 0.3)
	flameMaterial := scn.NewMaterial().N("flame").
		C(color.Color{R: 1.0, G: 0.9, B: 0.6}).
		E(color.Color{R: 1.0, G: 0.9, B: 0.6}, emission, true)
	var glassColorFactor = 0.75
	darkGlassMaterial := scn.NewMaterial().N("smudged_glass").
		C(color.NewColor(0.93*glassColorFactor, 0.94*glassColorFactor, 0.95*glassColorFactor)).
		T(0.96, false, scn.RefractionIndex_Glass).
		M(0.01, 0.05)

	keroseneLamp.GetFirstObjectByMaterialName("base").Material = brassMaterial
	keroseneLamp.GetFirstObjectByMaterialName("handle").Material = brassMaterial
	keroseneLamp.GetFirstObjectByMaterialName("knob").Material = brassMaterial
	keroseneLamp.GetFirstObjectByMaterialName("wick_holder").Material = brassMaterial
	keroseneLamp.GetFirstObjectByMaterialName("flame").Material = flameMaterial
	keroseneLamp.GetFirstObjectByMaterialName("glass").Material = darkGlassMaterial

	if emission == 0.0 {
		keroseneLamp.RemoveObjectsByMaterialName("flame")
	}

	return keroseneLamp
}

func loadKeroseneLamp(scale float64) *scn.FacetStructure {
	keroseneLamp := wavefrontobj2.ReadOrPanic(filepath.Join(ObjFileDir, "kerosene_lamp.obj"))

	ymin := keroseneLamp.Bounds.Ymin
	ymax := keroseneLamp.Bounds.Ymax
	keroseneLamp.Translate(&vec3.T{0.0, -ymin, 0.0})       // lamp base touch the ground (xz-plane)
	keroseneLamp.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize/scale to height == 1.0 units

	keroseneLamp.ScaleUniform(&vec3.Zero, scale)

	fmt.Printf("Kerosene lamp bounds: %+v\n", keroseneLamp.Bounds)

	return keroseneLamp
}
