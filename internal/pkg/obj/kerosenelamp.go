package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"
)

func NewKeroseneLamp(scale float64, emission float64) *scn.FacetStructure {
	keroseneLamp := loadKeroseneLamp(scale)

	flame := keroseneLamp.GetFirstObjectByMaterialName("flame")
	flameCenterBounds := flame.Bounds.Center()

	glass := keroseneLamp.GetFirstObjectByMaterialName("glass")
	glassCenterBounds := glass.Bounds.Center()

	brassMaterial := scn.NewMaterial().N("brass").
		C(color.NewColor(0.8/2, 0.60/2, 0.25/2)).
		M(0.8, 0.3)
	flameMaterial := scn.NewMaterial().N("flame").
		C(color.White).
		E(color.White, emission, false).
		CP("textures/misc/kerosenelamp/kerosenelamp_flame_wave.png", &vec3.T{flameCenterBounds[0], flame.Bounds.Ymin, flameCenterBounds[2]}, vec3.UnitZ, (vec3.UnitY).Scaled(flame.Bounds.SizeY()), false)
	glassMaterial := scn.NewMaterial().N("glass").
		C(color.NewColor(0.93, 0.94, 0.95)).
		T(0.95, false, scn.RefractionIndex_Glass).
		M(0.05, 0.05)
	glassMaterial.Diffuse = 0.0

	smudgedGlassMaterial := scn.NewMaterial().N("smudged_glass").
		C(color.White).
		T(1.0, false, scn.RefractionIndex_Air).
		CP("textures/misc/kerosenelamp/kerosenelamp_glass_wave_mod2.png", &vec3.T{glassCenterBounds[0], glass.Bounds.Ymin, glassCenterBounds[2]}, vec3.UnitX, (vec3.UnitY).Scaled(glass.Bounds.SizeY()), false)

	keroseneLamp.GetFirstObjectByMaterialName("base").Material = brassMaterial
	keroseneLamp.GetFirstObjectByMaterialName("handle").Material = brassMaterial
	keroseneLamp.GetFirstObjectByMaterialName("knob").Material = brassMaterial
	keroseneLamp.GetFirstObjectByMaterialName("wick_holder").Material = brassMaterial
	keroseneLamp.GetFirstObjectByMaterialName("flame").Material = flameMaterial
	keroseneLamp.GetFirstObjectByMaterialName("glass").Material = glassMaterial

	innerGlassSmudge := loadKeroseneLamp(scale).GetFirstObjectByMaterialName("glass")
	innerGlassSmudge.Scale(innerGlassSmudge.Bounds.Center(), &vec3.T{0.999, 1, 0.999})
	innerGlassSmudge.Material = smudgedGlassMaterial
	keroseneLamp.FacetStructures = append(keroseneLamp.FacetStructures, innerGlassSmudge)

	if emission == 0.0 {
		keroseneLamp.RemoveObjectsByMaterialName("flame")
	}

	return keroseneLamp
}

func loadKeroseneLamp(scale float64) *scn.FacetStructure {
	keroseneLamp := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "kerosene_lamp.obj"))

	ymin := keroseneLamp.Bounds.Ymin
	ymax := keroseneLamp.Bounds.Ymax
	keroseneLamp.Translate(&vec3.T{0.0, -ymin, 0.0})       // lamp base touch the ground (xz-plane)
	keroseneLamp.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize/scale to height == 1.0 units

	keroseneLamp.ScaleUniform(&vec3.Zero, scale)

	fmt.Printf("Kerosene lamp bounds: %+v\n", keroseneLamp.Bounds)

	return keroseneLamp
}
