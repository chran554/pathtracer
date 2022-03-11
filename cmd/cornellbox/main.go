package main

import (
	"encoding/json"
	"fmt"
	"os"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var ballRadius float64 = 20

var amountSamples = 256
var lensRadius float64 = 8
var antiAlias = true
var viewPlaneDistance = 1600.0

func main() {
	animation := scn.Animation{
		AnimationName:     "cornellbox",
		Frames:            []scn.Frame{},
		Width:             800 * 2,
		Height:            600 * 2,
		WriteRawImageFile: false,
	}

	scene := scn.Scene{
		Camera:  getCamera(),
		Spheres: []scn.Sphere{},
		Discs:   getBoxWalls(),
	}

	sphere1 := scn.Sphere{
		Name:   "Right sphere",
		Origin: vec3.T{ballRadius + (ballRadius / 2), ballRadius, 0},
		Radius: ballRadius,
		Material: scn.Material{
			Color: color.Color{R: 1, G: 1, B: 1},
		},
	}

	sphere2 := scn.Sphere{
		Name:   "Left sphere",
		Origin: vec3.T{-(ballRadius + (ballRadius / 2)), ballRadius, 0},
		Radius: ballRadius,
		Material: scn.Material{
			Color: color.Color{R: 1, G: 1, B: 1},
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

	fmt.Println("Wrote animation file:", filename)
}

func getBoxWalls() []scn.Disc {
	return []scn.Disc{
		{
			Name:   "Floor",
			Origin: vec3.T{0, 0, 0},
			Normal: vec3.T{0, 1, 0},
			Radius: 600,
			Material: scn.Material{
				Color: color.Color{R: 1, G: 1, B: 1},
			},
		},
		{
			Name:   "Right wall",
			Origin: vec3.T{ballRadius * 3, 0, 0},
			Normal: vec3.T{-1, 0, 0},
			Radius: 600,
			Material: scn.Material{
				Color: color.Color{R: 0.5, G: 0.5, B: 1},
			},
		},
		{
			Name:   "Left wall",
			Origin: vec3.T{-ballRadius * 3, 0, 0},
			Normal: vec3.T{1, 0, 0},
			Radius: 600,
			Material: scn.Material{
				Color: color.Color{R: 1, G: 0.5, B: 0.5},
			},
		},
		{
			Name:   "Roof",
			Origin: vec3.T{0, ballRadius * 6, 0},
			Normal: vec3.T{0, -1, 0},
			Radius: 600,
			Material: scn.Material{
				Color: color.Color{R: 1, G: 1, B: 1},
			},
		},
		{
			Name:   "Back wall",
			Origin: vec3.T{0, 0, ballRadius * 6},
			Normal: vec3.T{0, 0, -1},
			Radius: 600,
			Material: scn.Material{
				Color: color.Color{R: 1, G: 1, B: 1},
			},
		},
	}
}

func getCamera() scn.Camera {
	origin := vec3.T{0, ballRadius, -200}

	heading := vec3.T{-origin[0], -(origin[1] - ballRadius), -origin[2]}
	focalDistance := heading.Length()

	return scn.Camera{
		Origin:            origin,
		Heading:           heading,
		ViewUp:            vec3.T{0, 1, 0},
		ViewPlaneDistance: viewPlaneDistance,
		LensRadius:        lensRadius,
		FocalDistance:     focalDistance,
		Samples:           amountSamples,
		AntiAlias:         antiAlias,
	}
}
