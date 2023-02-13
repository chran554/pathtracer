package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"os"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
)

func NewCastle(scale *vec3.T) *scn.FacetStructure {
	object := loadCastle(scale)
	object.Purge()

	stainedGlassMaterial := scn.NewMaterial().N("stained_glass").
		C(color.Color{R: 0.95, G: 0.90, B: 0.60}, 1.0).
		T(0.8, false, 0.0).
		M(0.95, 0.3)
	stainedGlassObjects := object.GetObjectsByMaterialName("colored_glass")
	for _, stainedGlassObject := range stainedGlassObjects {
		stainedGlassObject.Material = stainedGlassMaterial
	}

	roofMaterial := scn.NewMaterial().N("roof").
		C(color.Color{R: 0.90, G: 0.80, B: 0.50}, 0.3)
	roofObjects := object.GetObjectsByMaterialName("erroded_cupper")
	for _, roofObject := range roofObjects {
		roofObject.Material = roofMaterial
	}

	glassMaterial := scn.NewMaterial().N("glass").
		C(color.Color{R: 0.93, G: 0.93, B: 0.95}, 1.0).
		T(0.8, false, 0.0).
		M(0.95, 0.1)
	glassObjects := object.GetObjectsByMaterialName("glass")
	for _, glassObject := range glassObjects {
		glassObject.Material = glassMaterial
	}

	return object
}

func loadCastle(scale *vec3.T) *scn.FacetStructure {
	var objFilename = "castle_02.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/" + objFilename

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
	fmt.Printf("Castle bounds: %+v\n", obj.Bounds)

	return obj
}
