package main

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"os"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/ply"
	scn "pathtracer/internal/pkg/scene"
)

var animationName = "ply_file_test"

var amountSamples = 1024 // 2500
var lensRadius = 5.0

var cameraDistanceFactor = 1.0

var imageWidth = 450
var imageHeight = 450
var magnification = 2.0

func main() {
	cornellBox := GetCornellBox(&vec3.T{500, 300, 500}, 9.0) // cm, as units. I.e. a 5x3x5m room

	boxHeight := 40.0
	boxWidth := 120.0
	box := obj.NewBox(obj.BoxPositive)
	box.Material = scn.NewMaterial().
		C(color.NewColorGrey(0.9)).
		M(0.4, 0.1).
		PP("textures/concrete/Polished-Concrete-Architextures.jpg", &vec3.T{0, 0, 0}, (&vec3.UnitX).Scaled(boxWidth), (&vec3.UnitZ).Add(&vec3.T{0, 0.5, 0}).Scaled(boxWidth))
	box.Translate(&vec3.T{-0.5, 0, -0.5})
	box.Scale(&vec3.Zero, &vec3.T{boxWidth * 2, boxHeight, boxWidth * 2})

	var plyFilename = "beethoven.ply"
	var plyFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/ply/" + plyFilename

	plyFile, err := os.Open(plyFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", plyFilenamePath, err.Error())
		return
	}
	defer plyFile.Close()

	plyFacetStructure, err := ply.ReadPlyFile(plyFile)
	if err != nil {
		fmt.Printf("could not read ply-file '%s': %s", plyFile.Name(), err.Error())
		return
	}
	plyFacetStructure.Translate(&vec3.T{0, -plyFacetStructure.Bounds.Ymin, 0})
	plyFacetStructure.ScaleUniform(&vec3.Zero, 25.0)
	plyFacetStructure.RotateY(&vec3.Zero, math.Pi-math.Pi/12.0)
	plyFacetStructure.Translate(&vec3.T{0, boxHeight, 0})
	plyFacetStructure.UpdateBounds()
	plyFacetStructure.UpdateNormals()
	//plyFacetStructure.UpdateVertexNormals()

	fmt.Printf("ply object bounds: %+v\n", plyFacetStructure.Bounds)

	plyFacetStructure.Material = scn.NewMaterial().M(0.1, 0.6).PP("textures/marble/white_marble.png", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(200), vec3.UnitY.Scaled(200))

	scene := scn.NewSceneNode().FS(cornellBox, box, plyFacetStructure)

	cameraOrigin := vec3.T{0, 60 * 3, -800}
	cameraOrigin.Scale(cameraDistanceFactor)
	focusPoint := vec3.T{0, 60 * 3, -30 * 3}
	camera := scn.NewCamera(&cameraOrigin, &focusPoint, amountSamples, magnification).A(lensRadius, "")

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false)

	frame := scn.NewFrame(animation.AnimationName, -1, camera, scene)

	animation.AddFrame(frame)

	anm.WriteAnimationToFile(animation, false)
}

func GetCornellBox(scale *vec3.T, lightIntensityFactor float64) *scn.FacetStructure {
	var cornellBoxFilename = "cornellbox.obj"
	var cornellBoxFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + cornellBoxFilename

	cornellBox := obj.ReadOrPanic(cornellBoxFilenamePath)

	cornellBox.Scale(&vec3.Zero, scale)
	cornellBox.ClearMaterials()

	cornellBox.Material = scn.NewMaterial().N("Cornell box default material").C(color.Color{R: 0.95, G: 0.95, B: 0.95})

	lampMaterial := scn.NewMaterial().N("Lamp").E(color.White, lightIntensityFactor, true)

	cornellBox.GetFirstObjectByName("Lamp_1").Material = lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_2").Material = lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_3").Material = lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_4").Material = lampMaterial

	backWallMaterial := *cornellBox.Material
	backWallMaterial.PP("textures/wallpaper/geometric-yellow-wallpaper.jpg", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(scale[0]), vec3.UnitY.Scaled(scale[0]*0.66))
	cornellBox.GetFirstObjectByName("Wall_back").Material = &backWallMaterial

	sideWallMaterial := *cornellBox.Material
	sideWallMaterial.PP("textures/wallpaper/geometric-yellow-wallpaper.jpg", &vec3.T{0, 0, 0}, vec3.UnitZ.Scaled(scale[0]), vec3.UnitY.Scaled(scale[0]*0.66))
	cornellBox.GetFirstObjectByName("Wall_left").Material = &sideWallMaterial
	cornellBox.GetFirstObjectByName("Wall_right").Material = &sideWallMaterial

	floorMaterial := *cornellBox.Material
	floorMaterial.M(0.6, 0.1).PP("textures/floor/Calacatta-Vena-French-Pattern-Architextures.jpg", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(scale[0]/2), vec3.UnitZ.Scaled(scale[0]/2))
	cornellBox.GetFirstObjectByName("Floor_2").Material = &floorMaterial

	return cornellBox
}
