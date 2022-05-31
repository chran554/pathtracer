package main

import (
	"encoding/json"
	"fmt"
	"os"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
	"strconv"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "reflective_test"

var ballRadius float64 = 20

var renderType = scn.Pathtracing

//var renderType = scn.Raycasting
var maxRecursionDepth = 5
var amountSamples = 10 * 1024
var lensRadius float64 = 2
var antiAlias = true

var viewPlaneDistance = 4000.0
var cameraDistanceFactor = 2.0

var lampEmissionFactor = 8.0

var imageWidth = 1600
var imageHeight = 500
var magnification = 1.0

func main() {
	animation := scn.Animation{
		AnimationName:     animationName,
		Frames:            []scn.Frame{},
		Width:             int(float64(imageWidth) * magnification),
		Height:            int(float64(imageHeight) * magnification),
		WriteRawImageFile: false,
	}

	scene := scn.SceneNode{
		Spheres: []*scn.Sphere{},
		Discs:   getBoxWalls(),
	}

	amountSpheres := 5
	sphereSpread := ballRadius * 2.0 * (float64(amountSpheres) + 1)
	sphereCC := sphereSpread / float64(amountSpheres)

	for i := 0; i <= amountSpheres; i++ {
		reflectiveness := float64(i) / float64(amountSpheres)
		sphere := scn.Sphere{
			Name:   "Sphere with reflectiveness of " + strconv.Itoa(i),
			Origin: vec3.T{-sphereSpread/2.0 + float64(i)*sphereCC, ballRadius, 0},
			Radius: ballRadius,
			Material: scn.Material{
				Color:      color.Color{R: 1, G: 1, B: 1},
				Glossiness: reflectiveness,
			},
		}
		scene.Spheres = append(scene.Spheres, &sphere)
	}

	lampEmission := color.White.Copy()
	lampEmission.Multiply(float32(lampEmissionFactor))
	lampLeft := scn.Sphere{
		Name:   "Lamp left",
		Origin: vec3.T{-0.3333 * sphereSpread, ballRadius*3 + ballRadius*2*0.75, -ballRadius * 3},
		Radius: ballRadius * 2,
		Material: scn.Material{
			Color:    color.Color{R: 1, G: 1, B: 1},
			Emission: &lampEmission,
		},
	}

	lampMiddle := scn.Sphere{
		Name:   "Lamp middle",
		Origin: vec3.T{0.0, ballRadius*3 + ballRadius*2*0.75, -ballRadius * 3},
		Radius: ballRadius * 2,
		Material: scn.Material{
			Color:    color.Color{R: 1, G: 1, B: 1},
			Emission: &lampEmission,
		},
	}

	lampRight := scn.Sphere{
		Name:   "Lamp right",
		Origin: vec3.T{0.3333 * sphereSpread, ballRadius*3 + ballRadius*2*0.75, -ballRadius * 3},
		Radius: ballRadius * 2,
		Material: scn.Material{
			Color:    color.Color{R: 1, G: 1, B: 1},
			Emission: &lampEmission,
		},
	}

	scene.Spheres = append(scene.Spheres, &lampLeft)
	scene.Spheres = append(scene.Spheres, &lampMiddle)
	scene.Spheres = append(scene.Spheres, &lampRight)

	camera := getCamera()

	frame := scn.Frame{
		Filename:   animation.AnimationName,
		FrameIndex: 0,
		Camera:     &camera,
		SceneNode:  &scene,
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

func getBoxWalls() []*scn.Disc {
	//roofTexture := scn.NewParallelImageProjection("textures/uv.png", vec3.T{0, ballRadius * 6, 0}, vec3.T{ballRadius, 0, 0}, vec3.T{0, 0, ballRadius})
	floorTexture := scn.NewParallelImageProjection("textures/tilesf4.jpeg", vec3.T{0, 0, 0}, vec3.T{ballRadius * 4, 0, 0}, vec3.T{0, 0, ballRadius * 4})

	floor := scn.Disc{
		Name:   "Floor",
		Origin: vec3.T{0, 0, 0},
		Normal: vec3.T{0, 1, 0},
		Radius: 600,
		Material: scn.Material{
			Color:      color.Color{R: 1, G: 1, B: 1},
			Projection: &floorTexture,
		},
	}

	roof := scn.Disc{
		Name:   "Roof",
		Origin: vec3.T{0, ballRadius * 3, 0},
		Normal: vec3.T{0, -1, 0},
		Radius: 600,
		Material: scn.Material{
			Color: color.Color{R: 0.9, G: 1, B: 0.95},
			//			Projection: &roofTexture,
		},
	}

	backWallTexture := scn.NewParallelImageProjection("textures/bricks_yellow.png",
		vec3.T{0, 0, 0},
		vec3.T{ballRadius * 9, 0, 0},
		vec3.T{0, ballRadius * 9, 0})

	backWall := scn.Disc{
		Name:   "Back wall",
		Origin: vec3.T{0, 0, ballRadius * 3},
		Normal: vec3.T{0, 0, -1},
		Radius: 600,
		Material: scn.Material{
			Color:      color.Color{R: 1, G: 1, B: 1},
			Projection: &backWallTexture,
		},
	}

	return []*scn.Disc{&floor, &roof, &backWall}
}
