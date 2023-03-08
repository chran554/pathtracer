package main

import (
	"fmt"
	"math"
	"os"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "experiment"

var environmentRadius = 500.0 * 1000.0
var environmentEmissionFactor = 1.0

var amountFrames = 1

var imageWidth = 150
var imageHeight = 100
var magnification = 5.0

var renderType = scn.Pathtracing
var amountSamples = 256 * 2 * 6 * 4 // 2048 // 4000 //* 8
var maxRecursion = 3

var smoothShadeFacets = true
var viewPlaneDistance = 800.0
var apertureSize = 0.0 // 0.75 // 2.0

func main() {

	/*
		object := obj.NewDragon02(100, true, true)
		pillar := object.GetFirstObjectByName("pillar")
		pillarCenter := *pillar.Bounds.Center()
		object.ReplaceMaterial("pillar", scn.NewMaterial().N("pillar").SP("textures/marble/white_marble_double_width.png", pillarCenter, vec3.UnitX, vec3.UnitY).M(0.2, 0.7))

		object.RotateY(&vec3.Zero, -math.Pi/8)
		object.UpdateBounds()
		objectBounds := object.Bounds
		//object.Material.T(0.95, true, scn.RefractionIndex_Glass)
		//object.UpdateVertexNormals(false)
		fmt.Printf("object bounds: %+v\n", objectBounds)

		// external light
		lamp := createLamp("dragon lamp", 70.0, 10.0, objectBounds.Center().Add(&vec3.T{-50, 200, -200}), color.Color{R: 0.9, G: 0.85, B: 0.8})
		lamp.RotateY(object.Bounds.Center(), -math.Pi/4)
	*/

	object := NewCorner(50)
	objectBounds := object.Bounds
	lamp := createLamp("corner lamp", 20.0, 300.0, &vec3.T{-175, 200, -200}, color.Color{R: 0.9, G: 0.85, B: 0.8})
	cameraHeight := 25.0
	cameraDistance := 550.0

	/*
		object := obj.NewCastle(&vec3.T{80, 80, 80})
		objectBounds := object.Bounds
		lampColor := color.Color{R: 1.0, G: 0.87, B: 0.5}
		castleLamp := createLamp("castle_lamp", 6.5, 15.0, objectBounds.Center().Add(&vec3.T{13, -8, -2}), lampColor)
		entranceLamp1 := createLamp("entrance_lamp_1", 3.0, 6.0, objectBounds.Center().Add(&vec3.T{0, -21, -3}), lampColor)
		entranceLamp2 := createLamp("entrance_lamp_2", 3.0, 4.0, objectBounds.Center().Add(&vec3.T{-15, -25, -3}), lampColor)
		entranceLamp3 := createLamp("entrance_lamp_3", 3.0, 4.0, objectBounds.Center().Add(&vec3.T{15, -25, -3}), lampColor)
		towerLamp1 := createLamp("tower_lamp_1", 1.3, 5.0, objectBounds.Center().Add(&vec3.T{-10.3, 1, -13}), lampColor)
		towerLamp2 := createLamp("tower_lamp_2", 1.3, 5.0, objectBounds.Center().Add(&vec3.T{10.3, 1, -13}), lampColor)
		tinnerLamp1 := createLamp("tinner_lamp_1", 3.0, 5.0, objectBounds.Center().Add(&vec3.T{10, 25, 6}), lampColor)
		var lamps *scn.Sphere{castleLamp, entranceLamp1, entranceLamp2, entranceLamp3, towerLamp1, towerLamp2, tinnerLamp1}
	*/

	// Sky dome
	// var environmentEnvironMap =
	environmentSphere := addEnvironmentMapping("textures/equirectangular/sunset horizon 2800x1400.jpg")
	// environmentSphere := addEnvironmentMapping("textures/equirectangular/nightsky.png")

	// Ground
	groundProjection := scn.NewParallelImageProjection("textures/ground/grass_short.png", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(80/2), vec3.UnitZ.Scaled(50/2))
	groundMaterial := scn.NewMaterial().N("Ground material").P(&groundProjection)
	ground := scn.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, environmentRadius, groundMaterial).N("Ground")

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false)

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		cameraStartAngle := -math.Pi / 2

		cameraOffset := &vec3.T{}
		cameraOffset[0] = math.Cos(animationProgress*2*math.Pi+cameraStartAngle) * cameraDistance
		cameraOffset[1] = cameraHeight
		cameraOffset[2] = math.Sin(animationProgress*2*math.Pi+cameraStartAngle) * cameraDistance

		focusPointOffset := cameraOffset.Normalized()
		focusPointOffset.Scale(60.0) // Focus point is 15 units from object center towards camera point
		// focusPointOffset.Add(&vec3.T{0, -10, 0}) // For castle, set focus point a bit lower than center of object

		cameraOrigin := objectBounds.Center().Add(cameraOffset)
		cameraFocusPoint := objectBounds.Center().Add(&focusPointOffset)

		xzRadius2 := math.Sqrt(cameraFocusPoint[0]*cameraFocusPoint[0] + cameraFocusPoint[2]*cameraFocusPoint[2])
		cameraFocusPoint[0] = math.Cos(animationProgress*2*math.Pi+cameraStartAngle) * xzRadius2
		cameraFocusPoint[2] = math.Sin(animationProgress*2*math.Pi+cameraStartAngle) * xzRadius2

		camera := scn.NewCamera(cameraOrigin, cameraFocusPoint, amountSamples, magnification).D(maxRecursion).A(apertureSize, "")

		scene := scn.NewSceneNode().
			S(environmentSphere, lamp).
			D(ground).
			FS(object)

		frame := scn.NewFrame(animationName, frameIndex, camera, scene)

		animation.AddFrame(frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

func createLamp(lampName string, size float64, strength float64, lampPosition *vec3.T, lampColor color.Color) *scn.Sphere {
	return &scn.Sphere{
		Name:     lampName,
		Origin:   lampPosition,
		Radius:   size,
		Material: scn.NewMaterial().N(lampName).C(color.White).E(lampColor, strength, true),
	}
}

func NewCorner(scale float64) *scn.FacetStructure {
	var facets []*scn.Facet

	p000 := &vec3.T{0, 0, 0}
	p101 := &vec3.T{1, 0, 1}
	p111 := &vec3.T{1, 1, 1}
	p010 := &vec3.T{0, 1, 0}
	p_01 := &vec3.T{-1, 0, 1}
	p_11 := &vec3.T{-1, 1, 1}

	facets = append(facets, &scn.Facet{Vertices: []*vec3.T{p000, p101, p111}})
	facets = append(facets, &scn.Facet{Vertices: []*vec3.T{p000, p111, p010}})
	facets = append(facets, &scn.Facet{Vertices: []*vec3.T{p000, p_11, p_01}})
	facets = append(facets, &scn.Facet{Vertices: []*vec3.T{p000, p010, p_11}})

	object := &scn.FacetStructure{Name: "experiment", Facets: facets}
	object.Material = scn.NewMaterial().N("experiment").C(color.White)

	object.ScaleUniform(&vec3.Zero, scale)
	object.UpdateBounds()
	object.ChangeWindingOrder()
	object.UpdateNormals()

	if smoothShadeFacets {
		object.UpdateVertexNormals(false)
	}

	return object
}

func addEnvironmentMapping(filename string) *scn.Sphere {
	origin := vec3.T{0, 0, 0}
	u := vec3.T{-0.2, 0, -1}
	v := vec3.T{0, 1, 0}
	material := scn.NewMaterial().E(color.White, environmentEmissionFactor, true).SP(filename, &origin, u, v)
	sphere := scn.NewSphere(&origin, environmentRadius, material).N("Environment mapping")

	return sphere
}

func createFile(name string) *os.File {
	objFile, err := os.Create(name)
	if err != nil {
		fmt.Printf("could not create file: '%s'\n%s\n", objFile.Name(), err.Error())
		os.Exit(1)
	}
	return objFile
}
