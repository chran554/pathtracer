package main

import (
	"encoding/json"
	"fmt"
	"os"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var ballRadius float64 = 20

var amountSamples = 10
var lensRadius float64 = 0
var antiAlias = true

func main() {
	animation := scn.Animation{
		AnimationName: "cornellbox",
		Frames:        []scn.Frame{},
		Width:         800 * 2,
		Height:        600 * 2,
	}

	focalDistance := float64(200.0)
	scene := scn.Scene{
		Camera:  getCamera(focalDistance),
		Spheres: []scn.Sphere{},
		Discs:   getBoxWalls(),
	}

	sphere1 := scn.Sphere{
		Origin: vec3.T{ballRadius + (ballRadius / 2), ballRadius, 0},
		Radius: ballRadius,
		Material: scn.Material{
			Color:    scn.Color{R: 1, G: 1, B: 1},
			Emission: &scn.Black,
		},
	}

	sphere2 := scn.Sphere{
		Origin: vec3.T{-(ballRadius + (ballRadius / 2)), ballRadius, 0},
		Radius: ballRadius,
		Material: scn.Material{
			Color:    scn.Color{R: 1, G: 1, B: 1},
			Emission: &scn.Black,
		},
	}

	scene.Spheres = append(scene.Spheres, sphere1)
	scene.Spheres = append(scene.Spheres, sphere2)

	frame := scn.Frame{
		Filename:   animation.AnimationName,
		FrameIndex: 0,
		Scene:      scene,
	}

	animation.Frames = append(animation.Frames, frame)

	jsonData, err := json.MarshalIndent(animation, "", "  ")
	if err != nil {
		fmt.Println("Ouupps, failed to marshal data", err)
		os.Exit(1)
	}

	filename := "scene/" + animation.AnimationName + ".animation.json"
	if err = os.WriteFile(filename, jsonData, 0644); err != nil {
		fmt.Println("Ouuupps, no file writing performed")
		os.Exit(1)
	}
	fmt.Println("Write animation file:", filename)
}

func getBoxWalls() []scn.Disc {
	return []scn.Disc{
		{ // Floor
			Origin: vec3.T{0, 0, 0},
			Normal: vec3.T{0, 1, 0},
			Radius: 600,
			Material: scn.Material{
				Color:    scn.Color{R: 1, G: 1, B: 1},
				Emission: &scn.Black,
			},
		},
		{ // Right wall
			Origin: vec3.T{ballRadius * 3, 0, 0},
			Normal: vec3.T{-1, 0, 0},
			Radius: 600,
			Material: scn.Material{
				Color:    scn.Color{R: 0.5, G: 0.5, B: 1},
				Emission: &scn.Black,
			},
		},
		{ // Left wall
			Origin: vec3.T{-ballRadius * 3, 0, 0},
			Normal: vec3.T{1, 0, 0},
			Radius: 600,
			Material: scn.Material{
				Color:    scn.Color{R: 1, G: 0.5, B: 0.5},
				Emission: &scn.Black,
			},
		},
		{ // Roof
			Origin: vec3.T{0, ballRadius * 6, 0},
			Normal: vec3.T{0, -1, 0},
			Radius: 600,
			Material: scn.Material{
				Color:    scn.Color{R: 1, G: 1, B: 1},
				Emission: &scn.Black,
			},
		},
		{ // back wall
			Origin: vec3.T{0, 0, ballRadius * 6},
			Normal: vec3.T{0, 0, -1},
			Radius: 600,
			Material: scn.Material{
				Color:    scn.Color{R: 1, G: 1, B: 1},
				Emission: &scn.Black,
			},
		},
	}
}

func getCamera(focalDistance float64) scn.Camera {
	origin := vec3.T{0, ballRadius, -200}

	return scn.Camera{
		Origin:            origin,
		Heading:           vec3.T{-origin[0], -(origin[1] - ballRadius), -origin[2]},
		ViewUp:            vec3.T{0, 1, 0},
		ViewPlaneDistance: 1600,
		LensRadius:        lensRadius,
		FocalDistance:     focalDistance,
		Samples:           amountSamples,
		AntiAlias:         antiAlias,
	}
}
