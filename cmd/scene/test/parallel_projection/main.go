package main

import (
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var discRadius float64 = 30

var amountSamples = 64
var lensRadius float64 = 0
var viewPlaneDistance = 1600.0
var magnification = 2.0

func main() {
	animation := scn.Animation{
		AnimationName:     "parallel_projection",
		Frames:            []scn.Frame{},
		Width:             int(800 * magnification),
		Height:            int(600 * magnification),
		WriteRawImageFile: false,
	}

	scene := scn.SceneNode{
		Spheres: []*scn.Sphere{},
		Discs:   []*scn.Disc{},
	}

	discOrigin := vec3.T{0, 0, 0}
	projectionOrigin := discOrigin

	projectionU := vec3.T{discRadius / 2.0, 0, 0}
	projectionV := vec3.T{0, discRadius / 2.0, 0}

	projection := scn.NewParallelImageProjection("textures/uv.png", projectionOrigin, projectionU, projectionV)
	//projection := scn.NewCylindricalImageProjection("textures/uv.png", projectionOrigin, projectionU, projectionV)

	disc := scn.Disc{
		Name:   "Textured disc",
		Origin: &discOrigin,
		Radius: discRadius,
		Normal: &vec3.T{0, 0, -1},
		Material: &scn.Material{
			Color:      &color.Color{R: 1, G: 1, B: 1},
			Emission:   &color.Black,
			Projection: &projection,
		},
	}

	scene.Discs = append(scene.Discs, &disc)

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
	origin := vec3.T{0, 0, -200}

	heading := vec3.T{-origin[0], -origin[1], -origin[2]}
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
	}
}
