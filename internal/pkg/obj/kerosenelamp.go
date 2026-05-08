package obj

import (
	"fmt"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/fileformat/wavefront"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

func NewKeroseneLamp(scale float64, emission float64, sootness float64) *scene.FacetStructure {
	flameTexture := floatimage.LoadOrPanic("textures/misc/kerosenelamp/kerosenelamp_flame_wave.png")
	sootSmudgeTexture := floatimage.LoadOrPanic("textures/misc/kerosenelamp/kerosenelamp_glass_wave_mod2.png")

	keroseneLamp := loadKeroseneLamp(scale)

	flame := keroseneLamp.GetFirstObjectByMaterialName("flame")
	flameCenterBounds := flame.Bounds.Center()

	glass := keroseneLamp.GetFirstObjectByMaterialName("glass")
	glassCenterBounds := glass.Bounds.Center()

	brassMaterial := scene.NewMaterial().N("brass").
		C(color.NewColor(0.8, 0.60, 0.25).Multiply(0.3)).
		M(0.8, 0.3)
	brassMaterial.ColorizeReflection = true

	flameMaterial := scene.NewMaterial().N("flame").
		C(color.White).
		T(1.0, false, scene.RefractionIndex_Air).
		E(color.White, emission, false).
		CP(flameTexture, &vec3.T{flameCenterBounds[0], flame.Bounds.Ymin, flameCenterBounds[2]}, vec3.UnitZ, (vec3.UnitY).Scaled(flame.Bounds.SizeY()*0.95), false)

	glassMaterial := scene.NewMaterial().N("glass").
		//C(color.NewColor(0.93, 0.94, 0.95)).
		C(color.NewColor(0.85, 0.87, 0.90)).
		T(0.90, false, scene.RefractionIndex_Glass).
		M(0.085, 0.015)

	smudgedGlassMaterial := scene.NewMaterial().N("smudged_glass").
		C(color.White).
		T(1.0-sootness, false, scene.RefractionIndex_Air).
		M(0.0, 1.0).
		CP(sootSmudgeTexture, &vec3.T{glassCenterBounds[0], glass.Bounds.Ymin, glassCenterBounds[2]}, vec3.UnitX, (vec3.UnitY).Scaled(glass.Bounds.SizeY()), false)

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

func loadKeroseneLamp(scale float64) *scene.FacetStructure {
	keroseneLamp := wavefront.ReadFacetStructureOrPanic(filepath.Join(ObjFileDir, "kerosene_lamp.obj"))

	ymin := keroseneLamp.Bounds.Ymin
	ymax := keroseneLamp.Bounds.Ymax
	keroseneLamp.Translate(&vec3.T{0.0, -ymin, 0.0})       // lamp base touch the ground (xz-plane)
	keroseneLamp.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize/scale to height == 1.0 units

	keroseneLamp.ScaleUniform(&vec3.Zero, scale)

	fmt.Printf("Kerosene lamp bounds: %+v\n", keroseneLamp.Bounds)

	return keroseneLamp
}
