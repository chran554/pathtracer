package obj

import (
	"fmt"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

// NewCornellBox creates a new cornell box (open in the back) with the center of the floor in origin (0,0,0).
// Left wall is blue and right wall is red.
// The scale is the total (width, height, depth) of the cornell box.
func NewCornellBox(scale *vec3.T, singleLight bool, lightIntensityFactor float64) *scn.FacetStructure {
	blueColor := color.NewColor(0.05, 0.05, 0.95)
	redColor := color.NewColor(0.95, 0.05, 0.05)

	cornellBox := loadCornellBox(scale, singleLight, lightIntensityFactor)
	cornellBox.GetFirstMaterialByName("Left").C(blueColor)
	cornellBox.GetFirstMaterialByName("Right").C(redColor)

	return cornellBox
}

// NewWhiteCornellBox creates a new, whole white, cornell box (open in the back) with the center of the floor in origin (0,0,0).
// The scale is the total (width, height, depth) of the cornell box.
func NewWhiteCornellBox(scale *vec3.T, singleLight bool, lightIntensityFactor float64) *scn.FacetStructure {
	return loadCornellBox(scale, singleLight, lightIntensityFactor)
}

func loadCornellBox(scale *vec3.T, singleLight bool, lightIntensityFactor float64) (cornellBox *scn.FacetStructure) {
	cornellBox = wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "cornellbox.obj"))

	cornellBox.CenterOn(&vec3.Zero)
	cornellBox.Scale(&vec3.Zero, &vec3.T{1 / cornellBox.Bounds.Xmax, 1 / cornellBox.Bounds.Ymax, 1 / cornellBox.Bounds.Zmax})
	cornellBox.ScaleUniform(&vec3.Zero, 0.5)
	cornellBox.Translate(&vec3.T{0, -cornellBox.Bounds.Ymin, 0})
	cornellBox.Scale(&vec3.Zero, scale)

	fmt.Printf("Cornell box bounds: %+v\n", cornellBox.Bounds)

	boxMaterial := scn.NewMaterial().N("cornellbox").C(color.NewColorGrey(1.0))

	cornellBox.ReplaceMaterial("Right", boxMaterial.Copy().N("Right"))
	cornellBox.ReplaceMaterial("Left", boxMaterial.Copy().N("Left"))
	cornellBox.ReplaceMaterial("Back", boxMaterial.Copy().N("Back"))
	cornellBox.ReplaceMaterial("Floor", boxMaterial.Copy().N("Floor"))
	cornellBox.ReplaceMaterial("Ceiling", boxMaterial.Copy().N("Ceiling"))

	lampMaterial := scn.NewMaterial().N("Lamp").E(color.White, lightIntensityFactor, true)

	if singleLight {
		cornellBox.RemoveObjectsBySubstructureName("Lamp_1_-_left_away")
		cornellBox.RemoveObjectsBySubstructureName("Lamp_2_-_left_close")
		cornellBox.RemoveObjectsBySubstructureName("Lamp_3_-_right_away")
		cornellBox.RemoveObjectsBySubstructureName("Lamp_4_-_right_close")

		lampPercentageOfCeiling := 2.0 / 3.0 // Two thirds in width and depth (i.e. 0.666*0.666 = 44.4% of the ceiling)
		lamp := createSingleLamp(scale, lampPercentageOfCeiling, lampMaterial)

		cornellBox.FacetStructures = append(cornellBox.FacetStructures, lamp)
	} else {
		cornellBox.GetFirstObjectBySubstructureName("Lamp_1_-_left_away").Material = lampMaterial
		cornellBox.GetFirstObjectBySubstructureName("Lamp_2_-_left_close").Material = lampMaterial
		cornellBox.GetFirstObjectBySubstructureName("Lamp_3_-_right_away").Material = lampMaterial
		cornellBox.GetFirstObjectBySubstructureName("Lamp_4_-_right_close").Material = lampMaterial
	}

	cornellBox.UpdateBounds()

	return cornellBox
}

func createSingleLamp(cornellBoxScale *vec3.T, lampPercentageOfCeiling float64, lampMaterial *scn.Material) *scn.FacetStructure {
	lampSizeX := (cornellBoxScale[0] / 2) * lampPercentageOfCeiling
	lampY := cornellBoxScale[1] - 0.001
	lampSizeZ := (cornellBoxScale[2] / 2) * lampPercentageOfCeiling

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
	return lamp
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
