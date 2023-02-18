package main

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"os"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"
)

var animationName = "gopher"

var amountFrames = 1

var amountSamples = 1024 * 4 * 3

var imageWidth = 400
var imageHeight = 500
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

	// Sky
	groundMaterial := scn.NewMaterial().PP("textures/floor/Calacatta-Vena-French-Pattern-Architextures.jpg", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(150), vec3.UnitZ.Scaled(150))
	ground := &scn.Disc{Name: "ground", Origin: &vec3.T{0, 0, 0}, Normal: &vec3.UnitY, Radius: 5000.0, Material: groundMaterial}

	// Sky
	skyMaterial := scn.NewMaterial().
		E(color.White, 0.5, true).
		SP("textures/equirectangular/wirebox 6192x3098.png", &vec3.T{0, 0, 0}, vec3.UnitX, vec3.UnitY)
	skyDome := scn.NewSphere(&vec3.T{0, 0, 0}, 5000, skyMaterial).N("sky dome")

	// Gopher
	gopher := GetGopher(&vec3.T{1, 1, 1})
	gopher.Translate(&vec3.T{0, -gopher.Bounds.Ymin, 0})
	gopher.ScaleUniform(&vec3.Zero, 2.0)
	gopher.RotateY(&vec3.Zero, math.Pi*5.0/6.0)
	gopher.Translate(&vec3.T{0, 0, 0})
	gopher.UpdateBounds()

	gopherLightMaterial := scn.NewMaterial().E(color.Color{R: 6.0, G: 5.3, B: 4.5}, 20, true)
	gopherLight := scn.NewSphere(&vec3.T{-150, 250, -175}, 15.0, gopherLightMaterial).N("Gopher light")

	scene := scn.NewSceneNode().
		S(gopherLight, skyDome).
		D(ground).
		FS(gopher)

	animation := scn.NewAnimation(animationName, width, height, magnification, false)

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		camera := getCamera(animationProgress)
		frame := scn.NewFrame(animationName, -1, camera, scene)
		animation.AddFrame(frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

func getCamera(animationProgress float64) *scn.Camera {
	cameraOrigin := &vec3.T{0, 200, -800}
	focusPoint := &vec3.T{0, 150, 0}

	// Animation
	angle := (math.Pi / 2.0) * animationProgress
	scn.RotateY(cameraOrigin, &vec3.Zero, angle)
	scn.RotateY(focusPoint, &vec3.Zero, angle)

	heading := focusPoint.Subed(cameraOrigin)
	focusDistance := heading.Length() * 1.75

	return scn.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).F(focusDistance)
}

func GetGopher(scale *vec3.T) *scn.FacetStructure {
	var objFilename = "go_gopher_color.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objFilename

	objFile, err := os.Open(objFilenamePath)
	if err != nil {
		message := fmt.Sprintf("ouupps, something went wrong loading file: '%s'\n%s\n", objFilenamePath, err.Error())
		panic(message)
	}
	defer objFile.Close()

	object, err := obj.Read(objFile)
	object.Scale(&vec3.Zero, scale)
	// obj.ClearMaterials()
	object.UpdateBounds()
	fmt.Printf("Gopher bounds: %+v\n", object.Bounds)

	return object
}
