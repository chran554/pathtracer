package main

import (
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
	"strconv"

	"github.com/ungerik/go3d/float64/vec3"
)

// https://en.wikipedia.org/wiki/List_of_refractive_indices
const (
	vacuumRefractionIndex = 1.0
	airRefractionIndex    = 1.000273
	waterRefractionIndex  = 1.333
	ice                   = 1.31
	pyrexGlass            = 1.470
	acrylicPlastic        = 1.495
	glassRefractionIndex  = 1.50
	windowGlass           = 1.52
	petPlastic            = 1.5750
	sapphire              = 1.762
	diamond               = 2.417
)

var animationName = "refraction_test"

var ballRadius float64 = 20

var renderType = scn.Pathtracing

// var renderType = scn.Raycasting
var maxRecursionDepth = 5
var amountSamples = 218
var lensRadius float64 = 2

var viewPlaneDistance = 4000.0
var cameraDistanceFactor = 2.0

var lampEmissionFactor = 11.0

var imageWidth = 1600
var imageHeight = 500
var magnification = 0.25

var sphereRefractionIndex = glassRefractionIndex

func main() {
	width := int(float64(imageWidth) * magnification)
	height := int(float64(imageHeight) * magnification)

	if width%2 == 1 {
		width++
	}
	if height%2 == 1 {
		height++
	}

	animation := scn.Animation{
		AnimationName:     animationName,
		Frames:            []scn.Frame{},
		Width:             width,
		Height:            height,
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
		transparency := float64(i) / float64(amountSpheres)
		sphere := scn.Sphere{
			Name:   "Glass sphere with transparency " + strconv.Itoa(i),
			Origin: &vec3.T{-sphereSpread/2.0 + float64(i)*sphereCC, ballRadius, 0},
			Radius: ballRadius,
			Material: &scn.Material{
				Color:           &color.Color{R: 0.97, G: 0.99, B: 1},
				Glossiness:      0.9,
				RefractionIndex: sphereRefractionIndex,
				Transparency:    transparency,
			},
		}
		scene.Spheres = append(scene.Spheres, &sphere)
	}

	lampEmission := color.White.Copy()
	lampEmission.Multiply(float32(lampEmissionFactor))
	lampLeft := scn.Sphere{
		Name:   "Lamp left",
		Origin: &vec3.T{-0.5 * sphereSpread, ballRadius*3 + ballRadius*2*0.75, -ballRadius * 3},
		Radius: ballRadius * 2,
		Material: &scn.Material{
			Color:    &color.Color{R: 1, G: 1, B: 1},
			Emission: lampEmission,
		},
	}

	lampMiddle := scn.Sphere{
		Name:   "Lamp middle",
		Origin: &vec3.T{0.0, ballRadius*3 + ballRadius*2*0.75, -ballRadius * 3},
		Radius: ballRadius * 2,
		Material: &scn.Material{
			Color:    &color.Color{R: 1, G: 1, B: 1},
			Emission: lampEmission,
		},
	}

	lampRight := scn.Sphere{
		Name:   "Lamp right",
		Origin: &vec3.T{0.5 * sphereSpread, ballRadius*3 + ballRadius*2*0.75, -ballRadius * 3},
		Radius: ballRadius * 2,
		Material: &scn.Material{
			Color:    &color.Color{R: 1, G: 1, B: 1},
			Emission: lampEmission,
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

	anm.WriteAnimationToFile(animation, false)
}

func getCamera() scn.Camera {
	origin := vec3.T{0, ballRadius, -400}
	origin.Scale(cameraDistanceFactor)

	heading := vec3.T{-origin[0], -(origin[1] - ballRadius), -origin[2]}
	focalDistance := heading.Length()

	return scn.Camera{
		Origin:            &origin,
		Heading:           &heading,
		ViewUp:            &vec3.T{0, 1, 0},
		ViewPlaneDistance: viewPlaneDistance,
		ApertureSize:      lensRadius,
		FocusDistance:     focalDistance,
		Samples:           amountSamples,
		AntiAlias:         true,
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
		Origin: &vec3.T{0, 0, 0},
		Normal: &vec3.T{0, 1, 0},
		Radius: 600,
		Material: &scn.Material{
			Color:      &color.Color{R: 1, G: 1, B: 1},
			Projection: &floorTexture,
		},
	}

	roof := scn.Disc{
		Name:   "Roof",
		Origin: &vec3.T{0, ballRadius * 5, 0},
		Normal: &vec3.T{0, -1, 0},
		Radius: 600,
		Material: &scn.Material{
			Color: &color.Color{R: 0.95, G: 1, B: 0.90},
			//			Projection: &roofTexture,
		},
	}

	backWallTexture := scn.NewParallelImageProjection("textures/silver_bricks.png",
		vec3.T{0, 0, 0},
		vec3.T{ballRadius * 3, 0, 0},
		vec3.T{0, ballRadius * 3, 0})

	backWall := scn.Disc{
		Name:   "Back wall",
		Origin: &vec3.T{0, 0, ballRadius * 3},
		Normal: &vec3.T{0, 0, -1},
		Radius: 600,
		Material: &scn.Material{
			Color:      &color.Color{R: 0.95, G: 0.93, B: 0.90},
			Projection: &backWallTexture,
		},
	}

	return []*scn.Disc{&floor, &roof, &backWall}
}
