package main

import (
	"fmt"
	"github.com/ungerik/go3d/float64/mat3"
	"math"
	"os"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "objectfile_test"

var cornellBoxFilename = "cornellbox.obj"
var cornellBoxFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/" + cornellBoxFilename

var amountAnimationFrames = 90

var ballRadius float64 = 60

var renderType = scn.Pathtracing
var maxRecursionDepth = 2 * 2
var amountSamples = 128 * 4
var lensRadius float64 = 0
var antiAlias = true

var viewPlaneDistance = 1000.0
var cameraDistanceFactor = 1.0

var imageWidth = 800
var imageHeight = 500
var magnification = 0.5

func main() {
	animation := scn.Animation{
		AnimationName:     animationName,
		Frames:            []scn.Frame{},
		Width:             int(float64(imageWidth) * magnification),
		Height:            int(float64(imageHeight) * magnification),
		WriteRawImageFile: true,
	}

	cornellBoxFile, err := os.Open(cornellBoxFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", cornellBoxFilenamePath, err.Error())
	}

	scale := 100.0
	cornellBox, err := obj.Read(cornellBoxFile)
	cornellBox.ScaleUniform(&vec3.Zero, scale)

	setObjectMaterial(cornellBox, "Wall_right", &color.Color{R: 0.9, G: 0.2, B: 0.2}, nil, false, 0.2, 0.2)
	setObjectMaterial(cornellBox, "Wall_left", &color.Color{R: 0.2, G: 0.2, B: 0.9}, nil, false, 0.2, 0.2)

	setObjectMaterial(cornellBox, "Wall_back", &color.Color{R: 0.7, G: 0.7, B: 0.7}, nil, false, 0.3, 0.2)
	setObjectMaterial(cornellBox, "Ceiling", &color.Color{R: 0.9, G: 0.9, B: 0.9}, nil, false, 0.0, 0.0)
	setObjectMaterial(cornellBox, "Floor_2", &color.Color{R: 1.0, G: 1.0, B: 1.0}, nil, false, 0.5, 0.05)

	setObjectProjection1(cornellBox, "Floor_2")

	lampEmission := color.Color{R: 9.0, G: 9.0, B: 9.0}
	lampColor := color.Color{R: 1.0, G: 1.0, B: 1.0}
	setObjectMaterial(cornellBox, "Lamp_1", &lampColor, &lampEmission, true, 0.0, 0)
	setObjectMaterial(cornellBox, "Lamp_2", &lampColor, &lampEmission, true, 0.0, 0)
	setObjectMaterial(cornellBox, "Lamp_3", &lampColor, &lampEmission, true, 0.0, 0)
	setObjectMaterial(cornellBox, "Lamp_4", &lampColor, &lampEmission, true, 0.0, 0)

	sphere1 := scn.Sphere{
		Name:   "Right sphere",
		Origin: vec3.T{ballRadius + (ballRadius / 2), ballRadius, 0},
		Radius: ballRadius,
		Material: &scn.Material{
			Color:      color.Color{R: 0.9, G: 0.9, B: 0.9},
			Glossiness: 0.1,
			Roughness:  0.2,
			Projection: &scn.ImageProjection{
				ProjectionType: scn.Spherical,
				ImageFilename:  "textures/equirectangular/football.png",
				Origin:         vec3.T{ballRadius + (ballRadius / 2), ballRadius, 0},
				U:              vec3.T{1, 0, 0},
				V:              vec3.T{0, 1, 0},
				RepeatU:        false,
				RepeatV:        false,
				FlipU:          false,
				FlipV:          false,
				Gamma:          0,
			},
		},
	}

	sphere2 := scn.Sphere{
		Name:   "Left sphere",
		Origin: vec3.T{-(ballRadius + (ballRadius / 2)), ballRadius, 0},
		Radius: ballRadius,
		Material: &scn.Material{
			Color: color.Color{R: 0.9, G: 0.9, B: 0.9},
		},
	}

	mirrorSphere1 := scn.Sphere{
		Name:   "Mirror sphere on floor",
		Origin: vec3.T{0, ballRadius / 1.5, ballRadius * 1.5},
		Radius: ballRadius / 1.5,
		Material: &scn.Material{
			Color:      color.Color{R: 0.8, G: 0.85, B: 0.9},
			Roughness:  0.20,
			Glossiness: 0.95,
		},
	}

	mirrorSphereRadius := scale * 0.75

	mirrorSphere2 := scn.Sphere{
		Name:   "Mirror sphere on wall left",
		Origin: vec3.T{-(scale*2 + mirrorSphereRadius*0.25), scale + mirrorSphereRadius, scale*2 + mirrorSphereRadius*0.25},
		Radius: mirrorSphereRadius,
		Material: &scn.Material{
			Color:      color.Color{R: 0.9, G: 0.9, B: 0.95},
			Roughness:  0.02,
			Glossiness: 0.95,
		},
	}

	mirrorSphere3 := scn.Sphere{
		Name:   "Mirror sphere on wall right",
		Origin: vec3.T{scale*2 + mirrorSphereRadius*0.25, scale + mirrorSphereRadius, scale*2 + mirrorSphereRadius*0.25},
		Radius: mirrorSphereRadius,
		Material: &scn.Material{
			Color:      color.Color{R: 0.9, G: 0.9, B: 0.95},
			Roughness:  0.02,
			Glossiness: 0.95,
		},
	}

	scene := scn.SceneNode{
		Spheres:         []*scn.Sphere{&sphere1, &sphere2, &mirrorSphere1, &mirrorSphere2, &mirrorSphere3},
		FacetStructures: []*scn.FacetStructure{cornellBox},
	}

	for animationFrameIndex := 0; animationFrameIndex < amountAnimationFrames; animationFrameIndex++ {
		animationProgress := float64(animationFrameIndex) / float64(amountAnimationFrames)

		cameraOrigin := vec3.T{0, scale * 0.5, -600}
		cameraOrigin.Scale(cameraDistanceFactor)
		focusPoint := vec3.T{0, scale, -scale / 2}

		maxHeightAngle := math.Pi / 14
		heightAngle := maxHeightAngle * (1.0 + math.Cos(animationProgress*2*math.Pi-math.Pi)) / 2.0
		rotationXMatrix := mat3.T{}
		rotationXMatrix.AssignXRotation(heightAngle)
		animatedCameraOrigin := rotationXMatrix.MulVec3(&cameraOrigin)

		maxSideAngle := math.Pi / 8
		sideAngle := maxSideAngle * -1 * math.Sin(animationProgress*2*math.Pi)
		rotationYMatrix := mat3.T{}
		rotationYMatrix.AssignYRotation(sideAngle)
		animatedCameraOrigin = rotationYMatrix.MulVec3(&animatedCameraOrigin)

		camera := getCamera(&animatedCameraOrigin, &focusPoint)

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
