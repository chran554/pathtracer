package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
)

// NewCornellBox creates a new cornell box (open in the back) with the center of the floor in origin (0,0,0).
// Left wall is blue and right wall is red.
// The scale is the total (width, height, depth) of the cornell box.
func NewCornellBox(scale *vec3.T, lightIntensityFactor float64) *scn.FacetStructure {
	blueColor := color.NewColor(0.05, 0.05, 0.95)
	redColor := color.NewColor(0.95, 0.05, 0.05)

	cornellBox := cornelBox(scale, lightIntensityFactor)
	cornellBox.GetFirstMaterialByName("left").C(blueColor)
	cornellBox.GetFirstMaterialByName("right").C(redColor)

	return cornellBox
}

// NewWhiteCornellBox creates a new, whole white, cornell box (open in the back) with the center of the floor in origin (0,0,0).
// The scale is the total (width, height, depth) of the cornell box.
func NewWhiteCornellBox(scale *vec3.T, lightIntensityFactor float64) *scn.FacetStructure {
	return cornelBox(scale, lightIntensityFactor)
}

func cornelBox(scale *vec3.T, lightIntensityFactor float64) (cornellBox *scn.FacetStructure) {
	var cornellBoxFilename = "cornellbox.obj"
	var cornellBoxFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + cornellBoxFilename

	cornellBox = ReadOrPanic(cornellBoxFilenamePath)
	cornellBox.Name = "cornellbox"

	cornellBox.CenterOn(&vec3.Zero)
	cornellBox.Scale(&vec3.Zero, &vec3.T{1 / cornellBox.Bounds.Xmax, 1 / cornellBox.Bounds.Ymax, 1 / cornellBox.Bounds.Zmax})
	cornellBox.ScaleUniform(&vec3.Zero, 0.5)
	cornellBox.Translate(&vec3.T{0, -cornellBox.Bounds.Ymin, 0})
	cornellBox.Scale(&vec3.Zero, scale)

	fmt.Printf("Cornell box bounds: %+v\n", cornellBox.Bounds)

	cornellBox.ClearMaterials()

	cornellBox.Material = scn.NewMaterial().N("cornellbox").C(color.NewColorGrey(0.95))

	cornellBox.GetFirstObjectByName("Right").Material = cornellBox.Material.Copy().N("right")
	cornellBox.GetFirstObjectByName("Left").Material = cornellBox.Material.Copy().N("left")
	cornellBox.GetFirstObjectByName("Back").Material = cornellBox.Material.Copy().N("back")
	cornellBox.GetFirstObjectByName("Floor").Material = cornellBox.Material.Copy().N("floor")
	cornellBox.GetFirstObjectByName("Ceiling").Material = cornellBox.Material.Copy().N("ceiling")

	lampMaterial := scn.NewMaterial().N("lamp").E(color.White, lightIntensityFactor, true)
	cornellBox.GetFirstObjectByName("Lamp_1").Material = lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_2").Material = lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_3").Material = lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_4").Material = lampMaterial

	return cornellBox
}
