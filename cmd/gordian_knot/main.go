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

var animationName = "gordian_knot"

var amountFrames = 1

var imageWidth = 1280
var imageHeight = 1024
var magnification = 1.0

var renderType = scn.Pathtracing
var amountSamples = 256
var maxRecursion = 3

var lampEmissionFactor = 2.0
var lampDistanceFactor = 1.5

var cameraDistanceFactor = 2.8

var startSphereRadius = 150.0
var maxSphereRecursionDepth = 6

var circleRadius = 200.0
var viewPlaneDistance = 800.0
var lensRadius = 0.0

var sphereMaterial = scn.Material{
	Color:      color.Color{R: 0.85, G: 0.95, B: 0.85},
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

		// spheres := getKnotBalls(50.0, 360*4, 150)
		// scene.Spheres = spheres

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

func getKnotBalls(ballRadius float64, amountBalls int, scale float32) []*scn.Sphere {
	spheres := make([]*scn.Sphere, 0)
	radiansPerBall := 360.0 / float64(amountBalls)

	for ballIndex := 0; ballIndex < amountBalls; ballIndex++ {
		t := radiansPerBall * float64(ballIndex)

		x := float64(scale) * (math.Cos(t) + 2.0*math.Cos(2.0*t))
		y := float64(scale) * (math.Sin(t) - 2.0*math.Sin(2.0*t))
		z := float64(scale) * math.Sin(3.0*t)

		sphere := getSphere(vec3.T{x, y, z}, ballRadius, fmt.Sprintf("%d", ballIndex))

		spheres = append(spheres, &sphere)
	}

	return spheres
}

func updateBoundingBoxes(sceneNode *scn.SceneNode) *scn.Bounds {
	bb := scn.NewBounds()

	for _, sphere := range sceneNode.Spheres {
		bb.AddSphereBounds(sphere)
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

func getSphere(origin vec3.T, radius float64, name string) scn.Sphere {
	return scn.Sphere{
		Name:     name,
		Origin:   origin,
		Radius:   radius,
		Material: &sphereMaterial,
	}
}

func addLampsToScene(scene *scn.SceneNode) {
	lampEmission := color.Color{R: 5, G: 5, B: 5}
	lampEmission.Multiply(float32(lampEmissionFactor))

	lamp1 := scn.Sphere{
		Name:   "Lamp 1 (right)",
		Origin: vec3.T{lampDistanceFactor * circleRadius * 1.5, lampDistanceFactor * circleRadius * 1.0, -lampDistanceFactor * circleRadius * 1.5},
		Radius: circleRadius * 0.75,
		Material: &scn.Material{
			Color:    color.Color{R: 1, G: 1, B: 1},
			Emission: &lampEmission,
		},
	}

	lamp2 := scn.Sphere{
		Name:   "Lamp 2 (left)",
		Origin: vec3.T{-lampDistanceFactor * circleRadius * 2.5, lampDistanceFactor * circleRadius * 1.5, -lampDistanceFactor * circleRadius * 2.0},
		Radius: circleRadius * 0.75,
		Material: &scn.Material{
			Color:    color.Color{R: 1, G: 1, B: 1},
			Emission: &lampEmission,
		},
	}

	scene.Spheres = append(scene.Spheres, &lamp1, &lamp2)
}

func addEnvironmentMapping(filename string, scene *scn.SceneNode) {
	origin := vec3.T{0, 0, 0}

	sphere := scn.Sphere{
		Name:   "Environment mapping",
		Origin: origin,
		Radius: environmentRadius,
		Material: &scn.Material{
			Color:         color.Color{R: 1.0, G: 1.0, B: 1.0},
			Emission:      &color.Color{R: 1.0 * environmentEmissionFactor, G: 1.0 * environmentEmissionFactor, B: 1.0 * environmentEmissionFactor},
			RayTerminator: true,
			Projection: &scn.ImageProjection{
				ProjectionType: scn.Spherical,
				ImageFilename:  filename,
				Gamma:          1.5,
				Origin:         origin,
				U:              vec3.T{1, 0, 0},
				V:              vec3.T{0, 1, 0},
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
		LensRadius:        lensRadius,
		FocalDistance:     focalDistance,
		Samples:           amountSamples,
		AntiAlias:         true,
		Magnification:     magnification,
		RenderType:        renderType,
		RecursionDepth:    maxRecursion,
	}
}
