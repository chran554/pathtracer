package obj

import (
	"fmt"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

// NewCornellBox creates a new cornell box (open in the back) with the center of the floor in origin (0,0,0).
// Left wall is blue and right wall is red.
// The scale is the total (width, height, depth) of the cornell box.
func NewCornellBox(scale *vec3.T, singleLight bool, lightIntensityFactor float64) *scn.FacetStructure {
	blueColor := color.NewColor(0.05, 0.05, 0.95)
	redColor := color.NewColor(0.95, 0.05, 0.05)

	cornellBox := cornelBox(scale, singleLight, lightIntensityFactor)
	cornellBox.GetFirstMaterialByName("left").C(blueColor)
	cornellBox.GetFirstMaterialByName("right").C(redColor)

	return cornellBox
}

// NewWhiteCornellBox creates a new, whole white, cornell box (open in the back) with the center of the floor in origin (0,0,0).
// The scale is the total (width, height, depth) of the cornell box.
func NewWhiteCornellBox(scale *vec3.T, singleLight bool, lightIntensityFactor float64) *scn.FacetStructure {
	return cornelBox(scale, singleLight, lightIntensityFactor)
}

func cornelBox(scale *vec3.T, singleLight bool, lightIntensityFactor float64) (cornellBox *scn.FacetStructure) {
	var cornellBoxFilenamePath = filepath.Join(ObjFileDir, "cornellbox.obj")

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

	if singleLight {
		cornellBox.RemoveObjectsByName("Lamp_1")
		cornellBox.RemoveObjectsByName("Lamp_2")
		cornellBox.RemoveObjectsByName("Lamp_3")
		cornellBox.RemoveObjectsByName("Lamp_4")

		lampPercentageOfCeiling := 2.0 / 3.0 // Two thirds in width and depth (i.e. 0.666*0.666 = 44.4% of the ceiling)

		lampSizeX := (scale[0] / 2) * lampPercentageOfCeiling
		lampY := scale[1] - 0.001
		lampSizeZ := (scale[2] / 2) * lampPercentageOfCeiling

		lamp := &scn.FacetStructure{
			Name:     "Lamp",
			Material: lampMaterial,
			Facets: getFlatRectangleFacets(
				&vec3.T{-lampSizeX, lampY, +lampSizeZ},
				&vec3.T{+lampSizeX, lampY, +lampSizeZ},
				&vec3.T{+lampSizeX, lampY, -lampSizeZ},
				&vec3.T{-lampSizeX, lampY, -lampSizeZ},
			),
		}

		cornellBox.FacetStructures = append(cornellBox.FacetStructures, lamp)

	} else {
		cornellBox.GetFirstObjectByName("Lamp_1").Material = lampMaterial
		cornellBox.GetFirstObjectByName("Lamp_2").Material = lampMaterial
		cornellBox.GetFirstObjectByName("Lamp_3").Material = lampMaterial
		cornellBox.GetFirstObjectByName("Lamp_4").Material = lampMaterial
	}

	cornellBox.UpdateBounds()

	return cornellBox
}

func getFlatRectangleFacets(p1, p2, p3, p4 *vec3.T) []*scn.Facet {
	v1 := p2.Subed(p1)
	v2 := p3.Subed(p1)
	normal := vec3.Cross(&v1, &v2)
	normal.Normalize()

	return []*scn.Facet{
		{Vertices: []*vec3.T{p1, p2, p4}, Normal: &normal},
		{Vertices: []*vec3.T{p4, p2, p3}, Normal: &normal},
	}
}
