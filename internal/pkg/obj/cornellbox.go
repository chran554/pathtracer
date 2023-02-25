package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"os"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
)

func NewCornellBox(scale *vec3.T, lightIntensityFactor float64) *scn.FacetStructure {
	var cornellBoxFilename = "cornellbox.obj"
	var cornellBoxFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + cornellBoxFilename

	cornellBoxFile, err := os.Open(cornellBoxFilenamePath)
	if err != nil {
		message := fmt.Sprintf("ouupps, something went wrong loading file: '%s'\n%s\n", cornellBoxFilenamePath, err.Error())
		panic(message)
	}
	defer cornellBoxFile.Close()

	cornellBox, err := Read(cornellBoxFile)
	cornellBox.Scale(&vec3.Zero, scale)
	cornellBox.ClearMaterials()

	cornellBox.Material = scn.NewMaterial().N("cornellbox").C(color.NewColorGrey(0.95))

	lampMaterial := scn.NewMaterial().N("lamp").E(color.White, lightIntensityFactor, true)

	cornellBox.GetFirstObjectByName("Lamp_1").Material = lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_2").Material = lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_3").Material = lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_4").Material = lampMaterial

	backWallProjection := scn.NewParallelImageProjection("textures/wallpaper/anemone-rose-flower-eucalyptus-leaves-pampas-grass.png", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(scale[0]), vec3.UnitY.Scaled(scale[0]*0.66))
	backWallMaterial := *cornellBox.Material
	backWallMaterial.Projection = &backWallProjection
	cornellBox.GetFirstObjectByName("Wall_back").Material = &backWallMaterial

	sideWallProjection := scn.NewParallelImageProjection("textures/wallpaper/anemone-rose-flower-eucalyptus-leaves-pampas-grass.png", &vec3.T{0, 0, 0}, vec3.UnitZ.Scaled(scale[0]), vec3.UnitY.Scaled(scale[0]*0.66))
	sideWallMaterial := *cornellBox.Material
	sideWallMaterial.Projection = &sideWallProjection
	cornellBox.GetFirstObjectByName("Wall_left").Material = &sideWallMaterial
	cornellBox.GetFirstObjectByName("Wall_right").Material = &sideWallMaterial

	floorProjection := scn.NewParallelImageProjection("textures/floor/tilesf4.jpeg", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(scale[0]*0.5), vec3.UnitZ.Scaled(scale[0]*0.5))
	floorMaterial := *cornellBox.Material
	floorMaterial.M(0.1, 0.2).P(&floorProjection)
	cornellBox.GetFirstObjectByName("Floor_2").Material = &floorMaterial

	return cornellBox
}
