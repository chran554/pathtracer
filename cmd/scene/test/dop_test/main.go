package main

import (
	"fmt"
	"os"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"
	"strconv"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "dop_test"

var ballRadius float64 = 30

var renderType = scn.Pathtracing

// var renderType = scn.Raycasting

var maxRecursionDepth = 4
var amountSamples = 128 * 4 * 4
var lensRadius = 12.0 // 7.0 // 15.0
var antiAlias = true

var viewPlaneDistance = 2000.0
var cameraDistanceFactor = 1.0

var imageWidth = 800
var imageHeight = 400
var magnification = 1.0

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

	var spheres []*scn.Sphere

	amountSpheres := 5
	sphereSpread := ballRadius * 2.0 * (float64(amountSpheres) + 1)
	sphereCC := sphereSpread / float64(amountSpheres)

	sphereMaterial := scn.Material{
		Color:      color.Color{R: 0.85, G: 0.95, B: 0.80},
		Glossiness: 0.40,
		Roughness:  0.05,
	}

	for i := 0; i <= amountSpheres; i++ {
		positionOffsetX := (-sphereSpread/2.0 + float64(i)*sphereCC) * 0.5
		positionOffsetZ := (-sphereSpread/2.0 + float64(i)*sphereCC) * 1.0

		sphere := scn.Sphere{
			Name:     "Glass sphere with transparency " + strconv.Itoa(i),
			Origin:   vec3.T{positionOffsetX, ballRadius, positionOffsetZ},
			Radius:   ballRadius,
			Material: &sphereMaterial,
		}
		spheres = append(spheres, &sphere)
	}

	//environment := getEnvironmentMapping()
	//spheres = append(spheres, &environment)

	scene := scn.SceneNode{
		Spheres:         spheres,
		FacetStructures: []*scn.FacetStructure{cornellBox},
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
		Color: color.Color{R: 0.95, G: 0.95, B: 0.95},
		//Roughness: 0.0,
	}

	lampMaterial := scn.Material{
		Name:          "Lamp",
		Color:         color.Color{R: 1.0, G: 1.0, B: 1.0},
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
	origin := vec3.T{0, ballRadius * 3, -800}
	origin.Scale(cameraDistanceFactor)

	heading := vec3.T{-origin[0], -(origin[1] - ballRadius), -origin[2]}
	focalDistance := heading.Length()

	return scn.Camera{
		Origin:            &origin,
		Heading:           &heading,
		ViewUp:            &vec3.T{0, 1, 0},
		ViewPlaneDistance: viewPlaneDistance,
		LensRadius:        lensRadius,
		FocalDistance:     focalDistance,
		Samples:           amountSamples,
		AntiAlias:         antiAlias,
		Magnification:     magnification,
		RenderType:        renderType,
		RecursionDepth:    maxRecursionDepth,
	}
}
