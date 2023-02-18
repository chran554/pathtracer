package main

import (
	"math"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "tessellated_sphere"

var environmentEnvironMap = "textures/equirectangular/sunset horizon 2800x1400.jpg"
var environmentRadius = 500.0 * 1000.0
var environmentEmissionFactor = 1.0

var amountFrames = 180

var imageWidth = 300
var imageHeight = 300
var magnification = 3.0

var amountSamples = 1000 * 24
var maxRecursion = 2

var apertureSize = 1.5

func main() {
	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false)

	// Sphere
	tessellatedSphere := obj.NewTessellatedSphere(5, false)

	tessellatedSphere.Translate(&vec3.T{0, -tessellatedSphere.Bounds.Ymin, 0})
	tessellatedSphere.ScaleUniform(&vec3.Zero, 30.0)
	tessellatedSphereBounds := tessellatedSphere.UpdateBounds()
	tessellatedSphere.Material = scn.NewMaterial().N("tessellated sphere").C(color.White)

	// Sky dome
	skyDomeOrigin := vec3.T{0, 0, 0}
	skyDomeMaterial := scn.NewMaterial().
		E(color.White, environmentEmissionFactor, true).
		SP(environmentEnvironMap, &skyDomeOrigin, vec3.T{-0.2, 0, -1}, vec3.T{0, 1, 0})
	skyDome := scn.NewSphere(&skyDomeOrigin, environmentRadius, skyDomeMaterial).N("sky dome")

	// Ground
	groundMaterial := scn.NewMaterial().N("Ground material").PP("textures/ground/soil-cracked.png", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(150), vec3.UnitZ.Scaled(150))
	ground := scn.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, environmentRadius, groundMaterial).N("Ground")

	// Camera
	cameraOrigin := tessellatedSphereBounds.Center().Add(&vec3.T{25, 25, -250})
	cameraFocusPoint := tessellatedSphereBounds.Center().Add(&vec3.T{0, 0, -(tessellatedSphereBounds.SizeZ() / 2) * 0.9})
	camera := scn.NewCamera(cameraOrigin, cameraFocusPoint, amountSamples, magnification).D(maxRecursion).A(apertureSize, "")

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		// Lamp
		lampMaterial := scn.NewMaterial().N("lamp").E(color.White, 125.0, true)
		lamp := scn.NewSphere(&vec3.T{50, 50, -35}, 4, lampMaterial).N("lamp")
		lamp.RotateY(tessellatedSphere.Bounds.Center(), -math.Pi/4+animationProgress*math.Pi/2)

		scene := scn.NewSceneNode().S(skyDome, lamp).D(ground).FS(tessellatedSphere)
		frame := scn.NewFrame(animation.AnimationName, frameIndex, camera, scene)
		animation.AddFrame(frame)
	}

	anm.WriteAnimationToFile(animation, false)
}
