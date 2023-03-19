package main

import (
	"math"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "tessellated_sphere_transform"

var environmentEnvironMap = "textures/equirectangular/sunset horizon 2800x1400.jpg"
var environmentRadius = 500.0 * 1000.0
var environmentEmissionFactor = 1.0

var amountFrames = 180

var imageWidth = 200
var imageHeight = 250
var magnification = 2.0

var amountSamples = 512 * 2 * 8
var maxRecursion = 5

var apertureSize = 1.5

func main() {
	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	// Sphere
	tessellatedSphere := obj.NewTessellatedSphere(5, false)
	tessellatedSphere.Material = scn.NewMaterial().N("tessellated sphere").M(0.1, 0.1)

	objectHeight := 60.0
	objectWidth := 30.0
	objectDepth := 6.0
	amountTwistTurns := 1.5

	tessellatedSphere.Translate(&vec3.T{0, -tessellatedSphere.Bounds.Ymin, 0}) // diameter 2 units
	tessellatedSphere.Scale(&vec3.Zero, &vec3.T{objectWidth / 2.0, objectHeight / 2.0, objectDepth / 2.0})
	tessellatedSphere.TwistY(&vec3.Zero, amountTwistTurns*(math.Pi*2)/objectHeight)
	tessellatedSphere.UpdateVertexNormals(false)
	tessellatedSphereBounds := tessellatedSphere.UpdateBounds()

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
	cameraOrigin := tessellatedSphereBounds.Center().Add(&vec3.T{25, 10, -250})
	cameraFocusPoint := tessellatedSphereBounds.Center().Add(&vec3.T{0, 0, -(tessellatedSphereBounds.SizeZ() / 2) * 0.8})
	camera := scn.NewCamera(cameraOrigin, cameraFocusPoint, amountSamples, magnification).D(maxRecursion).A(apertureSize, "")

	//for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
	for frameIndex := 88; frameIndex < 89; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		// Lamp
		lampMaterial := scn.NewMaterial().N("lamp").E(color.White, 50.0, true)
		lamp := scn.NewSphere(&vec3.T{300, 150, -150}, 50, lampMaterial).N("lamp")
		lamp.RotateY(tessellatedSphere.Bounds.Center(), -math.Pi/4+animationProgress*math.Pi/2)

		scene := scn.NewSceneNode().S(skyDome, lamp).D(ground).FS(tessellatedSphere)
		frame := scn.NewFrame(animation.AnimationName, frameIndex, camera, scene)
		animation.AddFrame(frame)
	}

	anm.WriteAnimationToFile(animation, false)
}
