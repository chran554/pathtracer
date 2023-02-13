package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"os"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
)

func NewGlassIkeaPokal(scale float64) *scn.FacetStructure {
	var objectFilename = "glass_ikea_pokal.obj"
	var objectFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objectFilename

	objectFile, err := os.Open(objectFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", objectFilenamePath, err.Error())
	}
	defer objectFile.Close()

	object, err := Read(objectFile)
	object.CenterOn(&vec3.Zero)
	object.Translate(&vec3.T{0, -object.Bounds.Ymin, 0})

	object.ScaleUniform(&vec3.Zero, scale/object.Bounds.Ymax)
	object.UpdateBounds()
	object.ClearMaterials()

	object.Material = scn.NewMaterial().
		N("Glass IKEA Pokal").
		C(color.Color{R: 0.95, G: 0.95, B: 0.97}, 1.0).
		M(0.2, 0.05).
		T(1.0, true, scn.RefractionIndex_Glass)

	return object
}

func NewGlassIkeaSkoja(scale float64) *scn.FacetStructure {
	var objectFilename = "glass_ikea_skoja.obj"
	var objectFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objectFilename

	objectFile, err := os.Open(objectFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", objectFilenamePath, err.Error())
	}
	defer objectFile.Close()

	object, err := Read(objectFile)
	object.CenterOn(&vec3.Zero)
	object.Translate(&vec3.T{0, -object.Bounds.Ymin, 0})

	object.ScaleUniform(&vec3.Zero, scale/object.Bounds.Ymax)
	object.UpdateBounds()

	liquidObject := object.GetFirstObjectByName("liquid")
	glassObject := object.GetFirstObjectByName("glass")

	glassMaterial := scn.NewMaterial().
		N("Glass IKEA Skoja").
		C(color.Color{R: 0.95, G: 0.95, B: 0.97}, 1.0).
		M(0.2, 0.05).
		T(1.0, true, scn.RefractionIndex_Glass)

	liquidMaterial := scn.NewMaterial().
		N("Red juice").
		C(color.Color{R: 0.97, G: 0.45, B: 0.47}, 1.0).
		M(0.2, 0.0).
		T(1.0, true, scn.RefractionIndex_Water)

	object.ClearMaterials()
	glassObject.Material = glassMaterial
	liquidObject.Material = liquidMaterial

	return object
}
