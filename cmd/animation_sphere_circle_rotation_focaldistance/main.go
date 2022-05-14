package main

import (
	"fmt"
	"math"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var amountFrames = 128
var amountBalls = 16
var ballRadius = 20.0
var circleRadius = 100.0

var amountSamples = 512
var lensRadius = 15.0

var nominalViewPlaneDistance = 800.0
var magnification = 1.0

var cameraOrigin = vec3.T{0, 200, -200}

func main() {
	animation := scn.Animation{
		AnimationName: "sphere_circle_rotation_focaldistance_hires",
		Frames:        []scn.Frame{},
		Width:         800,
		Height:        600,
	}

	nominalFocalDistance := cameraOrigin.Length()

	ballAngle := (2.0 * math.Pi) / float64(amountBalls)
	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)
		deltaFrameAngle := ballAngle * animationProgress

		// Focal plane distance animation
		//focalDistance := nominalFocalDistance
		focalDistance := nominalFocalDistance + circleRadius*math.Sin(math.Pi*2.0*animationProgress)

		// View plane distance animation
		viewPlaneDistance := nominalViewPlaneDistance
		//viewPlaneDistance := nominalViewPlaneDistance + (nominalViewPlaneDistance/2.0)*float64(math.Sin(math.Pi*2.0*animationProgress))

		scene := scn.SceneNode{
			Spheres: []scn.Sphere{},
			Discs:   getBottomPlate(),
		}

		for ballIndex := 0; ballIndex < amountBalls; ballIndex++ {
			s := 2.0 * math.Pi
			t := float64(ballIndex) / float64(amountBalls)
			angle := s * t
			x := circleRadius * math.Cos(angle+deltaFrameAngle)
			z := circleRadius * math.Sin(angle+deltaFrameAngle)

			sphere := scn.Sphere{
				Origin: vec3.T{x, ballRadius, z},
				Radius: ballRadius,
				Material: scn.Material{
					Color:    color.Color{R: 1, G: 1, B: 1},
					Emission: nil,
				},
			}

			scene.Spheres = append(scene.Spheres, sphere)
		}

		camera := getCamera(focalDistance, viewPlaneDistance)
		frame := scn.Frame{
			Filename:   animation.AnimationName + "_" + fmt.Sprintf("%06d", frameIndex),
			FrameIndex: frameIndex,
			Camera:     &camera,
			SceneNode:  &scene,
		}

		animation.Frames = append(animation.Frames, frame)
	}

	anm.WriteAnimationToFile(animation)
}

func getCamera(focalDistance float64, viewPlaneDistance float64) scn.Camera {
	return scn.Camera{
		Origin:            cameraOrigin,
		Heading:           vec3.T{-cameraOrigin[0], -(cameraOrigin[1] - ballRadius), -cameraOrigin[2]},
		ViewUp:            vec3.T{0, 1, 0},
		ViewPlaneDistance: viewPlaneDistance,
		LensRadius:        lensRadius,
		FocalDistance:     focalDistance,
		Samples:           amountSamples,
		AntiAlias:         true,
		Magnification:     magnification,
	}
}

func getBottomPlate() []scn.Disc {
	origin := vec3.T{0, 0, 0}

	u := vec3.T{50, 0, 0}
	v := vec3.T{0, 0, 50}
	parallelImageProjection := scn.NewParallelImageProjection("textures/white_marble.png", origin, u, v)
	return []scn.Disc{
		{
			Origin: origin,
			Normal: vec3.T{0, 1, 0},
			Radius: 600,
			Material: scn.Material{
				Color:      color.Color{R: 0.5, G: 0.5, B: 0.5},
				Emission:   nil,
				Projection: &parallelImageProjection,
			},
		},
	}
}
