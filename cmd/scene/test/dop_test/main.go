package main

import (
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
	"strconv"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "dop_test"

var ballRadius float64 = 30

var renderType = scn.Pathtracing

// var renderType = scn.Raycasting

var maxRecursionDepth = 4
var amountSamples = 128 * 4 * 4
var lensRadius = 15.0
var antiAlias = true

var viewPlaneDistance = 2000.0
var cameraDistanceFactor = 1.0

var imageWidth = 800
var imageHeight = 600
var magnification = 1.0

func main() {
	width := int(float64(imageWidth) * magnification)
	height := int(float64(imageHeight) * magnification)

	// Keep image proportions to an even amount of pixel for mp4 encoding
	if width%2 == 1 {
		width++
	}
	if height%2 == 1 {
		height++
	}

	animation := scn.Animation{
		AnimationName:     animationName,
		Frames:            []scn.Frame{},
		Width:             width,
		Height:            height,
		WriteRawImageFile: false,
	}

	scene := scn.SceneNode{
		Spheres: []*scn.Sphere{},
	}

	amountSpheres := 5
	sphereSpread := ballRadius * 2.0 * (float64(amountSpheres) + 1)
	sphereCC := sphereSpread / float64(amountSpheres)

	for i := 0; i <= amountSpheres; i++ {
		positionOffsetX := (-sphereSpread/2.0 + float64(i)*sphereCC) * 0.5
		positionOffsetZ := (-sphereSpread/2.0 + float64(i)*sphereCC) * 1.0

		sphere := scn.Sphere{
			Name:   "Glass sphere with transparency " + strconv.Itoa(i),
			Origin: vec3.T{positionOffsetX, ballRadius, positionOffsetZ},
			Radius: ballRadius,
			Material: &scn.Material{
				Color:      color.Color{R: 0.70, G: 0.95, B: 0.60},
				Glossiness: 0.80,
				Roughness:  0.00,
			},
		}
		scene.Spheres = append(scene.Spheres, &sphere)
	}

	environment := getEnvironmentMapping()
	scene.Spheres = append(scene.Spheres, &environment)

	camera := getCamera()

	frame := scn.Frame{
		Filename:   animation.AnimationName,
		FrameIndex: 0,
		Camera:     &camera,
		SceneNode:  &scene,
	}

	animation.Frames = append(animation.Frames, frame)

	anm.WriteAnimationToFile(animation, false)
}

func getEnvironmentMapping() scn.Sphere {
	sphericalImageProjection := scn.NewSphericalImageProjection("textures/equirectangular/ruins in the wilds.jpg", vec3.Zero, vec3.T{0, 0, 1}, vec3.UnitY)

	environmentMapping := scn.Material{
		Name:          "environment",
		Color:         color.Color{R: 1.0, G: 1.0, B: 1.0},
		Emission:      &color.Color{R: 1.0, G: 1.0, B: 1.0},
		Glossiness:    0.5,
		Roughness:     0.5,
		Projection:    &sphericalImageProjection,
		Transparency:  0.0,
		RayTerminator: true,
	}

	return scn.Sphere{
		Name:     "environment",
		Origin:   vec3.Zero,
		Radius:   1000.0,
		Material: &environmentMapping,
	}
}

func getCamera() scn.Camera {
	origin := vec3.T{0, ballRadius * 4, -800}
	origin.Scale(cameraDistanceFactor)

	heading := vec3.T{-origin[0], -(origin[1] - ballRadius), -origin[2]}
	focalDistance := heading.Length()

	return scn.Camera{
		Origin:            &origin,
		Heading:           &heading,
		ViewUp:            &vec3.T{0, 1, 0},
		ViewPlaneDistance: viewPlaneDistance,
		LensRadius:        lensRadius,
		FocalDistance:     focalDistance,
		Samples:           amountSamples,
		AntiAlias:         antiAlias,
		Magnification:     magnification,
		RenderType:        renderType,
		RecursionDepth:    maxRecursionDepth,
	}
}
