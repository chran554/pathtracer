package main

import (
	"fmt"
	"math"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

type projection struct {
	filename      string
	emission      *color.Color
	rayTerminator bool
}

var projectionTextures = []projection{
	{filename: "textures/planets/earth_daymap.jpg"},
	{filename: "textures/planets/jupiter2_6k_contrast.png"},
	{filename: "textures/planets/moonmap4k_2.png"},
	{filename: "textures/planets/mars.jpg"},
	//{filename: "textures/planets/sun.jpg", emission: &color.Color{R: 32.0, G: 32.0, B: 32.0}, rayTerminator: true},
	{filename: "textures/planets/sunmap.jpg", emission: &color.Color{R: 42.0, G: 42.0, B: 42.0}, rayTerminator: true},
	{filename: "textures/planets/venusmap.jpg"},
	{filename: "textures/planets/makemake_fictional.jpg"},
	{filename: "textures/planets/plutomap2k.jpg"},
}

// var environmentEnvironMap = "textures/planets/environmentmap/space_fake_02_flip.png"
// var environmentEnvironMap = "textures/equirectangular/open_grassfield_sunny_day.jpg"
// var environmentEnvironMap = "textures/equirectangular/forest_sunny_day.jpg"
var environmentEnvironMap = "textures/planets/environmentmap/Stellarium3.jpeg"

var animationName = "sphere_circle_rotation"

var amountFrames = 1

var imageWidth = 800
var imageHeight = 600
var magnification = 1.0 * 2

var renderType = scn.Pathtracing
var amountSamples = 512 * 2 * 8
var maxRecursion = 6

var lampEmissionFactor = 2.0
var lampDistanceFactor = 1.5

var cameraDistanceFactor = 2.5

var circleRadius = 200.0
var amountBalls = len(projectionTextures) * 2
var ballRadius = 40.0
var viewPlaneDistance = 600.0
var lensRadius = 0.05

var amountBallsToRotateBeforeMovieLoop = len(projectionTextures)

func main() {
	animation := getAnimation(int(float64(imageWidth)*magnification), int(float64(imageHeight)*magnification))

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		scene := scn.SceneNode{
			Spheres: []*scn.Sphere{},
			Discs:   []*scn.Disc{},
		}

		ballAnimationTravelAngle := (2.0 * math.Pi) * float64(amountBallsToRotateBeforeMovieLoop) / float64(amountBalls)

		deltaBallAngle := ballAnimationTravelAngle * animationProgress
		projectionAngle := (2.0 * math.Pi) * animationProgress

		addBallsToScene(deltaBallAngle, -projectionAngle, projectionTextures, &scene)

		addReflectiveCenterBall(&scene)

		// addSphericalProjectionCenterBall(&scene)

		// addOriginCoordinateSpheres(&scene)

		// addLampsToScene(&scene)

		// addBottomDisc(&scene)

		addEnvironmentMapping(environmentEnvironMap, &scene)

		camera := getCamera(magnification, animationProgress)

		frame := scn.Frame{
			Filename:   animation.AnimationName + "_" + fmt.Sprintf("%06d", frameIndex),
			FrameIndex: frameIndex,
			SceneNode:  &scene,
			Camera:     &camera,
		}

		animation.Frames = append(animation.Frames, frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

func addBottomDisc(scene *scn.SceneNode) {
	scene.Discs = append(scene.Discs, getBottomPlate())
}

func addReflectiveCenterBall(scene *scn.SceneNode) {
	mirrorSphereRadius := ballRadius * 3.0
	sphere := scn.Sphere{
		Name:   "Mirror sphere",
		Origin: vec3.T{0, mirrorSphereRadius * 1, 0},
		Radius: mirrorSphereRadius,
		Material: &scn.Material{
			Color:      color.Color{R: 0.90, G: 0.90, B: 0.90},
			Glossiness: 0.975,
		},
	}

	scene.Spheres = append(scene.Spheres, &sphere)
}

func addSphericalProjectionCenterBall(scene *scn.SceneNode) {
	projectionSphereRadius := ballRadius * 3.0
	projectionSphereOrigin := vec3.T{0, projectionSphereRadius * 3, 0}

	projectionOrigin := projectionSphereOrigin

	projection := scn.NewSphericalImageProjection(
		"textures/checkered 360x180.png",
		//"textures/uv.png",
		projectionOrigin,
		//vec3.T{0, 0, -projectionSphereRadius},
		vec3.T{projectionSphereRadius, 0, 0},
		vec3.T{0, projectionSphereRadius, 0})

	sphere := scn.Sphere{
		Name:   "Spherical projected",
		Origin: projectionSphereOrigin,
		Radius: projectionSphereRadius,
		Material: &scn.Material{
			Color:      color.Color{R: 1, G: 1, B: 1},
			Projection: &projection,
		},
	}

	scene.Spheres = append(scene.Spheres, &sphere)
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
		Origin: vec3.T{-lampDistanceFactor * circleRadius * 1.5, lampDistanceFactor * circleRadius * 1.5, -lampDistanceFactor * circleRadius * 1.5},
		Radius: circleRadius * 0.75,
		Material: &scn.Material{
			Color:    color.Color{R: 1, G: 1, B: 1},
			Emission: &lampEmission,
		},
	}

	scene.Spheres = append(scene.Spheres, &lamp1, &lamp2)
}

func addEnvironmentMapping(filename string, scene *scn.SceneNode) {
	environmentRadius := 100.0 * 1000.0

	origin := vec3.T{0, 0, 0}

	sphere := scn.Sphere{
		Origin: origin,
		Radius: environmentRadius,
		Material: &scn.Material{
			Color:         color.Color{R: 1.0, G: 1.0, B: 1.0},
			Emission:      &color.Color{R: 1.0, G: 1.0, B: 1.0},
			RayTerminator: true,
			Projection: &scn.ImageProjection{
				ProjectionType: scn.Spherical,
				ImageFilename:  filename,
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

func addOriginCoordinateSpheres(scene *scn.SceneNode) {
	sphereOrigin := scn.Sphere{
		Origin:   vec3.T{0, ballRadius, 0},
		Radius:   ballRadius / 2,
		Material: &scn.Material{Color: color.Color{R: 0.1, G: 0.1, B: 0.1}},
	}
	sphereX := scn.Sphere{
		Origin:   vec3.T{ballRadius / 2, ballRadius, 0},
		Radius:   ballRadius / 2,
		Material: &scn.Material{Color: color.Color{R: 1, G: 1, B: 0}},
	}
	sphereZ := scn.Sphere{
		Origin:   vec3.T{0, ballRadius, ballRadius / 2},
		Radius:   ballRadius / 2,
		Material: &scn.Material{Color: color.Color{R: 0, G: 1, B: 1}},
	}

	scene.Spheres = append(scene.Spheres, &sphereOrigin)
	scene.Spheres = append(scene.Spheres, &sphereX)
	scene.Spheres = append(scene.Spheres, &sphereZ)
}

func addBallsToScene(deltaBallAngle float64, projectionAngle float64, projectionData []projection, scene *scn.SceneNode) {
	for ballIndex := 0; ballIndex < amountBalls; ballIndex++ {
		s := 2.0 * math.Pi
		t := float64(ballIndex) / float64(amountBalls)
		ballNominalAngle := s * t

		ballAngle := ballNominalAngle + deltaBallAngle

		// "Spin" sphere circle counterclockwise
		x := circleRadius * math.Cos(ballAngle)
		z := circleRadius * math.Sin(ballAngle)

		ballOrigin := vec3.T{x, 1.5*ballRadius - 3.0*ballRadius*float64(ballIndex%2), z}

		projectionTextureIndex := ballIndex % len(projectionData)

		// "Spin" single sphere projection clockwise (give the impression of sphere clockwise rotation)
		projectionU := math.Cos(projectionAngle)
		projectionV := math.Sin(projectionAngle)

		sphere := scn.Sphere{
			Origin: ballOrigin,
			Radius: ballRadius,
			Material: &scn.Material{
				Color:         color.Color{R: 1, G: 1, B: 1},
				Emission:      projectionData[projectionTextureIndex].emission,
				RayTerminator: projectionData[projectionTextureIndex].rayTerminator,
				Projection: &scn.ImageProjection{
					ProjectionType: scn.Spherical,
					ImageFilename:  projectionData[projectionTextureIndex].filename,
					Origin:         ballOrigin,
					U:              vec3.T{projectionU, 0, projectionV},
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
}

func getAnimation(width int, height int) scn.Animation {
	animation := scn.Animation{
		AnimationName:     animationName,
		Frames:            []scn.Frame{},
		Width:             width,
		Height:            height,
		WriteRawImageFile: true,
	}
	return animation
}

func getBottomPlate() *scn.Disc {
	origin := vec3.T{0, 0, 0}
	normal := vec3.T{0, 1, 0}
	textureScale := 400.0
	return &scn.Disc{
		Origin: &origin,
		Normal: &normal,
		Radius: circleRadius * 2,
		Material: &scn.Material{
			Color:    color.Color{R: 1, G: 1, B: 1},
			Emission: nil,
			Projection: &scn.ImageProjection{
				ProjectionType: scn.Parallel,
				ImageFilename:  "textures/rock_wall.png",
				Origin:         origin,
				U:              vec3.T{textureScale, 0, 0},
				V:              vec3.T{0, 0, textureScale},
				RepeatU:        true,
				RepeatV:        true,
				FlipU:          false,
				FlipV:          false,
			},
		},
	}
}

func getCamera(magnification float64, progress float64) scn.Camera {
	degrees45 := math.Pi / 4.0
	strideAngle := degrees45 * math.Sin(2.0*math.Pi*progress)
	cameraDistance := 200.0 * cameraDistanceFactor
	cameraHeight := 100 * cameraDistanceFactor

	origin := vec3.T{
		cameraDistance * math.Cos(-math.Pi/2.0+strideAngle),
		cameraHeight + (cameraHeight/2.0)*math.Sin(2.0*math.Pi*2.0*progress),
		cameraDistance * math.Sin(-math.Pi/2.0+strideAngle),
	}

	// Static camera location
	// origin = vec3.T{0, cameraHeight, -cameraDistance}

	// Point heading towards center of sphere ring (heading vector starts in camera origin)
	heading := vec3.T{-origin[0], -(origin[1] - ballRadius), -origin[2]}

	focalDistance := heading.Length() - circleRadius - 0.5*ballRadius

	return scn.Camera{
		Origin:            &origin,
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
