package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"os"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
)

func NewSodaCanPepsi(scale float64) *scn.FacetStructure {
	sodaCan := loadSodaCan("textures/sodacan_pepsi.jpg", color.NewColor(0.2, 0.2, 0.9), scale)
	return sodaCan
}

func NewSodaCanCocaCola(scale float64) *scn.FacetStructure {
	sodaCan := loadSodaCan("textures/sodacan_cocacola.jpg", color.NewColor(0.9, 0.2, 0.2), scale)
	return sodaCan
}

func loadSodaCan(textureFileName string, tabColor color.Color, scale float64) *scn.FacetStructure {
	var objFilename = "sodacan.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objFilename

	objFile, err := os.Open(objFilenamePath)
	if err != nil {
		message := fmt.Sprintf("ouupps, something went wrong loading file: '%s'\n%s\n", objFilenamePath, err.Error())
		panic(message)
	}
	defer objFile.Close()

	obj, err := Read(objFile)
	ymin := obj.Bounds.Ymin
	ymax := obj.Bounds.Ymax
	obj.Translate(&vec3.T{0.0, -ymin, 0.0})       // can bottom touch the ground (xz-plane)
	obj.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize/scale to height == 1.0 units

	aluminumMaterialLid := scn.NewMaterial().M(0.3, 0.4)
	aluminumMaterialTab := scn.NewMaterial().C(tabColor).M(0.3, 0.4)
	aluminumMaterialBody := scn.NewMaterial().M(0.1, 0.1).CP(textureFileName, &vec3.T{0, 0.065, 0}, vec3.UnitX, vec3.T{0, 0.89 - 0.065, 0}, false)
	obj.ReplaceMaterial("lid", aluminumMaterialLid)
	obj.ReplaceMaterial("tab", aluminumMaterialTab)
	obj.ReplaceMaterial("body", aluminumMaterialBody)

	obj.ScaleUniform(&vec3.Zero, scale)

	obj.UpdateBounds()
	fmt.Printf("%s bounds: %+v\n", objFilename, obj.Bounds)

	return obj
}
