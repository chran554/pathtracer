package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
)

// NewSodaCanPepsi
// Normal soda can height is 11.6 cm.
func NewSodaCanPepsi(scale float64) *scn.FacetStructure {
	sodaCan := loadSodaCan("textures/sodacan/sodacan_pepsi.jpg", color.NewColor(0.2, 0.2, 0.9), scale)
	return sodaCan
}

// NewSodaCanCocaCola
// Normal soda can height is 11.6 cm.
func NewSodaCanCocaCola(scale float64) *scn.FacetStructure {
	sodaCan := loadSodaCan("textures/sodacan/sodacan_cocacola.jpg", color.NewColor(0.9, 0.2, 0.2), scale)
	return sodaCan
}

func loadSodaCan(textureFileName string, tabColor color.Color, scale float64) *scn.FacetStructure {
	var objFilename = "sodacan.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objFilename

	obj := ReadOrPanic(objFilenamePath)

	ymin := obj.Bounds.Ymin
	ymax := obj.Bounds.Ymax
	obj.Translate(&vec3.T{0.0, -ymin, 0.0})       // can bottom touch the ground (xz-plane)
	obj.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize/scale to height == 1.0 units

	aluminumMaterialLid := scn.NewMaterial().M(0.3, 0.4)
	aluminumMaterialTab := scn.NewMaterial().C(tabColor).M(0.3, 0.4)
	aluminumMaterialBody := scn.NewMaterial().
		M(0.2, 0.1).
		T(0.0, true, scn.RefractionIndex_AcrylicPlastic).
		CP(textureFileName, &vec3.T{0, 0.065, 0}, vec3.UnitX, vec3.T{0, 0.89 - 0.065, 0}, false)
	obj.ReplaceMaterial("lid", aluminumMaterialLid)
	obj.ReplaceMaterial("tab", aluminumMaterialTab)
	obj.ReplaceMaterial("body", aluminumMaterialBody)

	obj.ScaleUniform(&vec3.Zero, scale)

	obj.UpdateBounds()
	fmt.Printf("%s bounds: %+v\n", objFilename, obj.Bounds)

	return obj
}
