package main

import (
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var ballRadius float64 = 30

//var cameraOrigin = vec3.T{0, 0, -200} // Looking straight on
var cameraOrigin = vec3.T{0, 200, -200} // Looking straight on, slightly from above
//var cameraOrigin = vec3.T{0, -200, -200} // Looking straight on, slightly from below

var amountSamples = 256
var lensRadius float64 = 0
var antiAlias = true
var viewPlaneDistance = 1600.0
var magnification = 1.5

func main() {
	animation := scn.Animation{
		AnimationName:     "spherical_projection",
		Frames:            []scn.Frame{},
		Width:             int(float64(1200) * magnification),
		Height:            int(float64(600) * magnification),
		WriteRawImageFile: false,
	}

	scene := scn.SceneNode{
		Spheres: []*scn.Sphere{},
		Discs:   []*scn.Disc{},
	}

	sphere2Origin := vec3.T{0, 0, 0}
	sphere1Origin := sphere2Origin.Added(&vec3.T{-ballRadius * 2.2, 0, 0})
	sphere3Origin := sphere2Origin.Added(&vec3.T{ballRadius * 2.2, 0, 0})

	projection1Origin := sphere1Origin
	projection2Origin := sphere2Origin
	projection3Origin := sphere3Origin

	projectionU := vec3.T{0, 0, -ballRadius}
	projectionV := vec3.T{0, ballRadius, 0}

	projection1 := scn.NewSphericalImageProjection("textures/planets/earth_daymap.jpg", projection1Origin, projectionU.Inverted(), projectionV)
	projection2 := scn.NewSphericalImageProjection("textures/checkered 360x180 with lines.png", projection2Origin, projectionU, projectionV)
	projection3 := scn.NewSphericalImageProjection("textures/equirectangular/2560px-Plate_Carrée_with_Tissot's_Indicatrices_of_Distortion.svg.png", projection3Origin, projectionU.Inverted(), projectionV)

	sphere1 := scn.Sphere{
		Name:   "Textured sphere - Earth",
		Origin: sphere1Origin,
		Radius: ballRadius,
		Material: scn.Material{
			Color:      color.Color{R: 1, G: 1, B: 1},
			Emission:   &color.Black,
			Projection: &projection1,
		},
	}

	sphere2 := scn.Sphere{
		Name:   "Textured sphere - checkered",
		Origin: sphere2Origin,
		Radius: ballRadius,
		Material: scn.Material{
			Color:      color.Color{R: 1, G: 1, B: 1},
			Emission:   &color.Black,
			Projection: &projection2,
		},
	}

	sphere3 := scn.Sphere{
		Name:   "Textured sphere - Tissot's_Indicatrices_of_Distortion",
		Origin: sphere3Origin,
		Radius: ballRadius,
		Material: scn.Material{
			Color:      color.Color{R: 1, G: 1, B: 1},
			Emission:   &color.Black,
			Projection: &projection3,
		},
	}

	scene.Spheres = append(scene.Spheres, &sphere1)
	scene.Spheres = append(scene.Spheres, &sphere2)
	scene.Spheres = append(scene.Spheres, &sphere3)

	camera := getCamera()
	frame := scn.Frame{
		Filename:   animation.AnimationName,
		FrameIndex: 0,
		Camera:     &camera,
		SceneNode:  &scene,
	}

	animation.Frames = append(animation.Frames, frame)

	anm.WriteAnimationToFile(animation)
}

func getCamera() scn.Camera {
	heading := vec3.T{-cameraOrigin[0], -cameraOrigin[1], -cameraOrigin[2]}
	focalDistance := heading.Length()

	return scn.Camera{
		Origin:            cameraOrigin,
		Heading:           heading,
		ViewUp:            vec3.T{0, 1, 0},
		ViewPlaneDistance: viewPlaneDistance,
		LensRadius:        lensRadius,
		FocalDistance:     focalDistance,
		Samples:           amountSamples,
		AntiAlias:         antiAlias,
		Magnification:     magnification,
	}
}
