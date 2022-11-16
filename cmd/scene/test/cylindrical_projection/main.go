package main

import (
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var ballRadius float64 = 30

var amountSamples = 5
var lensRadius float64 = 0
var viewPlaneDistance = 1600.0

func main() {
	animation := scn.Animation{
		AnimationName:     "cylindrical_projection",
		Frames:            []scn.Frame{},
		Width:             800,
		Height:            600,
		WriteRawImageFile: false,
	}

	scene := scn.SceneNode{
		Spheres: []*scn.Sphere{},
		Discs:   []*scn.Disc{},
	}

	sphereOrigin := vec3.T{0, 0, 0}
	projectionOrigin := sphereOrigin
	projectionOrigin.Sub(&vec3.T{0, ballRadius, 0})

	projectionU := vec3.T{0, 0, ballRadius}
	projectionV := vec3.T{0, 2 * ballRadius, 0}

	projection := scn.NewCylindricalImageProjection("textures/planets/earth_daymap.jpg", projectionOrigin, projectionU, projectionV)
	//projection := scn.NewCylindricalImageProjection("textures/uv.png", projectionOrigin, projectionU, projectionV)

	sphere1 := scn.Sphere{
		Name:   "Textured sphere",
		Origin: &sphereOrigin,
		Radius: ballRadius,
		Material: &scn.Material{
			Color:      &color.Color{R: 1, G: 1, B: 1},
			Emission:   &color.Black,
			Projection: &projection,
		},
	}

	scene.Spheres = append(scene.Spheres, &sphere1)

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
	}
}
