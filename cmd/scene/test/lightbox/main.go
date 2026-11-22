package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

const (
	animationName = "lightbox"

	environmentRadius         = 500.0 * 1000.0
	environmentEmissionFactor = 1.0

	amountFrames = 180

	imageWidth    = 300
	imageHeight   = 300
	magnification = 1.5

	amountSamples = 512 * 4
	maxRecursion  = 3

	apertureSize = 1.5
)

func main() {
	var textureEnvironment = scene.NewTextureReferenceOrPanic("textures/equirectangular/sunset horizon 2800x1400.jpg", scene.TextureTypeAlbedo)
	var textureSoilCracked = scene.NewTextureReferenceOrPanic("textures/ground/soil-cracked.png", scene.TextureTypeAlbedo)
	var textureLightBox = floatimage.LoadOrPanic("textures/lights/lightboxtexture_2.0.png")

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	// Sky dome
	skyDomeOrigin := vec3.T{0, 0, 0}
	skyDomeMaterial := scene.NewMaterial().
		E(color.White, environmentEmissionFactor, true).
		SP(textureEnvironment, &skyDomeOrigin, vec3.T{-0.2, 0, -1}, vec3.T{0, 1, 0})
	skyDome := scene.NewSphere(&skyDomeOrigin, environmentRadius, skyDomeMaterial).N("sky dome")

	// Ground
	groundMaterial := scene.NewMaterial().N("Ground material").PP(textureSoilCracked, &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(150), vec3.UnitZ.Scaled(150))
	ground := scene.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, environmentRadius, groundMaterial).N("Ground")

	// Camera
	cameraOrigin := vec3.Zero.Added(&vec3.T{0, 60, -250})
	cameraFocusPoint := &vec3.T{0, 30, -10}
	camera := scene.NewCamera(&cameraOrigin, cameraFocusPoint, amountSamples, magnification).D(maxRecursion).A(apertureSize, nil)

	for frameIndex := range amountFrames {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		lamp, lampMaterial := obj.NewBoxWithEmission(obj.BoxCenteredYPositive, color.White, 40, textureLightBox)
		lamp.Material.Name = "lamp 1 material"
		lamp.ScaleUniform(&vec3.Zero, 30)
		lamp.Translate(&vec3.T{0, 15, 0})
		lamp.Material = lampMaterial
		lamp.RotateY(lamp.Bounds.Center(), 2*math.Pi*animationProgress)
		lamp.RotateX(lamp.Bounds.Center(), 2*math.Pi*animationProgress*2)

		scn := scene.NewSceneNode().S(skyDome).D(ground).FS(lamp)
		frame := scene.NewFrame(animation.AnimationName, frameIndex, camera, scn)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
