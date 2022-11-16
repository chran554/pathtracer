package main

import (
	"fmt"
	"math"
	"os"
	"pathtracer/cmd/obj/diamond/pkg/diamond"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "diamondsR4ever"

var cornellBoxFilename = "cornellbox.obj"
var cornellBoxFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/" + cornellBoxFilename

var amountAnimationFrames = 60

var imageWidth = 800
var imageHeight = 800
var magnification = 1.0

var renderType = scn.Pathtracing
var maxRecursionDepth = 2
var amountSamples = 128 * 8
var lensRadius float64 = 0

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

	environmentSphere := scn.Sphere{
		Name:   "Environment light",
		Origin: &vec3.T{0, 0, 0},
		Radius: 1000 * 1000 * 1000,
		Material: &scn.Material{
			Color:         &color.Color{R: 1.0, G: 1.0, B: 1.0},
			Emission:      &color.Color{R: 1.0, G: 1.0, B: 1.0},
			RayTerminator: true,
			Projection: &scn.ImageProjection{
				ProjectionType: scn.Spherical,
				ImageFilename:  "textures/planets/environmentmap/Stellarium3.jpeg",
				Origin:         &vec3.T{0, 0, 0},
				U:              &vec3.T{1, 0, 0},
				V:              &vec3.T{0, 1, 0},
				RepeatU:        false,
				RepeatV:        false,
				FlipU:          false,
				FlipV:          false,
				Gamma:          0,
			},
		},
	}

	lamp1 := scn.Sphere{
		Name:   "Lamp1",
		Origin: &vec3.T{roomScale * 2.0, roomScale * 0.5, -roomScale * 1.0},
		Radius: roomScale * 0.8,
		Material: &scn.Material{
			Color:         &color.Color{R: 1.0, G: 1.0, B: 1.0},
			Emission:      &color.Color{R: 20.0, G: 20.0, B: 20.0},
			RayTerminator: false,
		},
	}

	lamp2 := scn.Sphere{
		Name:   "Lamp2",
		Origin: &vec3.T{-roomScale * 2.0, roomScale * 0.5, -roomScale * 1.0},
		Radius: roomScale * 0.8,
		Material: &scn.Material{
			Color:         &color.Color{R: 1.0, G: 1.0, B: 1.0},
			Emission:      &color.Color{R: 9.0, G: 9.0, B: 9.0},
			RayTerminator: false,
		},
	}

	diamondMaterial := scn.Material{
		Color:           &color.Color{R: 0.4, G: 0.4, B: 0.4},
		Glossiness:      0.8,
		Roughness:       0.005,
		RefractionIndex: 2.42,
		Transparency:    0.0,
	}

	d := diamond.Diamond{
		GirdleDiameter:                         1.00, // 100.0%
		GirdleHeightRelativeGirdleDiameter:     0.03, //   3.0%
		CrownAngleDegrees:                      34.0, //  34.0°
		TableFacetSizeRelativeGirdle:           0.56, //  56.0%
		StarFacetSizeRelativeCrownSide:         0.55, //  55.0%
		PavilionAngleDegrees:                   41.1, //  41.1°
		LowerHalfFacetSizeRelativeGirdleRadius: 0.77, //  77.0%
	}

	diamond2 := diamond.NewDiamondRoundBrilliantCut(d, 75.0, diamondMaterial)
	diamond2.UpdateBounds()
	floorLevel := diamond2.Bounds.Ymin

	floor := scn.Disc{
		Name:   "Floor",
		Origin: &vec3.T{0, floorLevel, 0},
		Normal: &vec3.T{0, 1, 0},
		Radius: roomScale * 1.5,
		Material: &scn.Material{
			Color:         &color.Color{R: 1.5, G: 1.5, B: 1.5},
			Roughness:     0.1,
			Glossiness:    0.8,
			RayTerminator: false,
			Projection: &scn.ImageProjection{
				ProjectionType: scn.Parallel,
				ImageFilename:  "textures/white_marble.png",
				Origin:         &vec3.T{0, 0, 0},
				U:              &vec3.T{roomScale * 1.5 * 2, 0, 0},
				V:              &vec3.T{0, 0, roomScale * 1.5 * 2},
				RepeatU:        true,
				RepeatV:        true,
				FlipU:          false,
				FlipV:          false,
				Gamma:          0,
			},
		},
	}

	animationStep := 1.0 / float64(amountAnimationFrames)
	for animationFrameIndex := 0; animationFrameIndex < amountAnimationFrames; animationFrameIndex++ {
		animationProgress := float64(animationFrameIndex) * animationStep

		diamond := diamond.NewDiamondRoundBrilliantCut(d, 75.0, diamondMaterial)
		diamond.RotateY(&vec3.Zero, animationProgress*(math.Pi*2.0/8.0))
		diamond.RotateZ(&vec3.Zero, -15.0*animationStep*(math.Pi*2.0/8.0))
		diamond.RotateY(&vec3.Zero, 10.0*animationStep*(math.Pi*2.0/8.0))
		diamond.UpdateBounds()

		scene := scn.SceneNode{
			Spheres:         []*scn.Sphere{&lamp1, &lamp2, &environmentSphere},
			Discs:           []*scn.Disc{&floor},
			FacetStructures: []*scn.FacetStructure{diamond},
		}

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
		//animatedCameraOrigin := rotationYMatrix.MulVec3(&cameraOrigin)
		//animatedCameraOrigin = rotationYMatrix.MulVec3(&canimatedCameraOrigin)

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
	defer cornellBoxFile.Close()

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
		object.Material.Color = color
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
			Origin:         &vec3.Zero,
			U:              &vec3.T{200 * 2, 0, 0},
			V:              &vec3.T{0, 0, 200 * 2},
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
		ApertureSize:      lensRadius,
		FocusDistance:     focalDistance,
		Samples:           amountSamples,
		AntiAlias:         true,
		Magnification:     magnification,
		RenderType:        renderType,
		RecursionDepth:    maxRecursionDepth,
	}
}
