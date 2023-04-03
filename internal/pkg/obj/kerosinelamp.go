package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"
)

func NewKerosineLamp(kerosineScale *vec3.T) *scn.FacetStructure {
	kerosineLamp := loadKerosineLamp(kerosineScale)
	brassMaterial := scn.NewMaterial().N("brass").C(color.Color{R: 0.8, G: 0.7, B: 0.15}).M(0.8, 0.3)
	kerosineLamp.GetFirstObjectByName("base").Material = brassMaterial
	kerosineLamp.GetFirstObjectByName("handle").Material = brassMaterial
	kerosineLamp.GetFirstObjectByName("knob").Material = brassMaterial
	kerosineLamp.GetFirstObjectByName("wick_holder").Material = brassMaterial
	kerosineLamp.GetFirstObjectByName("flame").Material = scn.NewMaterial().
		C(color.Color{R: 1.0, G: 0.9, B: 0.7}).
		E(color.Color{R: 1.0, G: 0.9, B: 0.7}, 70.0, true)
	kerosineLamp.GetFirstObjectByName("glass").Material = scn.NewMaterial().
		C(color.Color{R: 0.93, G: 0.93, B: 0.93}).
		T(0.8, false, 0.0).
		M(0.95, 0.1)
	return kerosineLamp
}

func loadKerosineLamp(scale *vec3.T) *scn.FacetStructure {
	kerosineLamp := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "kerosine_lamp.obj"))

	ymin := kerosineLamp.Bounds.Ymin
	ymax := kerosineLamp.Bounds.Ymax
	kerosineLamp.Translate(&vec3.T{0.0, -ymin, 0.0})       // lamp base touch the ground (xz-plane)
	kerosineLamp.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize/scale to height == 1.0 units

	kerosineLamp.Scale(&vec3.Zero, scale)

	fmt.Printf("Kerosine lamp bounds: %+v\n", kerosineLamp.Bounds)

	return kerosineLamp
}
