package main

import (
	"fmt"
	"math"
	"os"
	"pathtracer/cmd/scene/diamondsR4ever/diamond"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "diamondsR4ever"

var cornellBoxFilename = "cornellbox.obj"
var cornellBoxFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/" + cornellBoxFilename

var amountAnimationFrames = 1

var ballRadius float64 = 60

var imageWidth = 800
var imageHeight = 800
var magnification = 1.0 // * 0.5

var renderType = scn.Pathtracing
var maxRecursionDepth = 2
var amountSamples = 128 // * 50 / 4
var lensRadius float64 = 0
var antiAlias = true

var viewPlaneDistance = 1000.0
var cameraDistanceFactor = 1.0

func main() {
	animation := scn.Animation{
		AnimationName:     animationName,
		Frames:            []scn.Frame{},
		Width:             int(float64(imageWidth) * magnification),
		Height:            int(float64(imageHeight) * magnification),
		WriteRawImageFile: true,
	}

	roomScale := 100.0
	// cornellBox := getCornellBox(cornellBoxFilenamePath, 100.0)
	diamond := diamond.GetDiamondRoundBrilliantCut(75.0)
	diamond.UpdateBounds()

	/*	environmentSphere := scn.Sphere{
			Name:   "Environment light",
			Origin: vec3.T{0, 0, 0},
			Radius: roomScale * 10,
			Material: &scn.Material{
				Color:         color.Color{R: 1.0, G: 1.0, B: 1.0},
				Emission:      &color.Color{R: 0.1, G: 0.1, B: 0.1},
				RayTerminator: true,
			},
		}
	*/
	lamp1 := scn.Sphere{
		Name:   "Lamp1",
		Origin: vec3.T{roomScale * 2.0, roomScale * 0.5, -roomScale * 1.0},
		Radius: roomScale * 0.8,
		Material: &scn.Material{
			Color:         color.Color{R: 1.0, G: 1.0, B: 1.0},
			Emission:      &color.Color{R: 15.0, G: 15.0, B: 15.0},
			RayTerminator: false,
		},
	}

	lamp2 := scn.Sphere{
		Name:   "Lamp2",
		Origin: vec3.T{-roomScale * 2.0, roomScale * 2.0, -roomScale * 1.0},
		Radius: roomScale * 0.8,
		Material: &scn.Material{
			Color:         color.Color{R: 1.0, G: 1.0, B: 1.0},
			Emission:      &color.Color{R: 3.0, G: 3.0, B: 3.0},
			RayTerminator: false,
		},
	}

	floorLevel := diamond.Bounds.Ymin
	floor := scn.Disc{
		Name:   "Floor",
		Origin: &vec3.T{0, floorLevel, 0},
		Normal: &vec3.T{0, 1, 0},
		Radius: roomScale * 10,
		Material: &scn.Material{
			Color:         color.Color{R: 0.90, G: 0.85, B: 0.95},
			Roughness:     1.0,
			RayTerminator: false,
		},
	}

	scene := scn.SceneNode{
		Spheres:         []*scn.Sphere{&lamp1, &lamp2},
		Discs:           []*scn.Disc{&floor},
		FacetStructures: []*scn.FacetStructure{diamond},
	}

	for animationFrameIndex := 0; animationFrameIndex < amountAnimationFrames; animationFrameIndex++ {
		animationStep := 1.0 / float64(amountAnimationFrames)
		// animationProgress := float64(animationFrameIndex) * animationStep
		diamond.RotateY(&vec3.Zero, animationStep*(math.Pi*2.0/8.0))

		cameraOrigin := vec3.T{20, roomScale * 0.25, -100}
		cameraOrigin.Scale(cameraDistanceFactor)
		focusPoint := vec3.T{0, 0, 0}

		//maxHeightAngle := math.Pi / 14
		//heightAngle := maxHeightAngle * (1.0 + math.Cos(animationProgress*2*math.Pi-math.Pi)) / 2.0
		//rotationXMatrix := mat3.T{}
		//rotationXMatrix.AssignXRotation(heightAngle)
		//animatedCameraOrigin := rotationXMatrix.MulVec3(&cameraOrigin)

		//maxSideAngle := math.Pi / 8
		//sideAngle := maxSideAngle * -1 * math.Sin(animationProgress*2*math.Pi)
		//rotationYMatrix := mat3.T{}
		//rotationYMatrix.AssignYRotation(sideAngle)
		//animatedCameraOrigin = rotationYMatrix.MulVec3(&animatedCameraOrigin)

		camera := getCamera(&cameraOrigin, &focusPoint)

		frame := scn.Frame{
			Filename:   animation.AnimationName + "_" + fmt.Sprintf("%06d", animationFrameIndex),
			FrameIndex: animationFrameIndex,
			Camera:     &camera,
			SceneNode:  &scene,
		}

		animation.Frames = append(animation.Frames, frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

func getCornellBox(cornellBoxFilenamePath string, scale float64) *scn.FacetStructure {
	cornellBoxFile, err := os.Open(cornellBoxFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", cornellBoxFilenamePath, err.Error())
	}

	cornellBox, err := obj.Read(cornellBoxFile)
	cornellBox.ScaleUniform(&vec3.Zero, scale)

	setObjectMaterial(cornellBox, "Wall_right", &color.Color{R: 0.9, G: 0.9, B: 0.9}, nil, false, 0.0, 0.0)
	setObjectMaterial(cornellBox, "Wall_left", &color.Color{R: 0.9, G: 0.9, B: 0.9}, nil, false, 0.0, 0.0)

	setObjectMaterial(cornellBox, "Wall_back", &color.Color{R: 0.9, G: 0.9, B: 0.9}, nil, false, 0.0, 0.0)
	setObjectMaterial(cornellBox, "Ceiling", &color.Color{R: 0.9, G: 0.9, B: 0.9}, nil, false, 0.0, 0.0)
	setObjectMaterial(cornellBox, "Floor_2", &color.Color{R: 1.0, G: 1.0, B: 1.0}, nil, false, 0.0, 0.0)

	// setObjectProjection1(cornellBox, "Floor_2")

	lampEmission := color.Color{R: 9.0, G: 9.0, B: 9.0}
	lampColor := color.Color{R: 1.0, G: 1.0, B: 1.0}
	setObjectMaterial(cornellBox, "Lamp_1", &lampColor, &lampEmission, true, 0.0, 0)
	setObjectMaterial(cornellBox, "Lamp_2", &lampColor, &lampEmission, true, 0.0, 0)
	setObjectMaterial(cornellBox, "Lamp_3", &lampColor, &lampEmission, true, 0.0, 0)
	setObjectMaterial(cornellBox, "Lamp_4", &lampColor, &lampEmission, true, 0.0, 0)

	return cornellBox
}

func setObjectMaterial(openBox *scn.FacetStructure, objectName string, color *color.Color, emission *color.Color, rayTerminator bool, glossiness float32, roughness float32) {
	object := openBox.GetFirstObjectByName(objectName)
	if object != nil {
		object.Material.Color = *color
		object.Material.Emission = emission
		object.Material.RayTerminator = rayTerminator
		object.Material.Glossiness = glossiness
		object.Material.Roughness = roughness
	} else {
		fmt.Printf("No " + objectName + " found")
		os.Exit(1)
	}
}

func setObjectProjection1(openBox *scn.FacetStructure, objectName string) {
	object := openBox.GetFirstObjectByName(objectName)
	if object != nil {
		object.Material.Projection = &scn.ImageProjection{
			ProjectionType: scn.Parallel,
			ImageFilename:  "textures/white_marble.png",
			Origin:         vec3.Zero,
			U:              vec3.T{200 * 2, 0, 0},
			V:              vec3.T{0, 0, 200 * 2},
			RepeatU:        true,
			RepeatV:        true,
			FlipU:          false,
			FlipV:          false,
			Gamma:          0,
		}
	} else {
		fmt.Printf("No " + objectName + " found")
		os.Exit(1)
	}
}

func getCamera(cameraOrigin, focusPoint *vec3.T) scn.Camera {
	heading := focusPoint.Subed(cameraOrigin)
	focalDistance := heading.Length()

	return scn.Camera{
		Origin:            cameraOrigin,
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
