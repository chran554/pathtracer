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

var renderType = scn.Pathtracing

//var renderType = scn.Raycasting

var maxRecursionDepth = 3
var amountSamples = 1000
var lensRadius = 5.0

var viewPlaneDistance = 800.0
var cameraDistanceFactor = 1.0

var imageWidth = 450
var imageHeight = 450
var magnification = 2.0

func main() {
	width := int(float64(imageWidth) * magnification)
	height := int(float64(imageHeight) * magnification)

	// Keep image proportions to an even amount of pixel for mp4 encoding
	if width%2 == 1 {
		width++
	}
	if height%2 == 1 {
		height++
	}

	cornellBox := GetCornellBox(&vec3.T{500, 300, 500}, 9.0) // cm, as units. I.e. a 5x3x5m room

	var plyFilename = "beethoven.ply"
	var plyFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/ply/" + plyFilename

	plyFile, err := os.Open(plyFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", plyFilenamePath, err.Error())
		return
	}
	defer plyFile.Close()

	plyObject, err := ply.ReadPly(plyFile)
	if err != nil {
		fmt.Printf("could not read ply-file '%s': %s", plyFile.Name(), err.Error())
		return
	}
	plyObject.Translate(&vec3.T{0, -plyObject.Bounds.Ymin, 0})
	plyObject.ScaleUniform(&vec3.Zero, 30.0)
	plyObject.RotateY(&vec3.Zero, math.Pi-math.Pi/12.0)
	plyObject.UpdateBounds()
	plyObject.UpdateNormals()

	fmt.Printf("ply object bounds: %+v\n", plyObject.Bounds)

	plyProjection := scn.NewParallelImageProjection("textures/white_marble.png", vec3.Zero, vec3.UnitX.Scaled(200), vec3.UnitY.Scaled(200))
	plyMaterial := &scn.Material{
		Color:      &color.White,
		Glossiness: 0.1,
		Roughness:  0.6,
		Projection: &plyProjection,
	}
	plyObject.Material = plyMaterial

	scene := scn.SceneNode{
		FacetStructures: []*scn.FacetStructure{cornellBox, plyObject},
	}

	camera := getCamera()

	animation := scn.Animation{
		AnimationName:     animationName,
		Frames:            []scn.Frame{},
		Width:             width,
		Height:            height,
		WriteRawImageFile: false,
	}

	frame := scn.Frame{
		Filename:   animation.AnimationName,
		FrameIndex: 0,
		Camera:     &camera,
		SceneNode:  &scene,
	}

	animation.Frames = append(animation.Frames, frame)

	anm.WriteAnimationToFile(animation, false)
}

func GetCornellBox(scale *vec3.T, lightIntensityFactor float32) *scn.FacetStructure {
	var cornellBoxFilename = "cornellbox.obj"
	var cornellBoxFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/" + cornellBoxFilename

	cornellBoxFile, err := os.Open(cornellBoxFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", cornellBoxFilenamePath, err.Error())
	}
	defer cornellBoxFile.Close()

	cornellBox, err := obj.Read(cornellBoxFile)
	cornellBox.Scale(&vec3.Zero, scale)
	cornellBox.ClearMaterials()

	cornellBox.Material = &scn.Material{
		Name:  "Cornell box default material",
		Color: &color.Color{R: 0.95, G: 0.95, B: 0.95},
		//Roughness: 0.0,
	}

	lampMaterial := scn.Material{
		Name:          "Lamp",
		Color:         &color.Color{R: 1.0, G: 1.0, B: 1.0},
		Emission:      (&color.Color{R: 1.0, G: 1.0, B: 1.0}).Multiply(lightIntensityFactor),
		RayTerminator: true,
	}
	cornellBox.GetFirstObjectByName("Lamp_1").Material = &lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_2").Material = &lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_3").Material = &lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_4").Material = &lampMaterial

	backWallProjection := scn.NewParallelImageProjection("textures/wallpaper/anemone-rose-flower-eucalyptus-leaves-pampas-grass.png", vec3.Zero, vec3.UnitX.Scaled(scale[0]), vec3.UnitY.Scaled(scale[0]*0.66))
	backWallMaterial := *cornellBox.Material
	backWallMaterial.Projection = &backWallProjection
	cornellBox.GetFirstObjectByName("Wall_back").Material = &backWallMaterial

	sideWallProjection := scn.NewParallelImageProjection("textures/wallpaper/anemone-rose-flower-eucalyptus-leaves-pampas-grass.png", vec3.Zero, vec3.UnitZ.Scaled(scale[0]), vec3.UnitY.Scaled(scale[0]*0.66))
	sideWallMaterial := *cornellBox.Material
	sideWallMaterial.Projection = &sideWallProjection
	cornellBox.GetFirstObjectByName("Wall_left").Material = &sideWallMaterial
	cornellBox.GetFirstObjectByName("Wall_right").Material = &sideWallMaterial

	floorProjection := scn.NewParallelImageProjection("textures/tilesf4.jpeg", vec3.Zero, vec3.UnitX.Scaled(scale[0]*0.5), vec3.UnitZ.Scaled(scale[0]*0.5))
	floorMaterial := *cornellBox.Material
	floorMaterial.Glossiness = 0.1
	floorMaterial.Roughness = 0.2
	floorMaterial.Projection = &floorProjection
	cornellBox.GetFirstObjectByName("Floor_2").Material = &floorMaterial

	return cornellBox
}

func getCamera() scn.Camera {
	origin := vec3.T{0, 60 * 3, -800}
	origin.Scale(cameraDistanceFactor)

	focusPoint := vec3.T{0, 60 * 3, -30 * 3}

	heading := focusPoint.Subed(&origin)
	focusDistance := heading.Length()

	return scn.Camera{
		Origin:            &origin,
		Heading:           &heading,
		ViewUp:            &vec3.T{0, 1, 0},
		ViewPlaneDistance: viewPlaneDistance,
		ApertureSize:      lensRadius,
		FocusDistance:     focusDistance,
		Samples:           amountSamples,
		AntiAlias:         true,
		Magnification:     magnification,
		RenderType:        renderType,
		RecursionDepth:    maxRecursionDepth,
	}
}
