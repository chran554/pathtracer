package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj2"
	scn "pathtracer/internal/pkg/scene"
)

func NewLamppost(scale float64, emission float64) *scn.FacetStructure {
	lamppost := loadLamppost(scale)
	lamppost.ClearMaterials()

	lamppostMaterial := scn.NewMaterial().N("lamppost").C(color.NewColor(0.20, 0.10, 0.08)).M(0.2, 0.3)

	lampMaterial0 := scn.NewMaterial().N("lamp_0").C(color.White).E(color.White, emission, true)
	lampMaterial1 := scn.NewMaterial().N("lamp_1").C(color.White).E(color.White, emission, true)
	lampMaterial2 := scn.NewMaterial().N("lamp_2").C(color.White).E(color.White, emission, true)
	lampMaterial3 := scn.NewMaterial().N("lamp_3").C(color.White).E(color.White, emission, true)

	lamppost.Material = lamppostMaterial

	lamppost.GetFirstObjectBySubstructureName("lamp_0").Material = lampMaterial0
	lamppost.GetFirstObjectBySubstructureName("lamp_1").Material = lampMaterial1
	lamppost.GetFirstObjectBySubstructureName("lamp_2").Material = lampMaterial2
	lamppost.GetFirstObjectBySubstructureName("lamp_3").Material = lampMaterial3

	return lamppost
}

func loadLamppost(scale float64) *scn.FacetStructure {
	lamppost := wavefrontobj2.ReadOrPanic(filepath.Join(ObjFileDir, "lamppost.obj"))

	ymin := lamppost.Bounds.Ymin
	ymax := lamppost.Bounds.Ymax
	lamppost.Translate(&vec3.T{0.0, -ymin, 0.0})       // lamp post base touch the ground (xz-plane)
	lamppost.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize to height == 1.0

	lamppost.ScaleUniform(&vec3.Zero, scale)

	fmt.Printf("Lamp post bounds: %+v\n", lamppost.Bounds)

	return lamppost
}
