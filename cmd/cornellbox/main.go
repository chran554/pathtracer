package main

import (
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "cornellbox"

var ballRadius float64 = 20

var renderType = scn.Pathtracing
var maxRecursionDepth = 6
var amountSamples = 2048
var lensRadius float64 = 2
var antiAlias = true

var viewPlaneDistance = 4000.0
var cameraDistanceFactor = 2.0

var imageWidth = 800
var imageHeight = 600
var magnification = 0.5

var roofHeight = ballRadius * 3.0

func main() {
	animation := scn.Animation{
		AnimationName:     animationName,
		Frames:            []scn.Frame{},
		Width:             int(float64(imageWidth) * magnification),
		Height:            int(float64(imageHeight) * magnification),
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
			Color: color.Color{R: 0.85, G: 0.9, B: 0.9},
		},
	}

	sphere2 := scn.Sphere{
		Name:   "Left sphere",
		Origin: vec3.T{-(ballRadius + (ballRadius / 2)), ballRadius, 0},
		Radius: ballRadius,
		Material: scn.Material{
			Color: color.Color{R: 0.9, G: 0.9, B: 0.85},
		},
	}

	lampEmission := color.Color{R: 0.9, G: 0.9, B: 0.85}
	lampEmission.Multiply(6.0)
	lampRadius := ballRadius * 6
	lamp := scn.Sphere{
		Name:   "Lamp",
		Origin: vec3.T{0, roofHeight + lampRadius*0.9, -ballRadius},
		Radius: lampRadius,
		Material: scn.Material{
			Color:    color.Color{R: 1, G: 1, B: 1},
			Emission: &lampEmission,
		},
	}

	scene.Spheres = append(scene.Spheres, sphere1)
	scene.Spheres = append(scene.Spheres, sphere2)
	scene.Spheres = append(scene.Spheres, lamp)

	frame := scn.Frame{
		Filename:   animation.AnimationName,
		FrameIndex: 0,
		Scene:      scene,
	}

	animation.Frames = append(animation.Frames, frame)

	anm.WriteAnimationToFile(animation)
}

func getCamera() scn.Camera {
	origin := vec3.T{0, ballRadius, -400}
	origin.Scale(cameraDistanceFactor)

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
		Magnification:     magnification,
		RenderType:        renderType,
		RecursionDepth:    maxRecursionDepth,
	}
}

func getBoxWalls() []scn.Disc {
	//roofTexture := scn.NewParallelImageProjection("textures/uv.png", vec3.T{0, ballRadius * 6, 0}, vec3.T{ballRadius, 0, 0}, vec3.T{0, 0, ballRadius})
	//floorTexture := scn.NewParallelImageProjection("textures/uv.png", vec3.T{0, 0, 0}, vec3.T{ballRadius, 0, 0}, vec3.T{0, 0, ballRadius})

	floor := scn.Disc{
		Name:   "Floor",
		Origin: vec3.T{0, 0, 0},
		Normal: vec3.T{0, 1, 0},
		Radius: 600,
		Material: scn.Material{
			Color: color.Color{R: 1, G: 1, B: 1},
			//			Projection: &floorTexture,
		},
	}

	roof := scn.Disc{
		Name:   "Roof",
		Origin: vec3.T{0, roofHeight, 0},
		Normal: vec3.T{0, -1, 0},
		Radius: 600,
		Material: scn.Material{
			Color: color.Color{R: 1, G: 1, B: 1},
			//			Projection: &roofTexture,
		},
	}

	rightWall := scn.Disc{
		Name:   "Right wall",
		Origin: vec3.T{ballRadius * 3, 0, 0},
		Normal: vec3.T{-1, 0, 0},
		Radius: 600,
		Material: scn.Material{
			Color: color.Color{R: 0.5, G: 0.5, B: 1},
		},
	}

	leftWall := scn.Disc{
		Name:   "Left wall",
		Origin: vec3.T{-ballRadius * 3, 0, 0},
		Normal: vec3.T{1, 0, 0},
		Radius: 600,
		Material: scn.Material{
			Color: color.Color{R: 1, G: 0.5, B: 0.5},
		},
	}

	backWall := scn.Disc{
		Name:   "Back wall",
		Origin: vec3.T{0, 0, ballRadius * 3},
		Normal: vec3.T{0, 0, -1},
		Radius: 600,
		Material: scn.Material{
			Color: color.Color{R: 1, G: 1, B: 1},
		},
	}

	return []scn.Disc{floor, roof, rightWall, leftWall, backWall}
}
