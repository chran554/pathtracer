package main

import (
	"fmt"
	"math"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "lamp_post"

var environmentEnvironMap = "textures/equirectangular/sunset horizon 2800x1400.jpg"
var environmentRadius = 500.0 * 1000.0
var environmentEmissionFactor = float32(1.0)

var amountFrames = 1

var imageWidth = 200
var imageHeight = 200
var magnification = 5.0

var renderType = scn.Pathtracing
var amountSamples = 512 * 2 * 4 * 2 * 2 * 2 / 3
var maxRecursion = 8

var viewPlaneDistance = 800.0
var apertureSize = 1.0 // 2.0

func main() {
	animation := getAnimation(int(float64(imageWidth)*magnification), int(float64(imageHeight)*magnification))
	animation.WriteRawImageFile = true

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		environmentSphere := addEnvironmentMapping(environmentEnvironMap)

		// Ground
		groundProjection := scn.NewParallelImageProjection("textures/ground/soil-cracked.png", vec3.Zero, vec3.UnitX.Scaled(150), vec3.UnitZ.Scaled(150))
		ground := scn.Disc{
			Name:     "Ground",
			Origin:   &vec3.Zero,
			Normal:   &vec3.UnitY,
			Radius:   environmentRadius,
			Material: &scn.Material{Name: "Ground material", Color: &color.White, Emission: &color.Black, Glossiness: 0.0, Roughness: 1.0, Projection: &groundProjection},
		}

		// Gopher
		gopher := obj.NewGopher(&vec3.T{50, 50, 50})
		gopher.RotateY(&vec3.Zero, math.Pi*10.0/10.0)
		gopher.Translate(&vec3.T{75, 0, 100})
		gopher.UpdateBounds()
		gopherBounds := gopher.Bounds

		// Kerosine lamp
		kerosineLamp := obj.NewKerosineLamp(&vec3.T{40, 40, 40})
		kerosineLamp.RotateY(&vec3.Zero, -math.Pi*4.0/10.0)
		kerosineLamp.Translate(&vec3.T{gopherBounds.Center()[0] + gopherBounds.SizeX()/2, 0, gopherBounds.Center()[2] - gopherBounds.SizeY()/2})
		kerosineLamp.UpdateBounds()

		// Lamp post
		lampPost := obj.NewLampPost(&vec3.T{200, 200, 200})

		cameraOrigin := gopher.Bounds.Center().Add(&vec3.T{-gopherBounds.SizeY(), -gopherBounds.SizeY() * 0.1, -250})
		cameraFocusPoint := gopherBounds.Center().Add(&vec3.T{0.0, 0.0, -(gopherBounds.SizeZ() / 2) * 0.9})
		//cameraFocusPoint := gopherBounds.Center().Add(&vec3.T{0.0, gopherBounds.SizeY() * 2, 0.0})
		camera := getCamera(magnification, animationProgress, cameraOrigin, cameraFocusPoint)

		scene := scn.SceneNode{
			Spheres:         []*scn.Sphere{&environmentSphere},
			Discs:           []*scn.Disc{&ground},
			ChildNodes:      []*scn.SceneNode{},
			FacetStructures: []*scn.FacetStructure{gopher, lampPost, kerosineLamp},
		}

		frame := scn.Frame{
			Filename:   animation.AnimationName + "_" + fmt.Sprintf("%06d", frameIndex),
			FrameIndex: frameIndex,
			Camera:     &camera,
			SceneNode:  &scene,
		}

		animation.Frames = append(animation.Frames, frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

func addEnvironmentMapping(filename string) scn.Sphere {
	origin := vec3.T{0, 0, 0}

	projection := scn.ImageProjection{
		ProjectionType: scn.ProjectionTypeSpherical,
		ImageFilename:  filename,
		Origin:         &origin,
		U:              &vec3.T{-0.2, 0, -1},
		V:              &vec3.T{0, 1, 0},
		RepeatU:        true,
		RepeatV:        true,
		FlipU:          false,
		FlipV:          false,
	}

	material := scn.Material{
		Color:         &color.Color{R: 1.0, G: 1.0, B: 1.0},
		Emission:      (&color.Color{R: 1.0, G: 1.0, B: 1.0}).Multiply(environmentEmissionFactor),
		RayTerminator: true,
		Projection:    &projection,
	}

	sphere := scn.Sphere{
		Name:     "Environment mapping",
		Origin:   &origin,
		Radius:   environmentRadius,
		Material: &material,
	}

	return sphere
}

func getAnimation(width int, height int) *scn.Animation {
	animation := &scn.Animation{
		AnimationName:     animationName,
		Frames:            []scn.Frame{},
		Width:             width,
		Height:            height,
		WriteRawImageFile: false,
	}
	return animation
}

func getCamera(magnification float64, progress float64, cameraOrigin *vec3.T, cameraFocusPoint *vec3.T) scn.Camera {

	// Point heading towards center of sphere ring (heading vector starts in camera origin)
	heading := cameraFocusPoint.Subed(cameraOrigin)

	focusDistance := heading.Length()

	return scn.Camera{
		Origin:            cameraOrigin,
		Heading:           &heading,
		ViewUp:            &vec3.T{0, 1, 0},
		ViewPlaneDistance: viewPlaneDistance,
		ApertureSize:      apertureSize,
		FocusDistance:     focusDistance,
		Samples:           amountSamples,
		AntiAlias:         true,
		Magnification:     magnification,
		RenderType:        renderType,
		RecursionDepth:    maxRecursion,
	}
}
