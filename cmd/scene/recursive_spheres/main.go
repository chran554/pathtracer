package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "recursive_spheres"

// var environmentEnvironMap = "textures/equirectangular/sunset horizon 2800x1400.jpg"
var skyDomeRadius = 100.0 * 1000.0

//var environmentEnvironMap = "textures/equirectangular/nightsky.png"

var skyDomeEmissionFactor = 1.0
var amountFrames = 1

var imageWidth = 1280

var imageHeight = 1024
var magnification = 1.0
var amountSamples = 1024 * 3

var startSphereRadius = 150.0

var maxSphereRecursionDepth = 6
var apertureSize = 3.0

var sphereMaterial = scene.NewMaterial().C(color.NewColorGrey(0.8)).M(0.70, 0.07)

func main() {
	environmentEnvironMap := floatimage.LoadOrPanic("textures/equirectangular/open_grassfield_sunny_day.jpg")
	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, false)

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

		// Sky dome
		skyDomeOrigin := vec3.T{0, 0, 0}
		skyDomeMaterial := scene.NewMaterial().
			E(color.White, skyDomeEmissionFactor, true).
			SP(environmentEnvironMap, &skyDomeOrigin, vec3.T{-0.75, 0, -0.25}, vec3.T{0, 1, 0})
		skyDome := scene.NewSphere(&skyDomeOrigin, skyDomeRadius, skyDomeMaterial).N("Environment mapping")

		cameraOrigin := ballsBounds.Center().Add(&vec3.T{0, ballsBounds.SizeY() * 1.5 / 10.0, -800})
		cameraFocusPoint := ballsBounds.Center().Add(&vec3.T{0, ballsBounds.SizeY() / 10.0, -ballsBounds.SizeZ() / 2.0 * 0.8})
		camera := scene.NewCamera(cameraOrigin, cameraFocusPoint, amountSamples, magnification).A(apertureSize, nil).V(700)

		scn := scene.NewSceneNode().S(skyDome).SN(recursiveBalls)

		frame := scene.NewFrame(animationName, frameIndex, camera, scn)

		animation.Frames = append(animation.Frames, frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func getRecursiveBalls(middleSphereRadius float64, maxRecursionDepth int) *scene.SceneNode {
	scn := scene.SceneNode{}

	origin := vec3.T{0, 0, 0}
	middleSphere := scene.NewSphere(&origin, middleSphereRadius, sphereMaterial).N("0")
	scn.Spheres = append(scn.Spheres, middleSphere)
	_getRecursiveBalls(middleSphere, maxRecursionDepth, 0, &scn)

	return &scn
}

func _getRecursiveBalls(parentSphere *scene.Sphere, maxRecursionDepth int, takenSide int, scn *scene.SceneNode) {
	var sceneSubNode scene.SceneNode

	if parentSphere.Radius < 5.0 || maxRecursionDepth == 0 {
		return
	}

	childRadius := parentSphere.Radius * 0.48
	childOffset := parentSphere.Radius + childRadius*1.05

	if takenSide != 2 { // offset in negative x
		childOrigin := parentSphere.Origin.Added(&vec3.T{-childOffset, 0, 0})
		sphere := scene.NewSphere(&childOrigin, childRadius, sphereMaterial).N(parentSphere.Name + " -x")
		sceneSubNode.Spheres = append(sceneSubNode.Spheres, sphere)
		_getRecursiveBalls(sphere, maxRecursionDepth-1, 1, &sceneSubNode)
	}

	if takenSide != 1 { // offset in positive x
		childOrigin := parentSphere.Origin.Added(&vec3.T{childOffset, 0, 0})
		sphere := scene.NewSphere(&childOrigin, childRadius, sphereMaterial).N(parentSphere.Name + " +x")
		sceneSubNode.Spheres = append(sceneSubNode.Spheres, sphere)
		_getRecursiveBalls(sphere, maxRecursionDepth-1, 2, &sceneSubNode)
	}

	if takenSide != 4 { // offset in negative y
		childOrigin := parentSphere.Origin.Added(&vec3.T{0, -childOffset, 0})
		sphere := scene.NewSphere(&childOrigin, childRadius, sphereMaterial).N(parentSphere.Name + " -y")
		sceneSubNode.Spheres = append(sceneSubNode.Spheres, sphere)
		_getRecursiveBalls(sphere, maxRecursionDepth-1, 3, &sceneSubNode)
	}

	if takenSide != 3 { // offset in positive y
		childOrigin := parentSphere.Origin.Added(&vec3.T{0, childOffset, 0})
		sphere := scene.NewSphere(&childOrigin, childRadius, sphereMaterial).N(parentSphere.Name + " +y")
		sceneSubNode.Spheres = append(sceneSubNode.Spheres, sphere)
		_getRecursiveBalls(sphere, maxRecursionDepth-1, 4, &sceneSubNode)
	}

	if takenSide != 6 { // offset in negative z
		childOrigin := parentSphere.Origin.Added(&vec3.T{0, 0, -childOffset})
		sphere := scene.NewSphere(&childOrigin, childRadius, sphereMaterial).N(parentSphere.Name + " -z")
		sceneSubNode.Spheres = append(sceneSubNode.Spheres, sphere)
		_getRecursiveBalls(sphere, maxRecursionDepth-1, 5, &sceneSubNode)
	}

	if takenSide != 5 { // offset in positive z
		childOrigin := parentSphere.Origin.Added(&vec3.T{0, 0, childOffset})
		sphere := scene.NewSphere(&childOrigin, childRadius, sphereMaterial).N(parentSphere.Name + " +z")
		sceneSubNode.Spheres = append(sceneSubNode.Spheres, sphere)
		_getRecursiveBalls(sphere, maxRecursionDepth-1, 6, &sceneSubNode)
	}

	scn.ChildNodes = append(scn.ChildNodes, &sceneSubNode)
}
