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
var maxRecursionDepth = 5
var amountSamples = 256 * 16
var lensRadius float64 = 0
var antiAlias = true

var viewPlaneDistance = 1500.0
var cameraDistanceFactor = 1.0

var imageWidth = 800
var imageHeight = 500
var magnification = 1.0

var roofHeight = ballRadius * 3.0

func main() {
	animation := scn.Animation{
		AnimationName:     animationName,
		Frames:            []scn.Frame{},
		Width:             int(float64(imageWidth) * magnification),
		Height:            int(float64(imageHeight) * magnification),
		WriteRawImageFile: true,
	}

	openBox := getBoxWalls()
	openBox.Scale(&vec3.Zero, &vec3.T{2 * ballRadius * 3, roofHeight, 3 * ballRadius * 3})
	openBox.Translate(&vec3.T{-ballRadius * 3, 0, -2 * ballRadius * 3})

	lampEmission := color.Color{R: 0.9, G: 0.9, B: 0.9}
	lampEmission.Multiply(8.0)
	lampHeight := roofHeight - 0.1
	lampSize := ballRadius * 2
	lamp := scn.FacetStructure{
		Material: &scn.Material{
			Emission:      &lampEmission,
			RayTerminator: true,
		},
		Facets: getFlatRectangleFacets(
			&vec3.T{-lampSize, lampHeight, +lampSize},
			&vec3.T{+lampSize, lampHeight, +lampSize},
			&vec3.T{+lampSize, lampHeight, -lampSize},
			&vec3.T{-lampSize, lampHeight, -lampSize},
		),
	}

	scene := scn.SceneNode{
		Spheres:         []*scn.Sphere{},
		FacetStructures: []*scn.FacetStructure{openBox, &lamp},
	}

	sphere1 := scn.Sphere{
		Name:   "Right sphere",
		Origin: vec3.T{ballRadius + (ballRadius / 2), ballRadius, 0},
		Radius: ballRadius,
		Material: &scn.Material{
			Color: color.Color{R: 0.9, G: 0.9, B: 0.9},
		},
	}

	sphere2 := scn.Sphere{
		Name:   "Left sphere",
		Origin: vec3.T{-(ballRadius + (ballRadius / 2)), ballRadius, 0},
		Radius: ballRadius,
		Material: &scn.Material{
			Color: color.Color{R: 0.9, G: 0.9, B: 0.9},
		},
	}

	scene.Spheres = append(scene.Spheres, &sphere1)
	scene.Spheres = append(scene.Spheres, &sphere2)

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
	cameraOrigin := vec3.T{0, ballRadius * 2, -15 * ballRadius}
	cameraOrigin.Scale(cameraDistanceFactor)

	focusPoint := vec3.T{0, ballRadius, 0}
	heading := focusPoint.Subed(&cameraOrigin)
	focalDistance := heading.Length()

	return scn.Camera{
		Origin:            &cameraOrigin,
		Heading:           &heading,
		ViewUp:            &vec3.T{0, 1, 0},
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

func getBoxWalls() *scn.FacetStructure {
	//roofTexture := scn.NewParallelImageProjection("textures/uv.png", vec3.T{0, ballRadius * 6, 0}, vec3.T{ballRadius, 0, 0}, vec3.T{0, 0, ballRadius})
	//floorTexture := scn.NewParallelImageProjection("textures/uv.png", vec3.T{0, 0, 0}, vec3.T{ballRadius, 0, 0}, vec3.T{0, 0, ballRadius})

	boxP1 := vec3.T{1, 1, 0} // Top right close            3----------2
	boxP2 := vec3.T{1, 1, 1} // Top right away            /          /|
	boxP3 := vec3.T{0, 1, 1} // Top left away            /          / |
	boxP4 := vec3.T{0, 1, 0} // Top left close          4----------1  |
	boxP5 := vec3.T{1, 0, 0} // Bottom right close      | (7)      |  6
	boxP6 := vec3.T{1, 0, 1} // Bottom right away       |          | /
	boxP7 := vec3.T{0, 0, 1} // Bottom left away        |          |/
	boxP8 := vec3.T{0, 0, 0} // Bottom left close       8----------5

	boxMaterial := scn.Material{
		Color: color.Color{R: 0.90, G: 0.90, B: 0.90},
	}

	leftWallMaterial := scn.Material{
		Color: color.Color{R: 0.85, G: 0.20, B: 0.20},
	}

	rightWallMaterial := scn.Material{
		Color: color.Color{R: 0.20, G: 0.20, B: 0.85},
	}

	box := scn.FacetStructure{
		Name: "Open box",
		FacetStructures: []*scn.FacetStructure{
			{
				Name:     "Roof",
				Material: &boxMaterial,
				Facets:   getFlatRectangleFacets(&boxP1, &boxP2, &boxP3, &boxP4),
			},
			{
				Name:     "Floor",
				Material: &boxMaterial,
				Facets:   getFlatRectangleFacets(&boxP8, &boxP7, &boxP6, &boxP5),
			},
			{
				Name:     "Back wall",
				Material: &boxMaterial,
				Facets:   getFlatRectangleFacets(&boxP6, &boxP7, &boxP3, &boxP2),
			},
			{
				Name:     "Right side wall",
				Material: &rightWallMaterial,
				Facets:   getFlatRectangleFacets(&boxP6, &boxP2, &boxP1, &boxP5),
			},
			{
				Name:     "Left side wall",
				Material: &leftWallMaterial,
				Facets:   getFlatRectangleFacets(&boxP7, &boxP8, &boxP4, &boxP3),
			},
		},
	}

	return &box
}

func getFlatRectangleFacets(p1, p2, p3, p4 *vec3.T) []*scn.Facet {
	v1 := p2.Subed(p1)
	v2 := p3.Subed(p1)
	normal := vec3.Cross(&v1, &v2)
	normal.Normalize()

	return []*scn.Facet{
		{
			Vertices: []*vec3.T{p1, p2, p4},
			Normal:   &normal,
		},
		{
			Vertices: []*vec3.T{p4, p2, p3},
			Normal:   &normal,
		},
	}
}
