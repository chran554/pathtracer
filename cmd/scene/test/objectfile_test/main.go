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
var cornellBoxFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + cornellBoxFilename

var amountAnimationFrames = 90

var ballRadius float64 = 60

var renderType = scn.Pathtracing
var maxRecursionDepth = 6
var amountSamples = 128 * 4

var viewPlaneDistance = 1000.0
var cameraDistanceFactor = 1.0

var imageWidth = 800
var imageHeight = 500
var magnification = 0.5

func main() {
	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true)

	cornellBoxFile, err := os.Open(cornellBoxFilenamePath)
	if err != nil {
		message := fmt.Sprintf("ouupps, something went wrong loading file: '%s'\n%s\n", cornellBoxFilenamePath, err.Error())
		panic(message)
	}
	defer cornellBoxFile.Close()

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
	setObjectMaterial(cornellBox, "Lamp_1", &lampColor, &lampEmission, true, 0.0, 1.0)
	setObjectMaterial(cornellBox, "Lamp_2", &lampColor, &lampEmission, true, 0.0, 1.0)
	setObjectMaterial(cornellBox, "Lamp_3", &lampColor, &lampEmission, true, 0.0, 1.0)
	setObjectMaterial(cornellBox, "Lamp_4", &lampColor, &lampEmission, true, 0.0, 1.0)

	sphere1Material := scn.NewMaterial().C(color.NewColorGrey(0.9)).M(0.1, 0.2).SP("textures/equirectangular/football.png", &vec3.T{ballRadius + (ballRadius / 2), ballRadius, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})
	sphere1 := scn.NewSphere(&vec3.T{ballRadius + (ballRadius / 2), ballRadius, 0}, ballRadius, sphere1Material).N("Right sphere")

	sphere2Material := scn.NewMaterial().C(color.NewColorGrey(0.9))
	sphere2 := scn.NewSphere(&vec3.T{-(ballRadius + (ballRadius / 2)), ballRadius, 0}, ballRadius, sphere2Material).N("Left sphere")

	mirrorSphereMaterial := scn.NewMaterial().C(color.NewColor(0.8, 0.85, 0.9)).M(0.95, 0.2)
	mirrorSphere1 := scn.NewSphere(&vec3.T{0, ballRadius / 1.5, ballRadius * 1.5}, ballRadius/1.5, mirrorSphereMaterial).N("Mirror sphere on floor")

	mirrorSphereRadius := scale * 0.75

	wallMirrorMaterial := scn.NewMaterial().C(color.NewColor(0.9, 0.9, 0.95)).M(0.95, 0.02)
	mirrorSphere2 := scn.NewSphere(&vec3.T{-(scale*2 + mirrorSphereRadius*0.25), scale + mirrorSphereRadius, scale*2 + mirrorSphereRadius*0.25}, mirrorSphereRadius, wallMirrorMaterial).N("Mirror sphere on wall left")
	mirrorSphere3 := scn.NewSphere(&vec3.T{scale*2 + mirrorSphereRadius*0.25, scale + mirrorSphereRadius, scale*2 + mirrorSphereRadius*0.25}, mirrorSphereRadius, wallMirrorMaterial).N("Mirror sphere on wall right")

	scene := scn.NewSceneNode().S(sphere1, sphere2, mirrorSphere1, mirrorSphere2, mirrorSphere3).FS(cornellBox)

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

		camera := scn.NewCamera(&animatedCameraOrigin, &focusPoint, amountSamples, magnification).V(viewPlaneDistance)

		frame := scn.NewFrame(animation.AnimationName, animationFrameIndex, camera, scene)
		animation.AddFrame(frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

func setObjectMaterial(openBox *scn.FacetStructure, objectName string, color *color.Color, emission *color.Color, rayTerminator bool, glossiness float64, roughness float64) {
	object := openBox.GetFirstObjectByName(objectName)
	if object != nil {
		object.Material.Color = color
		object.Material.Emission = emission
		object.Material.RayTerminator = rayTerminator
		object.Material.Glossiness = glossiness
		object.Material.Roughness = roughness
	} else {
		fmt.Printf("No " + objectName + " object with material found")
		os.Exit(1)
	}
}

func setObjectProjection1(openBox *scn.FacetStructure, objectName string) {
	object := openBox.GetFirstObjectByName(objectName)
	if object != nil {
		projection := scn.NewParallelImageProjection("textures/white_marble.png", &vec3.T{0, 0, 0}, vec3.T{200 * 2, 0, 0}, vec3.T{0, 0, 200 * 2})
		object.Material.Projection = &projection
	} else {
		fmt.Printf("No " + objectName + " found")
		os.Exit(1)
	}
}
