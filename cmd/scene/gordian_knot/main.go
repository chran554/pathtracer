package main

import (
	"fmt"
	"math"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

// var environmentEnvironMap = "textures/planets/environmentmap/space_fake_02_flip.png"
// var environmentEnvironMap = "textures/equirectangular/open_grassfield_sunny_day.jpg"

var environmentEnvironMap = "textures/equirectangular/forest_sunny_day.jpg"

// var environmentEnvironMap = "textures/equirectangular/canyon 3200x1600.jpeg"
var environmentRadius = 100.0 * 100.0 // 100m (if 1 unit is 1 cm)
var environmentEmissionFactor = float32(1.7)

//var environmentEnvironMap = "textures/planets/environmentmap/Stellarium3.jpeg"
//var environmentRadius = 100.0 * 80.0 // 80m (if 1 unit is 1 cm)

var animationName = "gordian_knot"

var amountFrames = 36 * 10

var imageWidth = 1280
var imageHeight = 1024
var magnification = 0.5

var renderType = scn.Pathtracing
var amountSamples = 256
var maxRecursion = 3

var cameraDistanceFactor = 2.8

var viewPlaneDistance = 800.0
var lensRadius = 0.0

var pipeLength = 0.83333
var pipeRadius = 40.0
var amountBalls = 75
var scale = 100.0

var sphereMaterial = scn.Material{
	Color:      &color.Color{R: 0.95, G: 0.95, B: 0.95},
	Glossiness: 0.90,
	Roughness:  0.10,
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

		//ballSpeed := amountBalls / 5.0 // All balls make 1/5:th of a full loop during the whole animation
		ballSpeed := float64(amountBalls) / pipeLength // All balls make a full loop during the whole animation
		spheres := getKnotBalls(pipeRadius, amountBalls, scale, animationProgress, ballSpeed)
		scene.Spheres = spheres

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

func getKnotBalls(ballRadius float64, amountBalls int, scale float64, animationProgress float64, ballSpeed float64) []*scn.Sphere {
	spheres := make([]*scn.Sphere, 0)
	radianDistanceBetweenBalls := (math.Pi * 2.0 * pipeLength) / float64(amountBalls)

	for ballIndex := 0; ballIndex < amountBalls; ballIndex++ {
		t := (radianDistanceBetweenBalls * float64(ballIndex)) + (animationProgress * ballSpeed * radianDistanceBetweenBalls)

		x := scale * (math.Cos(t) + 2.0*math.Cos(2.0*t))
		y := scale * (math.Sin(t) - 2.0*math.Sin(2.0*t))
		z := scale * -1.0 * math.Sin(3.0*t)

		sphere := getSphere(vec3.T{x, y, z}, ballRadius, fmt.Sprintf("sphere #%d", ballIndex))

		spheres = append(spheres, &sphere)
	}

	return spheres
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

func getSphere(origin vec3.T, radius float64, name string) scn.Sphere {
	return scn.Sphere{
		Name:     name,
		Origin:   &origin,
		Radius:   radius,
		Material: &sphereMaterial,
	}
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
				U:              &vec3.T{0, 0, 1},
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
		cameraDistance * math.Sin(-math.Pi/2.0+strideAngle*0.5+degrees45),
	}

	// Point heading towards center of sphere ring (heading vector starts in camera origin)
	heading := vec3.T{-cameraOrigin[0], -cameraOrigin[1], -cameraOrigin[2]}

	focalDistance := heading.Length()

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
