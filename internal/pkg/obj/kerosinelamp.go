package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"os"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
)

func NewKerosineLamp(kerosineScale *vec3.T) *scn.FacetStructure {
	kerosineLamp := loadKerosineLamp(kerosineScale)
	brassMaterial := scn.NewMaterial().N("brass").C(color.Color{R: 0.8, G: 0.7, B: 0.15}, 1.0).M(0.8, 0.3)
	kerosineLamp.GetFirstObjectByName("base").Material = brassMaterial
	kerosineLamp.GetFirstObjectByName("handle").Material = brassMaterial
	kerosineLamp.GetFirstObjectByName("knob").Material = brassMaterial
	kerosineLamp.GetFirstObjectByName("wick_holder").Material = brassMaterial
	kerosineLamp.GetFirstObjectByName("flame").Material = scn.NewMaterial().
		C(color.Color{R: 1.0, G: 0.9, B: 0.7}, 1.0).
		E(color.Color{R: 1.0, G: 0.9, B: 0.7}, 70.0, true)
	kerosineLamp.GetFirstObjectByName("glass").Material = scn.NewMaterial().
		C(color.Color{R: 0.93, G: 0.93, B: 0.93}, 1.0).
		T(0.8, false, 0.0).
		M(0.95, 0.1)
	return kerosineLamp
}

func loadKerosineLamp(scale *vec3.T) *scn.FacetStructure {
	var objFilename = "kerosine_lamp.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objFilename

	objFile, err := os.Open(objFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", objFilenamePath, err.Error())
	}
	defer objFile.Close()

	obj, err := Read(objFile)
	ymin := obj.Bounds.Ymin
	ymax := obj.Bounds.Ymax
	obj.Translate(&vec3.T{0.0, -ymin, 0.0})       // lamp base touch the ground (xz-plane)
	obj.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize/scale to height == 1.0 units

	obj.Scale(&vec3.Zero, scale)

	obj.UpdateBounds()
	fmt.Printf("Kerosine lamp bounds: %+v\n", obj.Bounds)

	return obj
}
