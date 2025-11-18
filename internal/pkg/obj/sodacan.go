package obj

import (
	"fmt"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

// NewSodaCanPepsi
// Normal soda can height is 11.6 cm.
func NewSodaCanPepsi(scale float64) *scn.FacetStructure {
	sodaCan := loadSodaCan(filepath.Join(TexturesDir, "sodacan/sodacan_pepsi.jpg"), color.NewColor(0.2, 0.2, 0.9), scale)
	return sodaCan
}

// NewSodaCanMtnDew
// Normal soda can height is 11.6 cm.
func NewSodaCanMtnDew(scale float64) *scn.FacetStructure {
	sodaCan := loadSodaCan(filepath.Join(TexturesDir, "sodacan/sodacan_mtn-dew.png"), color.NewColor(0.2, 0.9, 0.2), scale)
	return sodaCan
}

// NewSodaCanFantaOrange
// Normal soda can height is 11.6 cm.
func NewSodaCanFantaOrange(scale float64) *scn.FacetStructure {
	sodaCan := loadSodaCan(filepath.Join(TexturesDir, "sodacan/sodacan_fanta-orange.png"), color.NewColor(0.937, 0.455, 0.047), scale)
	return sodaCan
}

// NewSodaCanCocaColaClassic
// Normal soda can height is 11.6 cm.
func NewSodaCanCocaColaClassic(scale float64) *scn.FacetStructure {
	sodaCan := loadSodaCan(filepath.Join(TexturesDir, "sodacan/sodacan_cocacola.jpg"), color.NewColor(0.9, 0.2, 0.2), scale)
	return sodaCan
}

func NewSodaCanCocaColaModern(scale float64) *scn.FacetStructure {
	sodaCan := loadSodaCan(filepath.Join(TexturesDir, "sodacan/sodacan_cocacola_02.png"), color.NewColor(0.9, 0.2, 0.2), scale)
	return sodaCan
}

func NewSodaCanTest(scale float64) *scn.FacetStructure {
	sodaCan := loadSodaCan(filepath.Join(TexturesDir, "test/checkered 360x180 with lines.png"), color.NewColor(0.65, 0.55, 0.2), scale)
	return sodaCan
}

func loadSodaCan(textureFileName string, tabColor color.Color, scale float64) *scn.FacetStructure {
	sodaCan := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "sodacan.obj"))
	sodaCan.CenterOn(&vec3.Zero)

	ymin := sodaCan.Bounds.Ymin
	ymax := sodaCan.Bounds.Ymax
	sodaCan.Translate(&vec3.T{0.0, -ymin, 0.0})       // can bottom touch the ground (xz-plane)
	sodaCan.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize/scale to height == 1.0 units

	sodaCanBody := sodaCan.GetFirstObjectBySubstructureName("body")
	bodyProjectionBottomOffset := 0.065

	aluminumMaterialLid := scn.NewMaterial().N("lid").M(0.3, 0.4)
	aluminumMaterialTab := scn.NewMaterial().N("tab").C(tabColor).M(0.3, 0.4)
	aluminumMaterialBody := scn.NewMaterial().N("body").
		M(0.2, 0.15).
		T(0.0, false, scn.RefractionIndex_AcrylicPlastic).
		CP(floatimage.Load(textureFileName), &vec3.T{0, bodyProjectionBottomOffset, 0}, vec3.UnitX, vec3.T{0, sodaCanBody.Bounds.Ymax - bodyProjectionBottomOffset - 0.005, 0}, false)
	//CP(textureFileName, &vec3.T{0, 0.065, 0}, vec3.UnitX, vec3.T{0, 0.89 - 0.065, 0}, false)

	sodaCan.ReplaceMaterial("lid", aluminumMaterialLid)
	sodaCan.ReplaceMaterial("tab", aluminumMaterialTab)
	sodaCan.ReplaceMaterial("body", aluminumMaterialBody)

	sodaCan.ScaleUniform(&vec3.Zero, scale)

	fmt.Printf("Sodacan %s bounds: %+v\n", textureFileName, sodaCan.Bounds)

	return sodaCan
}
