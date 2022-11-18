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

var environmentEnvironMap = "textures/equirectangular/sunset horizon 2800x1400.jpg"
var environmentRadius = 100.0 * 1000.0
var environmentEmissionFactor = float32(1.0)

var amountFrames = 1

var imageWidth = 1280
var imageHeight = 1024
var magnification = 1.0

var renderType = scn.Pathtracing
var amountSamples = 2000
var maxRecursion = 4

var ballsLightEmissionFactor float32 = 2

var startSphereRadius = 150.0
var maxSphereRecursionDepth = 6

var viewPlaneDistance = 800.0
var apertureSize = 5.0

var sphereMaterial = scn.Material{
	Color:      (&color.Color{R: 0.85, G: 0.95, B: 0.85}).Multiply(0.50),
	Glossiness: 0.85,
	Roughness:  0.01,
	Emission:   (&color.Color{R: 0.85, G: 0.95, B: 0.85}).Multiply(0.05),
}

func main() {
	animation := getAnimation(int(float64(imageWidth)*magnification), int(float64(imageHeight)*magnification))
	animation.WriteRawImageFile = true

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		recursiveBalls := getRecursiveBalls(startSphereRadius, maxSphereRecursionDepth)
		ballsBounds := recursiveBalls.UpdateBounds()
		recursiveBalls.Translate(&vec3.T{0, -ballsBounds.Ymin, 0})
		recursiveBalls.RotateY(&vec3.Zero, math.Pi/10)
		recursiveBalls.RotateX(&vec3.Zero, math.Pi/12)
		ballsBounds = recursiveBalls.UpdateBounds()
		fmt.Printf("Balls bounds: %+v   (center: %+v)\n", ballsBounds, ballsBounds.Center())

		ballsLightDistanceFactor := 400.0
		ballsLightPosition := ballsBounds.Center().Add(&vec3.T{-ballsLightDistanceFactor, ballsLightDistanceFactor, -2.0 * ballsLightDistanceFactor})
		ballsLightEmission := (&color.Color{R: 15, G: 14.0, B: 12.0}).Multiply(ballsLightEmissionFactor)
		ballsLightMaterial := scn.Material{Color: &color.White, Emission: ballsLightEmission, Glossiness: 0.0, Roughness: 1.0, RayTerminator: true}
		ballsLight := scn.Sphere{Name: "Balls light", Origin: ballsLightPosition, Radius: 200.0, Material: &ballsLightMaterial}

		environmentSphere := addEnvironmentMapping(environmentEnvironMap)

		// Ground

		groundProjection := scn.NewParallelImageProjection("textures/ground/soil-cracked.png", vec3.Zero, vec3.UnitX.Scaled(250*2), vec3.UnitZ.Scaled(250*2))
		ground := scn.Disc{
			Name:   "Ground",
			Origin: &vec3.Zero,
			Normal: &vec3.UnitY,
			Radius: environmentRadius,
			Material: &scn.Material{
				Name:       "Ground material",
				Color:      &color.White,
				Emission:   &color.Black,
				Glossiness: 0.0,
				Roughness:  1.0,
				Projection: &groundProjection,
			},
		}

		cameraFocusPoint := ballsBounds.Center().Add(&vec3.T{0, ballsBounds.SizeZ() / 10.0, -ballsBounds.SizeZ() / 2.0 * 0.8})
		cameraOrigin := ballsBounds.Center().Add(&vec3.T{0, ballsBounds.SizeZ() * 2.0 / 10.0, -800})
		camera := getCamera(magnification, animationProgress, cameraOrigin, cameraFocusPoint)

		scene := scn.SceneNode{
			Spheres:    []*scn.Sphere{&environmentSphere, &ballsLight},
			Discs:      []*scn.Disc{&ground},
			ChildNodes: []*scn.SceneNode{recursiveBalls},
		}

		frame := scn.Frame{
			Filename:   animation.AnimationName + "_" + fmt.Sprintf("%06d", frameIndex),
			FrameIndex: frameIndex,
			Camera:     &camera,
			SceneNode:  &scene,
		}

		animation.Frames = append(animation.Frames, frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

func getRecursiveBalls(middleSphereRadius float64, maxRecursionDepth int) *scn.SceneNode {
	scene := scn.SceneNode{}

	middleSphere := getSphere(vec3.T{0, 0, 0}, middleSphereRadius, "0")
	scene.Spheres = append(scene.Spheres, &middleSphere)
	_getRecursiveBalls(middleSphere, maxRecursionDepth, 0, &scene)

	return &scene
}

func _getRecursiveBalls(parentSphere scn.Sphere, maxRecursionDepth int, takenSide int, scene *scn.SceneNode) {
	var sceneSubNode scn.SceneNode

	if parentSphere.Radius < 5.0 || maxRecursionDepth == 0 {
		return
	}

	childRadius := parentSphere.Radius * 0.48
	childOffset := parentSphere.Radius + childRadius*1.05

	if takenSide != 2 { // offset in negative x
		sphere := getSphere(parentSphere.Origin.Added(&vec3.T{-childOffset, 0, 0}), childRadius, parentSphere.Name+" -x")
		sceneSubNode.Spheres = append(sceneSubNode.Spheres, &sphere)
		_getRecursiveBalls(sphere, maxRecursionDepth-1, 1, &sceneSubNode)
	}

	if takenSide != 1 { // offset in positive x
		sphere := getSphere(parentSphere.Origin.Added(&vec3.T{childOffset, 0, 0}), childRadius, parentSphere.Name+" +x")
		sceneSubNode.Spheres = append(sceneSubNode.Spheres, &sphere)
		_getRecursiveBalls(sphere, maxRecursionDepth-1, 2, &sceneSubNode)
	}

	if takenSide != 4 { // offset in negative y
		sphere := getSphere(parentSphere.Origin.Added(&vec3.T{0, -childOffset, 0}), childRadius, parentSphere.Name+" -y")
		sceneSubNode.Spheres = append(sceneSubNode.Spheres, &sphere)
		_getRecursiveBalls(sphere, maxRecursionDepth-1, 3, &sceneSubNode)
	}

	if takenSide != 3 { // offset in positive y
		sphere := getSphere(parentSphere.Origin.Added(&vec3.T{0, childOffset, 0}), childRadius, parentSphere.Name+" +y")
		sceneSubNode.Spheres = append(sceneSubNode.Spheres, &sphere)
		_getRecursiveBalls(sphere, maxRecursionDepth-1, 4, &sceneSubNode)
	}

	if takenSide != 6 { // offset in negative z
		sphere := getSphere(parentSphere.Origin.Added(&vec3.T{0, 0, -childOffset}), childRadius, parentSphere.Name+" -z")
		sceneSubNode.Spheres = append(sceneSubNode.Spheres, &sphere)
		_getRecursiveBalls(sphere, maxRecursionDepth-1, 5, &sceneSubNode)
	}

	if takenSide != 5 { // offset in positive z
		sphere := getSphere(parentSphere.Origin.Added(&vec3.T{0, 0, childOffset}), childRadius, parentSphere.Name+" +z")
		sceneSubNode.Spheres = append(sceneSubNode.Spheres, &sphere)
		_getRecursiveBalls(sphere, maxRecursionDepth-1, 6, &sceneSubNode)
	}

	scene.ChildNodes = append(scene.ChildNodes, &sceneSubNode)
}

func getSphere(origin vec3.T, radius float64, name string) scn.Sphere {
	return scn.Sphere{
		Name:     name,
		Origin:   &origin,
		Radius:   radius,
		Material: &sphereMaterial,
	}
}

func addEnvironmentMapping(filename string) scn.Sphere {
	origin := vec3.T{0, 0, 0}

	projection := scn.ImageProjection{
		ProjectionType: scn.Spherical,
		ImageFilename:  filename,
		Gamma:          1.0,
		Origin:         &origin,
		U:              &vec3.T{-0.2, 0, -1},
		V:              &vec3.T{0, 1, 0},
		RepeatU:        true,
		RepeatV:        true,
		FlipU:          false,
		FlipV:          false,
	}

	material := scn.Material{
		Color:         &color.Color{R: 1.0, G: 1.0, B: 1.0},
		Emission:      &color.Color{R: 1.0 * environmentEmissionFactor, G: 1.0 * environmentEmissionFactor, B: 1.0 * environmentEmissionFactor},
		RayTerminator: true,
		Projection:    &projection,
	}

	sphere := scn.Sphere{
		Name:     "Environment mapping",
		Origin:   &origin,
		Radius:   environmentRadius,
		Material: &material,
	}

	return sphere
}

func getAnimation(width int, height int) scn.Animation {
	animation := scn.Animation{
		AnimationName:     animationName,
		Frames:            []scn.Frame{},
		Width:             width,
		Height:            height,
		WriteRawImageFile: false,
	}
	return animation
}

func getCamera(magnification float64, progress float64, cameraOrigin *vec3.T, cameraFocusPoint *vec3.T) scn.Camera {

	// Point heading towards center of sphere ring (heading vector starts in camera origin)
	heading := cameraFocusPoint.Subed(cameraOrigin)

	focusDistance := heading.Length()

	return scn.Camera{
		Origin:            cameraOrigin,
		Heading:           &heading,
		ViewUp:            &vec3.T{0, 1, 0},
		ViewPlaneDistance: viewPlaneDistance,
		ApertureSize:      apertureSize,
		FocusDistance:     focusDistance,
		Samples:           amountSamples,
		AntiAlias:         true,
		Magnification:     magnification,
		RenderType:        renderType,
		RecursionDepth:    maxRecursion,
	}
}
