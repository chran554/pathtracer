package main

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"
)

var animationName = "window_test"

var maxRecursionDepth = 8
var amountSamples = 1024 * 12
var lensRadius = 5.0

var viewPlaneDistance = 600.0
var cameraDistanceFactor = 1.0

var imageWidth = 450
var imageHeight = 300
var magnification = 2.0

func main() {
	roomHeight := 40.0
	roomDepth := 170.0
	roomWidth := 60.0
	room := obj.NewBox(obj.BoxCenteredYPositive)
	room.Scale(&vec3.Zero, &vec3.T{roomWidth * 2, roomHeight * 2, roomDepth * 2})
	room.Translate(&vec3.T{0, 0, -roomDepth / 3})
	room.Material = scn.NewMaterial().C(color.Color{R: 0.9, G: 0.8, B: 0.7})

	room.GetObjectsBySubstructureName("xmin")[0].Material = scn.NewMaterial().C(color.NewColor(0.75, 0.25, 0.25))
	room.GetObjectsBySubstructureName("xmax")[0].Material = scn.NewMaterial().C(color.NewColor(0.25, 0.25, 0.75))
	room.UpdateBounds()
	fmt.Printf("Room bounds: %+v\n", room.Bounds)

	fract := 0.001
	windowWidth := 20.0
	windowHeight := 35.0
	windowHeightOverFloor := 32.5
	windowX := roomWidth - fract
	window := createWindow(windowX, windowHeightOverFloor, windowWidth, windowHeight)

	// Diffuse sphere
	sphere1 := scn.NewSphere(&vec3.T{0, 12, -30}, 12, scn.NewMaterial().C(color.NewColor(0.9, 0.8, 0.7)).M(0.0, 1.0))

	// Mirror sphere
	sphere2 := scn.NewSphere(&vec3.T{28, 16, 15}, 16, scn.NewMaterial().C(color.NewColor(0.97, 0.97, 0.843)).M(0.8, 0.0))

	scene := scn.NewSceneNode().
		FS(room, window).
		S(sphere1, sphere2)

	origin := vec3.T{0, 40, -160 - 56}
	origin.Scale(cameraDistanceFactor)
	focusPoint := vec3.T{0, 30, 0}
	camera := scn.NewCamera(&origin, &focusPoint, amountSamples, magnification).
		A(lensRadius, "").
		V(viewPlaneDistance).
		D(maxRecursionDepth)

	frame := scn.NewFrame(animationName, -1, camera, scene)

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, false)
	animation.AddFrame(frame)

	anm.WriteAnimationToFile(animation, false)
}

func createWindow(windowX float64, windowHeightOverFloor float64, windowWidth float64, windowHeight float64) *scn.FacetStructure {
	windowBoardThickness := 1.0
	windowBoardWidth := 2.0

	windowGlass := obj.GetRectangleFacets(&vec3.T{windowX, windowHeightOverFloor, -windowWidth / 2}, &vec3.T{windowX, windowHeightOverFloor, windowWidth / 2}, &vec3.T{windowX, windowHeightOverFloor + windowHeight, windowWidth / 2}, &vec3.T{windowX, windowHeightOverFloor + windowHeight, -windowWidth / 2})

	windowBench := obj.NewBox(obj.BoxCentered)
	windowBench.Scale(&vec3.Zero, &vec3.T{windowBoardWidth * 2, windowBoardWidth / 2, windowWidth / 2})
	windowBench.Translate(&vec3.T{windowX, windowHeightOverFloor, 0})

	windowTopBoard := obj.NewBox(obj.BoxCentered)
	windowTopBoard.Scale(&vec3.Zero, &vec3.T{windowBoardThickness, windowBoardWidth / 2, windowWidth / 2})
	windowTopBoard.Translate(&vec3.T{windowX, windowHeightOverFloor + windowHeight, 0})

	windowLeftBoard := obj.NewBox(obj.BoxCentered)
	windowLeftBoard.Scale(&vec3.Zero, &vec3.T{windowBoardThickness, windowHeight / 2, windowBoardWidth / 2})
	windowLeftBoard.Translate(&vec3.T{windowX, windowHeightOverFloor + windowHeight/2, windowWidth / 2})

	windowRightBoard := obj.NewBox(obj.BoxCentered)
	windowRightBoard.Scale(&vec3.Zero, &vec3.T{windowBoardThickness, windowHeight / 2, windowBoardWidth / 2})
	windowRightBoard.Translate(&vec3.T{windowX, windowHeightOverFloor + windowHeight/2, -windowWidth / 2})

	windowFrameMaterial := scn.NewMaterial().M(0.6, 1.0)
	windowFrame := &scn.FacetStructure{
		SubstructureName: "window frame",
		Material:         windowFrameMaterial,
		FacetStructures:  []*scn.FacetStructure{windowBench, windowTopBoard, windowLeftBoard, windowRightBoard},
	}
	window := &scn.FacetStructure{
		SubstructureName: "Window",
		Facets:           windowGlass,
		FacetStructures:  []*scn.FacetStructure{windowFrame},
		Material:         scn.NewMaterial().C(color.Color{R: 0.75, G: 0.75, B: 1.0}).E(color.White, 24, true),
	}
	return window
}
