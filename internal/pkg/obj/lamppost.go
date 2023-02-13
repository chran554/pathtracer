package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"os"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
)

func NewLampPost(lampPostScale *vec3.T) *scn.FacetStructure {
	lampPostMaterial := scn.Material{
		Name:          "lamp post",
		Color:         &color.Color{R: 0.9, G: 0.4, B: 0.3},
		Emission:      &color.Black,
		Glossiness:    0.2,
		Roughness:     0.3,
		RayTerminator: false,
	}

	lampMaterial := scn.NewMaterial().N("lamp").C(color.White, 1.0).E(color.White, 2.0, true)

	lampPost := loadLampPost(lampPostScale)
	lampPost.ClearMaterials()
	lampPost.Material = &lampPostMaterial

	lampPost.GetFirstObjectByName("lamp_0").Material = lampMaterial
	lampPost.GetFirstObjectByName("lamp_1").Material = lampMaterial
	lampPost.GetFirstObjectByName("lamp_2").Material = lampMaterial
	lampPost.GetFirstObjectByName("lamp_3").Material = lampMaterial
	return lampPost
}

func loadLampPost(scale *vec3.T) *scn.FacetStructure {
	var objFilename = "lamp_post.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objFilename

	objFile, err := os.Open(objFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", objFilenamePath, err.Error())
	}
	defer objFile.Close()

	obj, err := Read(objFile)
	ymin := obj.Bounds.Ymin
	ymax := obj.Bounds.Ymax
	obj.Translate(&vec3.T{0.0, -ymin, 0.0})       // lamp post base touch the ground (xz-plane)
	obj.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize to height == 1.0

	obj.Scale(&vec3.Zero, scale)

	obj.UpdateBounds()
	fmt.Printf("Lamp post bounds: %+v\n", obj.Bounds)

	return obj
}
