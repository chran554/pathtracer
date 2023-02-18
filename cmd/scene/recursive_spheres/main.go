package main

import (
	"fmt"
	"math"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "recursive_spheres"

var environmentEnvironMap = "textures/equirectangular/nightsky.png"

// var environmentEnvironMap = "textures/equirectangular/sunset horizon 2800x1400.jpg"
var environmentRadius = 100.0 * 1000.0
var environmentEmissionFactor = 1.0

var amountFrames = 1

var imageWidth = 1280
var imageHeight = 1024
var magnification = 1.0

var amountSamples = 128 // 2000

var ballsLightEmissionFactor = 0.01

var startSphereRadius = 150.0
var maxSphereRecursionDepth = 6

var apertureSize = 5.0

var sphereMaterial = scn.NewMaterial().C(color.Color{R: 0.42, G: 0.47, B: 0.42}).
	M(0.85, 0.01).
	E(color.Color{R: 0.85, G: 0.95, B: 0.85}, 0.22, false)

func main() {
	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true)

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		recursiveBalls := getRecursiveBalls(startSphereRadius, maxSphereRecursionDepth)
		ballsBounds := recursiveBalls.UpdateBounds()
		recursiveBalls.Translate(&vec3.T{0, -ballsBounds.Ymin, 0})
		ballsBounds = recursiveBalls.UpdateBounds()
		fmt.Printf("Balls bounds: %+v   (center: %+v)\n", ballsBounds, ballsBounds.Center())

		animationAngle := animationProgress * (math.Pi / 2.0)
		recursiveBalls.RotateY(&vec3.Zero, animationAngle)

		recursiveBalls.RotateX(&vec3.Zero, math.Pi/12)
		recursiveBalls.RotateY(&vec3.Zero, math.Pi/8)

		ballsLightDistanceFactor := 400.0
		ballsLightPosition := ballsBounds.Center().Add(&vec3.T{-ballsLightDistanceFactor, ballsLightDistanceFactor, -2.0 * ballsLightDistanceFactor})
		ballsLightMaterial := scn.NewMaterial().E(color.Color{R: 15, G: 14.0, B: 12.0}, ballsLightEmissionFactor, true)
		ballsLight := scn.NewSphere(ballsLightPosition, 200, ballsLightMaterial).N("Balls light")

		// Sky dome

		environmentOrigin := vec3.T{0, 0, 0}
		environmentMaterial := scn.NewMaterial().
			E(color.White, environmentEmissionFactor, true).
			SP(environmentEnvironMap, &environmentOrigin, vec3.T{-0.2, 0, -1}, vec3.T{0, 1, 0})
		environmentSphere := scn.NewSphere(&environmentOrigin, environmentRadius, environmentMaterial).N("Environment mapping")

		// Ground

		groundMaterial := scn.NewMaterial().N("Ground material").PP("textures/ground/soil-cracked.png", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(250*2), vec3.UnitZ.Scaled(250*2))
		ground := scn.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, environmentRadius, groundMaterial).N("Ground")

		cameraOrigin := ballsBounds.Center().Add(&vec3.T{0, ballsBounds.SizeY() * 1.5 / 10.0, -800})
		cameraFocusPoint := ballsBounds.Center().Add(&vec3.T{0, ballsBounds.SizeY() / 10.0, -ballsBounds.SizeZ() / 2.0 * 0.8})
		camera := scn.NewCamera(cameraOrigin, cameraFocusPoint, amountSamples, magnification).A(apertureSize, "")

		scene := scn.NewSceneNode().
			S(environmentSphere, ballsLight).
			D(ground).
			SN(recursiveBalls)

		frame := scn.NewFrame(animationName, frameIndex, camera, scene)

		animation.Frames = append(animation.Frames, frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

func getRecursiveBalls(middleSphereRadius float64, maxRecursionDepth int) *scn.SceneNode {
	scene := scn.SceneNode{}

	origin := vec3.T{0, 0, 0}
	middleSphere := scn.NewSphere(&origin, middleSphereRadius, sphereMaterial).N("0")
	scene.Spheres = append(scene.Spheres, middleSphere)
	_getRecursiveBalls(middleSphere, maxRecursionDepth, 0, &scene)

	return &scene
}

func _getRecursiveBalls(parentSphere *scn.Sphere, maxRecursionDepth int, takenSide int, scene *scn.SceneNode) {
	var sceneSubNode scn.SceneNode

	if parentSphere.Radius < 5.0 || maxRecursionDepth == 0 {
		return
	}

	childRadius := parentSphere.Radius * 0.48
	childOffset := parentSphere.Radius + childRadius*1.05

	if takenSide != 2 { // offset in negative x
		childOrigin := parentSphere.Origin.Added(&vec3.T{-childOffset, 0, 0})
		sphere := scn.NewSphere(&childOrigin, childRadius, sphereMaterial).N(parentSphere.Name + " -x")
		sceneSubNode.Spheres = append(sceneSubNode.Spheres, sphere)
		_getRecursiveBalls(sphere, maxRecursionDepth-1, 1, &sceneSubNode)
	}

	if takenSide != 1 { // offset in positive x
		childOrigin := parentSphere.Origin.Added(&vec3.T{childOffset, 0, 0})
		sphere := scn.NewSphere(&childOrigin, childRadius, sphereMaterial).N(parentSphere.Name + " +x")
		sceneSubNode.Spheres = append(sceneSubNode.Spheres, sphere)
		_getRecursiveBalls(sphere, maxRecursionDepth-1, 2, &sceneSubNode)
	}

	if takenSide != 4 { // offset in negative y
		childOrigin := parentSphere.Origin.Added(&vec3.T{0, -childOffset, 0})
		sphere := scn.NewSphere(&childOrigin, childRadius, sphereMaterial).N(parentSphere.Name + " -y")
		sceneSubNode.Spheres = append(sceneSubNode.Spheres, sphere)
		_getRecursiveBalls(sphere, maxRecursionDepth-1, 3, &sceneSubNode)
	}

	if takenSide != 3 { // offset in positive y
		childOrigin := parentSphere.Origin.Added(&vec3.T{0, childOffset, 0})
		sphere := scn.NewSphere(&childOrigin, childRadius, sphereMaterial).N(parentSphere.Name + " +y")
		sceneSubNode.Spheres = append(sceneSubNode.Spheres, sphere)
		_getRecursiveBalls(sphere, maxRecursionDepth-1, 4, &sceneSubNode)
	}

	if takenSide != 6 { // offset in negative z
		childOrigin := parentSphere.Origin.Added(&vec3.T{0, 0, -childOffset})
		sphere := scn.NewSphere(&childOrigin, childRadius, sphereMaterial).N(parentSphere.Name + " -z")
		sceneSubNode.Spheres = append(sceneSubNode.Spheres, sphere)
		_getRecursiveBalls(sphere, maxRecursionDepth-1, 5, &sceneSubNode)
	}

	if takenSide != 5 { // offset in positive z
		childOrigin := parentSphere.Origin.Added(&vec3.T{0, 0, childOffset})
		sphere := scn.NewSphere(&childOrigin, childRadius, sphereMaterial).N(parentSphere.Name + " +z")
		sceneSubNode.Spheres = append(sceneSubNode.Spheres, sphere)
		_getRecursiveBalls(sphere, maxRecursionDepth-1, 6, &sceneSubNode)
	}

	scene.ChildNodes = append(scene.ChildNodes, &sceneSubNode)
}
