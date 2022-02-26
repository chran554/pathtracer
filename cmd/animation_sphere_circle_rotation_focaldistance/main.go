package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/vec3"
)

var amountFrames = 64
var amountBalls = 16
var ballRadius float32 = 20
var circleRadius float32 = 100

var amountSamples = 256
var lensRadius float32 = 15

var cameraOrigin = vec3.T{0, 200, -200}

func main() {
	animation := scn.Animation{
		AnimationName: "sphere_circle_rotation_focaldistance",
		Frames:        []scn.Frame{},
		Width:         600,
		Height:        450,
	}

	nominalFocalDistance := cameraOrigin.Length()
	ballAngle := (2.0 * math.Pi) / float64(amountBalls)
	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)
		deltaFrameAngle := ballAngle * animationProgress

		focalDistance := nominalFocalDistance + circleRadius*float32(math.Sin(math.Pi*2.0*animationProgress))
		scene := scn.Scene{
			Camera:  getCamera(focalDistance),
			Spheres: []scn.Sphere{},
			Discs:   getBottomPlate(),
		}

		for ballIndex := 0; ballIndex < amountBalls; ballIndex++ {
			s := 2.0 * math.Pi
			t := float64(ballIndex) / float64(amountBalls)
			angle := s * t
			x := circleRadius * float32(math.Cos(angle+deltaFrameAngle))
			z := circleRadius * float32(math.Sin(angle+deltaFrameAngle))

			sphere := scn.Sphere{
				Origin: vec3.T{x, ballRadius, z},
				Radius: ballRadius,
				Material: scn.Material{
					Color:    scn.Color{R: 1, G: 1, B: 1},
					Emission: scn.Black,
				},
			}

			scene.Spheres = append(scene.Spheres, sphere)
		}

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
	os.WriteFile(filename, jsonData, 0644)
	fmt.Println("Write animation file:", filename)
}

func getBottomPlate() []scn.Disc {
	return []scn.Disc{
		{
			Origin: vec3.T{0, 0, 0},
			Normal: vec3.T{0, 1, 0},
			Radius: 600,
			Material: scn.Material{
				Color:    scn.Color{R: 0.5, G: 0.5, B: 0.5},
				Emission: scn.Black,
			},
		},
	}
}

func getCamera(focalDistance float32) scn.Camera {

	return scn.Camera{
		Origin:            cameraOrigin,
		Heading:           vec3.T{-cameraOrigin[0], -(cameraOrigin[1] - ballRadius), -cameraOrigin[2]},
		ViewUp:            vec3.T{0, 1, 0},
		ViewPlaneDistance: 400,
		LensRadius:        lensRadius,
		FocalDistance:     focalDistance,
		Samples:           amountSamples,
		AntiAlias:         true,
	}
}
