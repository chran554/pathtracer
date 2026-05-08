package obj

import (
	"fmt"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/fileformat/wavefront"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

func NewLamppost(scale float64, emission float64) *scene.FacetStructure {
	lamppost := loadLamppost(scale)
	lamppost.ClearMaterials()

	lamppostMaterial := scene.NewMaterial().N("lamppost").
		C(color.NewColor(0.20, 0.10, 0.08)).
		M(0.2, 0.3)

	warmWhiteColor := color.KelvinTemperatureColor2(5000)
	lampMaterial0 := scene.NewMaterial().N("lamp_0").C(color.White).E(warmWhiteColor, emission, true)
	lampMaterial1 := scene.NewMaterial().N("lamp_1").C(color.White).E(warmWhiteColor, emission, true)
	lampMaterial2 := scene.NewMaterial().N("lamp_2").C(color.White).E(warmWhiteColor, emission, true)
	lampMaterial3 := scene.NewMaterial().N("lamp_3").C(color.White).E(warmWhiteColor, emission, true)

	lamppost.Material = lamppostMaterial

	lamppost.GetFirstObjectBySubstructureName("lamp_0").Material = lampMaterial0
	lamppost.GetFirstObjectBySubstructureName("lamp_1").Material = lampMaterial1
	lamppost.GetFirstObjectBySubstructureName("lamp_2").Material = lampMaterial2
	lamppost.GetFirstObjectBySubstructureName("lamp_3").Material = lampMaterial3

	return lamppost
}

func loadLamppost(scale float64) *scene.FacetStructure {
	lamppost := wavefront.ReadFacetStructureOrPanic(filepath.Join(ObjFileDir, "lamppost.obj"))

	ymin := lamppost.Bounds.Ymin
	ymax := lamppost.Bounds.Ymax
	lamppost.Translate(&vec3.T{0.0, -ymin, 0.0})       // lamp post base touch the ground (xz-plane)
	lamppost.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize to height == 1.0

	lamppost.ScaleUniform(&vec3.Zero, scale)

	fmt.Printf("Lamp post bounds: %+v\n", lamppost.Bounds)

	return lamppost
}
