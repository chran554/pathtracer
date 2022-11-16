package main

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"math/rand"
	"os"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"
)

var animationName = "aperture_shape_test2"

var renderType = scn.Pathtracing

var amountFrames = 90

var maxRecursionDepth = 3
var amountSamples = 200
var apertureRadius = 50.0

var viewPlaneDistance = 600.0

var imageWidth = 1000
var imageHeight = 800
var magnification = 0.4

var lampEmissionFactor float32 = 1.5
var sphereEmissionFactor float32 = 6.0
var amountSpheres = 6
var sphereRadius float64 = 8
var distanceFactor float64 = 14

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

	cornellBox := GetCornellBox(&vec3.T{500, 300, 500}, lampEmissionFactor) // cm, as units. I.e. a 5x3x5m room

	sphereSpread := distanceFactor * 6.0

	quadrantSpheres := make([][]*scn.Sphere, 4)
	for i := range quadrantSpheres {
		quadrantSpheres[i] = make([]*scn.Sphere, 0)
	}

	for sz := 0; sz < amountSpheres; sz++ {
		for sx := 0; sx < amountSpheres; sx++ {
			positionOffsetX := sphereSpread * float64(sx+1)
			positionOffsetZ := sphereSpread * float64(sz)

			r := float32(0.5 + rand.Float64()*0.5)
			g := float32(0.5 + rand.Float64()*0.5)
			b := float32(0.5 + rand.Float64()*0.5)

			sphereMaterial := scn.Material{
				Color:         &color.Color{R: r, G: g, B: b},
				Emission:      (&color.Color{R: r, G: g, B: b}).Multiply(sphereEmissionFactor),
				Glossiness:    0.0,
				Roughness:     1.0,
				RayTerminator: true,
			}

			for quad := 0; quad < 4; quad++ {
				sphere := scn.Sphere{
					Origin:   &vec3.T{positionOffsetX, sphereRadius * 3.0, positionOffsetZ},
					Radius:   sphereRadius,
					Material: &sphereMaterial,
				}

				sphere.RotateY(&vec3.Zero, float64(quad)*math.Pi/2)

				quadrantSpheres[quad] = append(quadrantSpheres[quad], &sphere)
			}
		}
	}

	scene := scn.SceneNode{
		FacetStructures: []*scn.FacetStructure{cornellBox},
		ChildNodes: []*scn.SceneNode{
			{Spheres: quadrantSpheres[0]},
			{Spheres: quadrantSpheres[1]},
			{Spheres: quadrantSpheres[2]},
			{Spheres: quadrantSpheres[3]},
		},
	}

	animation := scn.Animation{
		AnimationName:     animationName,
		Frames:            []scn.Frame{},
		Width:             width,
		Height:            height,
		WriteRawImageFile: false,
	}

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		camera := getCamera(animationProgress)

		frame := scn.Frame{
			Filename:   animationName + "_" + fmt.Sprintf("%06d", frameIndex),
			FrameIndex: 0,
			Camera:     &camera,
			SceneNode:  &scene,
		}

		animation.Frames = append(animation.Frames, frame)
	}

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

	//backWallProjection := scn.NewParallelImageProjection("textures/wallpaper/anemone-rose-flower-eucalyptus-leaves-pampas-grass.png", vec3.Zero, vec3.UnitX.Scaled(scale[0]), vec3.UnitY.Scaled(scale[0]*0.66))
	//backWallMaterial := *cornellBox.Material
	//backWallMaterial.Projection = &backWallProjection
	//cornellBox.GetFirstObjectByName("Wall_back").Material = &backWallMaterial

	//sideWallProjection := scn.NewParallelImageProjection("textures/wallpaper/anemone-rose-flower-eucalyptus-leaves-pampas-grass.png", vec3.Zero, vec3.UnitZ.Scaled(scale[0]), vec3.UnitY.Scaled(scale[0]*0.66))
	//sideWallMaterial := *cornellBox.Material
	//sideWallMaterial.Projection = &sideWallProjection
	//cornellBox.GetFirstObjectByName("Wall_left").Material = &sideWallMaterial
	//cornellBox.GetFirstObjectByName("Wall_right").Material = &sideWallMaterial

	projectionZoom := 0.33
	//floorProjection := scn.NewParallelImageProjection("textures/7451-diffuse 02.png", vec3.Zero, vec3.UnitX.Scaled(scale[0]*projectionZoom), vec3.UnitZ.Scaled(scale[0]*projectionZoom))
	floorProjection := scn.NewParallelImageProjection("textures/tilesf4.jpeg", vec3.Zero, vec3.UnitX.Scaled(scale[0]*projectionZoom), vec3.UnitZ.Scaled(scale[0]*projectionZoom))
	floorMaterial := *cornellBox.Material
	floorMaterial.Glossiness = 0.0
	floorMaterial.Roughness = 1.0
	floorMaterial.Projection = &floorProjection
	cornellBox.GetFirstObjectByName("Floor_2").Material = &floorMaterial

	return cornellBox
}

func getCamera(animationProgress float64) scn.Camera {
	cameraOrigin := vec3.T{0, distanceFactor * 15.0, -500}
	focusPoint := vec3.T{0, distanceFactor * 2.0, -distanceFactor * 22.0}

	// Animation
	angle := (math.Pi / 2.0) * animationProgress
	scn.RotateY(&cameraOrigin, &vec3.Zero, angle)
	scn.RotateY(&focusPoint, &vec3.Zero, angle)

	heading := focusPoint.Subed(&cameraOrigin)
	focalDistance := heading.Length() * 1.75

	return scn.Camera{
		Origin:            &cameraOrigin,
		Heading:           &heading,
		ViewUp:            &vec3.T{0, 1, 0},
		ViewPlaneDistance: viewPlaneDistance,
		ApertureSize:      apertureRadius,
		//ApertureShape:     "textures/aperture/heart.png",
		//ApertureShape: "textures/aperture/hexagon.png",
		//ApertureShape: "textures/aperture/letter_F.png",
		ApertureShape:  "textures/aperture/star.png", // Done
		FocusDistance:  focalDistance,
		Samples:        amountSamples,
		AntiAlias:      true,
		Magnification:  magnification,
		RenderType:     renderType,
		RecursionDepth: maxRecursionDepth,
	}
}
