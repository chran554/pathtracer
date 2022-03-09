package main

import (
	"encoding/json"
	"fmt"
	"os"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var ballRadius float64 = 30

var amountSamples = 9
var lensRadius float64 = 0
var antiAlias = true
var viewPlaneDistance = 1600.0

func main() {
	animation := scn.Animation{
		AnimationName:     "cylindrical_projection",
		Frames:            []scn.Frame{},
		Width:             800,
		Height:            600,
		WriteRawImageFile: false,
	}

	scene := scn.Scene{
		Camera:  getCamera(),
		Spheres: []scn.Sphere{},
		Discs:   []scn.Disc{},
	}

	sphereOrigin := vec3.T{0, ballRadius, 0}
	projectionOrigin := sphereOrigin
	projectionOrigin.Sub(&vec3.T{0, ballRadius, 0})
	projectionU := vec3.T{ballRadius, 0, ballRadius}
	projectionV := vec3.T{0, 2 * ballRadius, 0}
	projection := scn.NewCylindricalImageProjection("textures/planets/earth_daymap.jpg", projectionOrigin, projectionU, projectionV)

	sphere1 := scn.Sphere{
		Name:   "Textured sphere",
		Origin: sphereOrigin,
		Radius: ballRadius,
		Material: scn.Material{
			Color:      scn.Color{R: 1, G: 1, B: 1},
			Emission:   &scn.Black,
			Projection: &projection,
		},
	}

	scene.Spheres = append(scene.Spheres, sphere1)

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

func getCamera() scn.Camera {
	origin := vec3.T{0, 100, -200}

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
