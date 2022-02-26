package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/vec3"
)

var amountFrames = 32
var amountBalls = 16
var ballRadius float32 = 20
var circleRadius float32 = 100

var amountSamples = 256
var lensRadius float32 = 10

func main() {
	animation := scn.Animation{
		AnimationName: "sphere_circle_rotation",
		Frames:        []scn.Frame{},
		Width:         800,
		Height:        600,
	}

	ballAngle := float32(2.0*math.Pi) / float32(amountBalls)
	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		deltaFrameAngle := ballAngle * (float32(frameIndex) / float32(amountFrames))

		scene := scn.Scene{
			Camera:  getCamera(),
			Spheres: []scn.Sphere{},
			Discs:   getBottomPlate(),
		}

		for ballIndex := 0; ballIndex < amountBalls; ballIndex++ {
			s := float32(2.0 * math.Pi)
			t := float32(ballIndex) / float32(amountBalls)
			angle := s * t
			x := circleRadius * float32(math.Cos(float64(angle+deltaFrameAngle)))
			z := circleRadius * float32(math.Sin(float64(angle+deltaFrameAngle)))

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

	json, err := json.MarshalIndent(animation, "", "  ")
	if err != nil {
		fmt.Println("Ouupps", err)
	}

	filename := "scene/" + animation.AnimationName + ".animation.json"
	os.WriteFile(filename, json, 0644)
	fmt.Println("Write animation file:", filename)
}

func getBottomPlate() []scn.Disc {
	return []scn.Disc{
		{
			Origin: vec3.T{0, 0, 0},
			Normal: vec3.T{0, 1, 0},
			Radius: 600,
			Material: scn.Material{
				Color:    scn.Color{R: 0.75, G: 0.75, B: 0.75},
				Emission: scn.Black,
			},
		},
	}
}

func getCamera() scn.Camera {
	origin := vec3.T{0, 100, -200}

	return scn.Camera{
		Origin:            origin,
		Heading:           vec3.T{-origin[0], -(origin[1] - ballRadius), -origin[2]},
		ViewUp:            vec3.T{0, 1, 0},
		ViewPlaneDistance: 400,
		LensRadius:        lensRadius,
		FocalDistance:     200,
		Samples:           amountSamples,
		AntiAlias:         true,
	}
}
