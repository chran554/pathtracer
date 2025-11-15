package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	anm "pathtracer/internal/pkg/renderfile"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "lightbox"

var environmentRadius = 500.0 * 1000.0
var environmentEmissionFactor = 1.0

var amountFrames = 180

var imageWidth = 300
var imageHeight = 300
var magnification = 1.5

var amountSamples = 512 * 4
var maxRecursion = 3

var apertureSize = 1.5

func main() {
	var textureEnvironment = floatimage.Load("textures/equirectangular/sunset horizon 2800x1400.jpg")
	var textureSoilCracked = floatimage.Load("textures/ground/soil-cracked.png")
	var textureLightBox = floatimage.Load("textures/lights/lightboxtexture_2.0.png")

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	// Sky dome
	skyDomeOrigin := vec3.T{0, 0, 0}
	skyDomeMaterial := scn.NewMaterial().
		E(color.White, environmentEmissionFactor, true).
		SP(textureEnvironment, &skyDomeOrigin, vec3.T{-0.2, 0, -1}, vec3.T{0, 1, 0})
	skyDome := scn.NewSphere(&skyDomeOrigin, environmentRadius, skyDomeMaterial).N("sky dome")

	// Ground
	groundMaterial := scn.NewMaterial().N("Ground material").PP(textureSoilCracked, &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(150), vec3.UnitZ.Scaled(150))
	ground := scn.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, environmentRadius, groundMaterial).N("Ground")

	// Camera
	cameraOrigin := vec3.Zero.Added(&vec3.T{0, 60, -250})
	cameraFocusPoint := &vec3.T{0, 30, -10}
	camera := scn.NewCamera(&cameraOrigin, cameraFocusPoint, amountSamples, magnification).D(maxRecursion).A(apertureSize, nil)

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		lamp, lampMaterial := obj.NewBoxWithEmission(obj.BoxCenteredYPositive, color.White, 40, textureLightBox)
		lamp.Material.Name = "lamp 1 material"
		lamp.ScaleUniform(&vec3.Zero, 30)
		lamp.Translate(&vec3.T{0, 15, 0})
		lamp.Material = lampMaterial
		lamp.RotateY(lamp.Bounds.Center(), 2*math.Pi*animationProgress)
		lamp.RotateX(lamp.Bounds.Center(), 2*math.Pi*animationProgress*2)

		scene := scn.NewSceneNode().S(skyDome).D(ground).FS(lamp)
		frame := scn.NewFrame(animation.AnimationName, frameIndex, camera, scene)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := anm.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
