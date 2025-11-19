package main

import (
	"fmt"
	"math"
	"math/rand"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

// "Lonesome Gopher in love" scene

var environmentEnvironMap = "textures/equirectangular/sunset horizon 2800x1400.jpg"
var environmentRadius = 100.0 * 1000.0
var environmentEmissionFactor = 0.15

var animationName = "aperture_shape_test"
var amountFrames = 1

var imageWidth = 1280
var imageHeight = 1024
var magnification = 1.0

var amountSamples = 512 * 2 * 10

var gopherLightEmissionFactor = 0.0 // 75.0

var lampEmissionFactor = 1.5

var amountSpheres = 1000
var sphereRadius = 250.0

var sphereMinDistance = -1500.0
var sphereMaxDistance = environmentRadius * 0.3

var apertureSize = 20.0 //500.0

func main() {
	// Environment sphere

	skyDomeOrigin := vec3.T{0, 0, 0}
	skyDomeMaterial := scene.NewMaterial().
		E(color.White, environmentEmissionFactor, true).
		SP(floatimage.Load(environmentEnvironMap), &skyDomeOrigin, vec3.T{0, 0, 1}, vec3.T{0, 1, 0})
	skyDome := scene.NewSphere(&skyDomeOrigin, environmentRadius, skyDomeMaterial).N("sky dome")

	// Gopher

	gopher := obj.NewGopher(600)
	gopher.RotateY(&vec3.Zero, math.Pi*7.0/8.0)
	gopher.UpdateBounds()
	gopher.Translate(&vec3.T{0, -gopher.Bounds.Ymin, -gopher.Bounds.Zmin * 0.8})
	gopher.UpdateBounds()

	gopherLightPosition := vec3.T{gopher.Bounds.Center()[0] - 350, gopher.Bounds.Center()[1] + 350, gopher.Bounds.Center()[2] - 700}
	gopherLightMaterial := scene.NewMaterial().E(color.NewColorKelvin(5000), gopherLightEmissionFactor, true)
	gopherLight := scene.NewSphere(&gopherLightPosition, 80, gopherLightMaterial).N("Gopher light")

	// Ground

	groundProjection := scene.NewParallelImageProjection(floatimage.Load("textures/ground/soil-cracked.png"), &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(gopher.Bounds.SizeY()*2), vec3.UnitZ.Scaled(gopher.Bounds.SizeY()*2))
	groundMaterial := scene.NewMaterial().N("Ground material").P(&groundProjection)
	ground := scene.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, environmentRadius, groundMaterial).N("Ground")

	// Spheres

	childSceneNode := scene.NewSceneNode()
	for sphereIndex := 0; sphereIndex < amountSpheres; sphereIndex++ {
		sphere := getSphere(sphereRadius, sphereMinDistance, sphereMaxDistance)
		childSceneNode.S(sphere)
	}

	scn := scene.NewSceneNode().
		S(skyDome, gopherLight).
		FS(gopher).
		D(ground).
		SN(childSceneNode)

	startAngle := math.Pi / 2.0
	xPosMax := 250.0
	yPosMax := gopher.Bounds.SizeY() * 0.75
	yPosMin := gopher.Bounds.SizeY() * 0.15

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		focusPoint := vec3.T{0, gopher.Bounds.SizeY() * 0.75, 0}

		angle := animationProgress * 2.0 * math.Pi
		xPos := xPosMax * math.Cos(angle+startAngle)
		yPos := yPosMin + (yPosMax-yPosMin)*(math.Sin(angle+startAngle)+1.0)/2.0
		cameraOrigin := vec3.T{xPos, yPos, -600}

		camera := scene.NewCamera(&cameraOrigin, &focusPoint, amountSamples, magnification).
			A(apertureSize, floatimage.Load("textures/aperture/heart.png")).
			V(600)

		frame := scene.NewFrame(animationName, frameIndex, camera, scn)

		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func getSphere(radius float64, minDistance, maxDistance float64) *scene.Sphere {
	r := 0.50 + rand.Float64()*0.50
	g := 0.35 + rand.Float64()*0.45
	b := 0.35 + rand.Float64()*0.45

	sphereColor := color.NewColor(r, g, b)

	sphereMaterial := scene.NewMaterial().
		C(sphereColor).
		E(sphereColor, lampEmissionFactor, true)

	x := (rand.Float64() - 0.5) * 2 * (maxDistance * 2 / 2.0)
	y := math.Pow(rand.Float64(), 2.0) * (maxDistance * 2 / 2.0)
	z := minDistance + rand.Float64()*(maxDistance-minDistance)

	origin := vec3.T{x, y, z}

	return scene.NewSphere(&origin, radius, sphereMaterial)
}
