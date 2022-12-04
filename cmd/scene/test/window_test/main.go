package main

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
)

var animationName = "window_test"

var renderType = scn.Pathtracing

//var renderType = scn.Raycasting

var maxRecursionDepth = 4
var amountSamples = 1000 * 4
var lensRadius = 10.0

var viewPlaneDistance = 600.0
var cameraDistanceFactor = 1.0

var imageWidth = 450
var imageHeight = 450
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

	boxHeight := 40.0
	boxDepth := 170.0
	boxWidth := 60.0
	box := scn.NewBox(scn.BoxCenteredYPositive)
	box.Scale(&vec3.Zero, &vec3.T{boxWidth * 2, boxHeight * 2, boxDepth * 2})
	box.Translate(&vec3.T{0, 0, -boxDepth / 3})
	box.Material = &scn.Material{
		Color:      (&color.Color{R: 0.9, G: 0.8, B: 0.7}).Multiply(1.0),
		Roughness:  1.0,
		Glossiness: 0.0,
	}
	box.GetFirstObjectBySubstructureName("xmin").Material = &scn.Material{Color: &color.Color{R: 0.75, G: 0.25, B: 0.25}, Roughness: 1.0}
	box.GetFirstObjectBySubstructureName("xmax").Material = &scn.Material{Color: &color.Color{R: 0.25, G: 0.25, B: 0.75}, Roughness: 1.0}
	box.UpdateBounds()
	fmt.Printf("Box bounds: %+v\n", box.Bounds)

	fract := 0.001
	windowWidth := 20.0
	windowHeight := 35.0
	windowHeightOverFloor := 32.5
	windowX := boxWidth - fract
	windowBoardThickness := 1.0
	windowBoardWidth := 2.0

	windowGlass := scn.GetRectangleFacets(&vec3.T{windowX, windowHeightOverFloor, -windowWidth / 2}, &vec3.T{windowX, windowHeightOverFloor, windowWidth / 2}, &vec3.T{windowX, windowHeightOverFloor + windowHeight, windowWidth / 2}, &vec3.T{windowX, windowHeightOverFloor + windowHeight, -windowWidth / 2})

	windowBench := scn.NewBox(scn.BoxCentered)
	windowBench.Scale(&vec3.Zero, &vec3.T{windowBoardWidth * 2, windowBoardWidth / 2, windowWidth / 2})
	windowBench.Translate(&vec3.T{windowX, windowHeightOverFloor, 0})

	windowTopBoard := scn.NewBox(scn.BoxCentered)
	windowTopBoard.Scale(&vec3.Zero, &vec3.T{windowBoardThickness, windowBoardWidth / 2, windowWidth / 2})
	windowTopBoard.Translate(&vec3.T{windowX, windowHeightOverFloor + windowHeight, 0})

	windowLeftBoard := scn.NewBox(scn.BoxCentered)
	windowLeftBoard.Scale(&vec3.Zero, &vec3.T{windowBoardThickness, windowHeight / 2, windowBoardWidth / 2})
	windowLeftBoard.Translate(&vec3.T{windowX, windowHeightOverFloor + windowHeight/2, windowWidth / 2})

	windowRightBoard := scn.NewBox(scn.BoxCentered)
	windowRightBoard.Scale(&vec3.Zero, &vec3.T{windowBoardThickness, windowHeight / 2, windowBoardWidth / 2})
	windowRightBoard.Translate(&vec3.T{windowX, windowHeightOverFloor + windowHeight/2, -windowWidth / 2})

	windowFrameMaterial := scn.Material{Color: &color.White, Roughness: 1.0, Glossiness: 6.0}
	windowFrame := &scn.FacetStructure{
		SubstructureName: "window frame",
		Material:         &windowFrameMaterial,
		FacetStructures:  []*scn.FacetStructure{windowBench, windowTopBoard, windowLeftBoard, windowRightBoard},
	}
	window := scn.FacetStructure{
		SubstructureName: "Window",
		Facets:           windowGlass,
		FacetStructures:  []*scn.FacetStructure{windowFrame},
		Material: &scn.Material{
			Color:         &color.Color{R: 0.75, G: 0.75, B: 1.0},
			Emission:      &color.Color{R: 24.0, G: 24.0, B: 24.0},
			Roughness:     1.0,
			Glossiness:    0.0,
			RayTerminator: true,
		},
	}

	// Diffuse sphere
	sphere1 := &scn.Sphere{
		Origin:   &vec3.T{0, 12, -30},
		Radius:   12,
		Material: &scn.Material{Color: &color.Color{R: 0.9, G: 0.8, B: 0.7}, Roughness: 1.0, Glossiness: 0.0},
	}

	// Mirror sphere
	sphere2 := &scn.Sphere{
		Origin:   &vec3.T{28, 16, 15},
		Radius:   16,
		Material: &scn.Material{Color: &color.Color{R: 0.97, G: 0.97, B: 0.843}, Roughness: 0.0, Glossiness: 0.8},
	}

	scene := scn.SceneNode{
		FacetStructures: []*scn.FacetStructure{box, &window},
		Spheres:         []*scn.Sphere{sphere1, sphere2},
	}

	origin := vec3.T{0, 40, -160 - 56}
	origin.Scale(cameraDistanceFactor)
	focusPoint := vec3.T{0, 40, 0}
	camera := getCamera(origin, focusPoint)

	frame := scn.Frame{
		Filename:   animationName,
		FrameIndex: 0,
		Camera:     &camera,
		SceneNode:  &scene,
	}

	animation := scn.Animation{
		AnimationName:     animationName,
		Frames:            []scn.Frame{frame},
		Width:             width,
		Height:            height,
		WriteRawImageFile: true,
	}

	anm.WriteAnimationToFile(animation, false)
}

func getCamera(origin, focusPoint vec3.T) scn.Camera {
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
