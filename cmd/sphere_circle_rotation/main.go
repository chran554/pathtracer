package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

type projection struct {
	filename string
}

var width = 800
var height = 600

var amountFrames = 180
var amountBalls = 16
var ballRadius float64 = 20
var circleRadius float64 = 100
var amountSamples = 128
var lensRadius float64 = 6

func main() {
	animation := getAnimation(width, height)

	ballAngle := 8.0 * (2.0 * math.Pi) / float64(amountBalls)

	projectionData := []projection{
		{filename: "textures/planets/earth_daymap.jpg"},
		{filename: "textures/planets/jupiter2_6k_contrast.png"},
		{filename: "textures/planets/moonmap4k_2.png"},
		{filename: "textures/planets/mars.jpg"},
		{filename: "textures/planets/sun.jpg"},
		{filename: "textures/planets/venusmap.jpg"},
		{filename: "textures/planets/makemake_fictional.jpg"},
		{filename: "textures/planets/plutomap2k.jpg"},
	}

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		scene := scn.Scene{
			Camera:  getCamera(),
			Spheres: []scn.Sphere{},
			Discs:   getBottomPlate(),
		}

		deltaFrameAngle := ballAngle * (float64(frameIndex) / float64(amountFrames))
		addBallsToScene(deltaFrameAngle, projectionData, &scene)

		//addOriginCoordinateSpheres(&scene)

		frame := scn.Frame{
			Filename:   animation.AnimationName + "_" + fmt.Sprintf("%06d", frameIndex),
			FrameIndex: frameIndex,
			Scene:      scene,
		}

		animation.Frames = append(animation.Frames, frame)
	}

	jsonData, err := json.MarshalIndent(animation, "", "  ")
	if err != nil {
		fmt.Println("Ouupps", err)
	}

	filename := "scene/" + animation.AnimationName + ".animation.json"
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		fmt.Println("Could not write animation file:", filename)
		os.Exit(1)
	} else {
		fmt.Println("Wrote animation file:", filename)
	}
}

func addOriginCoordinateSpheres(scene *scn.Scene) {
	sphereOrigin := scn.Sphere{
		Origin:   vec3.T{0, ballRadius, 0},
		Radius:   ballRadius / 2,
		Material: scn.Material{Color: scn.Color{R: 0.1, G: 0.1, B: 0.1}},
	}
	sphereX := scn.Sphere{
		Origin:   vec3.T{ballRadius / 2, ballRadius, 0},
		Radius:   ballRadius / 2,
		Material: scn.Material{Color: scn.Color{R: 1, G: 1, B: 0}},
	}
	sphereZ := scn.Sphere{
		Origin:   vec3.T{0, ballRadius, ballRadius / 2},
		Radius:   ballRadius / 2,
		Material: scn.Material{Color: scn.Color{R: 0, G: 1, B: 1}},
	}
	scene.Spheres = append(scene.Spheres, sphereOrigin)
	scene.Spheres = append(scene.Spheres, sphereX)
	scene.Spheres = append(scene.Spheres, sphereZ)
}

func addBallsToScene(deltaFrameAngle float64, projectionData []projection, scene *scn.Scene) {
	for ballIndex := 0; ballIndex < amountBalls; ballIndex++ {
		s := 2.0 * math.Pi
		t := float64(ballIndex) / float64(amountBalls)
		angle := s * t
		x := circleRadius * math.Cos(angle+deltaFrameAngle)
		z := circleRadius * math.Sin(angle+deltaFrameAngle)

		projectionDataIndex := ballIndex % len(projectionData)
		ballOrigin := vec3.T{x, ballRadius, z}
		projectionOrigin := ballOrigin
		projectionOrigin.Sub(&vec3.T{0, ballRadius, 0})

		sphere := scn.Sphere{
			Origin: ballOrigin,
			Radius: ballRadius,
			Material: scn.Material{
				Color:    scn.Color{R: 1, G: 1, B: 1},
				Emission: nil,
				Projection: &scn.ImageProjection{
					ProjectionType: scn.Cylindrical,
					ImageFilename:  projectionData[projectionDataIndex].filename,
					Origin:         projectionOrigin,
					U:              vec3.T{x, 0, z},
					V:              vec3.T{0, 2 * ballRadius, 0},
					RepeatU:        true,
					RepeatV:        true,
					FlipU:          false,
					FlipV:          false,
				},
			},
		}

		scene.Spheres = append(scene.Spheres, sphere)
	}
}

func getAnimation(width int, height int) scn.Animation {
	animation := scn.Animation{
		AnimationName:     "sphere_circle_rotation",
		Frames:            []scn.Frame{},
		Width:             width,
		Height:            height,
		WriteRawImageFile: false,
	}
	return animation
}

func getBottomPlate() []scn.Disc {
	origin := vec3.T{0, 0, 0}
	normal := vec3.T{0, 1, 0}
	textureScale := 400.0
	return []scn.Disc{
		{
			Origin: origin,
			Normal: normal,
			Radius: 600,
			Material: scn.Material{
				Color:    scn.Color{R: 1, G: 1, B: 1},
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
		},
	}
}

func getCamera() scn.Camera {
	cameraDistanceFactor := 1.6
	origin := vec3.T{0 * cameraDistanceFactor, 100 * cameraDistanceFactor, -200 * cameraDistanceFactor}
	heading := vec3.T{-origin[0], -(origin[1] - ballRadius), -origin[2]}

	focalDistance := heading.Length() - 0.8*circleRadius

	return scn.Camera{
		Origin:            origin,
		Heading:           heading,
		ViewUp:            vec3.T{0, 1, 0},
		ViewPlaneDistance: 1000,
		LensRadius:        lensRadius,
		FocalDistance:     focalDistance,
		Samples:           amountSamples,
		AntiAlias:         true,
	}
}
