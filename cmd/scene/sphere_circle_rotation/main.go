package main

import (
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

var amountSamples = 512 * 2 * 8

var cameraDistanceFactor = 2.5

var circleRadius = 200.0
var amountBalls = len(projectionTextures) * 2
var ballRadius = 40.0
var viewPlaneDistance = 600.0
var lensRadius = 0.05

var amountBallsToRotateBeforeMovieLoop = len(projectionTextures)

func main() {
	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true)

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		ballAnimationTravelAngle := (2.0 * math.Pi) * float64(amountBallsToRotateBeforeMovieLoop) / float64(amountBalls)

		deltaBallAngle := ballAnimationTravelAngle * animationProgress
		projectionAngle := (2.0 * math.Pi) * animationProgress

		// Balls
		balls := addBallsToScene(deltaBallAngle, -projectionAngle, projectionTextures)

		// Reflective Center Ball
		mirrorSphereRadius := ballRadius * 3.0
		mirrorMaterial := scn.NewMaterial().C(color.Color{R: 0.90, G: 0.90, B: 0.90}).M(0.975, 0.0)
		reflectiveCenterBall := scn.NewSphere(&vec3.T{0, mirrorSphereRadius * 1, 0}, mirrorSphereRadius, mirrorMaterial).N("Mirror sphere")

		// Sky Dome
		skyDomeRadius := 100.0 * 1000.0
		skyDomeOrigin := &vec3.T{0, 0, 0}
		skyDomeMaterial := scn.NewMaterial().E(color.White, 1.0, true).SP(environmentEnvironMap, skyDomeOrigin, vec3.UnitY, vec3.UnitY)
		skyDome := scn.NewSphere(skyDomeOrigin, skyDomeRadius, skyDomeMaterial)

		camera := getCamera(magnification, animationProgress)

		scene := scn.NewSceneNode().
			S(balls...).
			S(reflectiveCenterBall, skyDome)

		frame := scn.NewFrame(animationName, frameIndex, camera, scene)

		animation.Frames = append(animation.Frames, frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

func addBallsToScene(deltaBallAngle float64, projectionAngle float64, projectionData []projection) []*scn.Sphere {
	var balls []*scn.Sphere

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

		material := scn.NewMaterial().
			E(*projectionData[projectionTextureIndex].emission, 1.0, projectionData[projectionTextureIndex].rayTerminator).
			SP(projectionData[projectionTextureIndex].filename, &ballOrigin, vec3.T{projectionU, 0, projectionV}, vec3.T{0, 1, 0})

		sphere := scn.NewSphere(&ballOrigin, ballRadius, material)

		balls = append(balls, sphere)
	}

	return balls
}

func getCamera(magnification float64, progress float64) *scn.Camera {
	degrees45 := math.Pi / 4.0
	strideAngle := degrees45 * math.Sin(2.0*math.Pi*progress)
	cameraDistance := 200.0 * cameraDistanceFactor
	cameraHeight := 100 * cameraDistanceFactor

	cameraOrigin := vec3.T{
		cameraDistance * math.Cos(-math.Pi/2.0+strideAngle),
		cameraHeight + (cameraHeight/2.0)*math.Sin(2.0*math.Pi*2.0*progress),
		cameraDistance * math.Sin(-math.Pi/2.0+strideAngle),
	}

	// Point heading towards center of sphere ring (heading vector starts in camera origin)
	focusPoint := vec3.T{0, ballRadius, 0}
	heading := focusPoint.Subed(&cameraOrigin)
	focusDistance := heading.Length() - circleRadius - 0.5*ballRadius

	return scn.NewCamera(&cameraOrigin, &focusPoint, amountSamples, magnification).
		V(viewPlaneDistance).A(lensRadius, "").F(focusDistance)
}
