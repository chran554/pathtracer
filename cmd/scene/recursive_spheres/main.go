package main

import (
	"fmt"
	"math"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

//var environmentEnvironMap = "textures/planets/environmentmap/space_fake_02_flip.png"
//var environmentEnvironMap = "textures/equirectangular/open_grassfield_sunny_day.jpg"

//var environmentEnvironMap = "textures/equirectangular/forest_sunny_day.jpg"

var environmentEnvironMap = "textures/equirectangular/canyon 3200x1600.jpeg"
var environmentRadius = 100.0 * 800.0 // 80m (if 1 unit is 1 cm)
var environmentEmissionFactor = float32(2.0)

//var environmentEnvironMap = "textures/planets/environmentmap/Stellarium3.jpeg"
//var environmentRadius = 100.0 * 80.0 // 80m (if 1 unit is 1 cm)

var animationName = "recursive_spheres"

var amountFrames = 1

var imageWidth = 1280
var imageHeight = 1024
var magnification = 1.0

var renderType = scn.Pathtracing
var amountSamples = 1024
var maxRecursion = 8

var lampEmissionFactor = 2.0
var lampDistanceFactor = 1.5

var cameraDistanceFactor = 2.8

var startSphereRadius = 150.0
var maxSphereRecursionDepth = 6

var circleRadius = 200.0
var viewPlaneDistance = 500.0
var lensRadius = 0.0

var sphereMaterial = scn.Material{
	Color:      &color.Color{R: 0.85, G: 0.95, B: 0.85},
	Glossiness: 0.95,
}

func main() {
	animation := getAnimation(int(float64(imageWidth)*magnification), int(float64(imageHeight)*magnification))
	animation.WriteRawImageFile = true

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		scene := scn.SceneNode{
			Spheres: []*scn.Sphere{},
			Discs:   []*scn.Disc{},
		}

		getRecursiveBalls(startSphereRadius, maxSphereRecursionDepth, &scene)

		// addReflectiveCenterBall(&scene)

		// addSphericalProjectionCenterBall(&scene)

		// addOriginCoordinateSpheres(&scene)

		//addLampsToScene(&scene)

		// addBottomDisc(&scene)

		addEnvironmentMapping(environmentEnvironMap, &scene)

		camera := getCamera(magnification, animationProgress)

		updateBoundingBoxes(&scene)
		scene.Bounds = nil

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

func updateBoundingBoxes(sceneNode *scn.SceneNode) *scn.Bounds {
	bb := scn.NewBounds()

	for _, sphere := range sceneNode.Spheres {
		bb.AddBounds(sphere.Bounds())
	}

	for _, disc := range sceneNode.Discs {
		bb.AddDiscBounds(disc)
	}

	for _, childNode := range sceneNode.ChildNodes {
		childBb := updateBoundingBoxes(childNode)
		bb.AddBounds(childBb)
	}

	sceneNode.Bounds = &bb

	return &bb
}

func getRecursiveBalls(middleSphereRadius float64, maxRecursionDepth int, scene *scn.SceneNode) {
	middleSphere := getSphere(vec3.T{0, 0, 0}, middleSphereRadius, "0")

	scene.Spheres = append(scene.Spheres, &middleSphere)
	_getRecursiveBalls(middleSphere, maxRecursionDepth, 0, scene)
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

func addLampsToScene(scene *scn.SceneNode) {
	lampEmission := color.Color{R: 5, G: 5, B: 5}
	lampEmission.Multiply(float32(lampEmissionFactor))

	lamp1 := scn.Sphere{
		Name:   "Lamp 1 (right)",
		Origin: &vec3.T{lampDistanceFactor * circleRadius * 1.5, lampDistanceFactor * circleRadius * 1.0, -lampDistanceFactor * circleRadius * 1.5},
		Radius: circleRadius * 0.75,
		Material: &scn.Material{
			Color:    &color.Color{R: 1, G: 1, B: 1},
			Emission: &lampEmission,
		},
	}

	lamp2 := scn.Sphere{
		Name:   "Lamp 2 (left)",
		Origin: &vec3.T{-lampDistanceFactor * circleRadius * 2.5, lampDistanceFactor * circleRadius * 1.5, -lampDistanceFactor * circleRadius * 2.0},
		Radius: circleRadius * 0.75,
		Material: &scn.Material{
			Color:    &color.Color{R: 1, G: 1, B: 1},
			Emission: &lampEmission,
		},
	}

	scene.Spheres = append(scene.Spheres, &lamp1, &lamp2)
}

func addEnvironmentMapping(filename string, scene *scn.SceneNode) {
	origin := vec3.T{0, 0, 0}

	sphere := scn.Sphere{
		Name:   "Environment mapping",
		Origin: &origin,
		Radius: environmentRadius,
		Material: &scn.Material{
			Color:         &color.Color{R: 1.0, G: 1.0, B: 1.0},
			Emission:      &color.Color{R: 1.0 * environmentEmissionFactor, G: 1.0 * environmentEmissionFactor, B: 1.0 * environmentEmissionFactor},
			RayTerminator: true,
			Projection: &scn.ImageProjection{
				ProjectionType: scn.Spherical,
				ImageFilename:  filename,
				Gamma:          1.5,
				Origin:         &origin,
				U:              &vec3.T{1, 0, 0},
				V:              &vec3.T{0, 1, 0},
				RepeatU:        true,
				RepeatV:        true,
				FlipU:          false,
				FlipV:          false,
			},
		},
	}

	scene.Spheres = append(scene.Spheres, &sphere)
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

func getCamera(magnification float64, progress float64) scn.Camera {
	degrees45 := math.Pi / 4.0
	strideAngle := degrees45 * math.Sin(2.0*math.Pi*progress)
	cameraDistance := 200.0 * cameraDistanceFactor
	cameraHeight := 75.0 * cameraDistanceFactor

	cameraOrigin := vec3.T{
		cameraDistance * math.Cos(-math.Pi/2.0+strideAngle+degrees45),
		cameraHeight + (cameraHeight/2.0)*math.Sin(2.0*math.Pi*2.0*progress),
		cameraDistance * math.Sin(-math.Pi/2.0+strideAngle+degrees45),
	}

	// Static camera location
	// cameraOrigin = vec3.T{0, cameraHeight, -cameraDistance}

	// Point heading towards center of sphere ring (heading vector starts in camera origin)
	heading := vec3.T{-cameraOrigin[0], -cameraOrigin[1], -cameraOrigin[2]}

	focalDistance := heading.Length() - startSphereRadius - 0.3*startSphereRadius

	return scn.Camera{
		Origin:            &cameraOrigin,
		Heading:           &heading,
		ViewUp:            &vec3.T{0, 1, 0},
		ViewPlaneDistance: viewPlaneDistance,
		ApertureSize:      lensRadius,
		FocusDistance:     focalDistance,
		Samples:           amountSamples,
		AntiAlias:         true,
		Magnification:     magnification,
		RenderType:        renderType,
		RecursionDepth:    maxRecursion,
	}
}
