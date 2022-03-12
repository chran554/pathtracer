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
var antiAlias = true
var viewPlaneDistance = 1600.0

func main() {
	animation := scn.Animation{
		AnimationName:     "spherical_projection",
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

	sphereOrigin := vec3.T{0, 0, 0}
	projectionOrigin := sphereOrigin.Subed(&vec3.T{0, ballRadius, 0})

	projectionU := vec3.T{0, 0, ballRadius}
	projectionV := vec3.T{0, 2.0 * ballRadius, 0}

	//sphericalProjection := scn.NewSphericalImageProjection("textures/propeller_brick.png", projectionOrigin, projectionU, projectionV)

	//sphericalProjection := scn.NewSphericalImageProjection("textures/uv.png", projectionOrigin, projectionU, projectionV)
	//cylindricalProjection := scn.NewCylindricalImageProjection("textures/uv.png", projectionOrigin, projectionU, projectionV)

	//sphericalProjection := scn.NewSphericalImageProjection("textures/equirectangular/Blue_Marble_3840px-2002.png", projectionOrigin, projectionU, projectionV)
	//projection := scn.NewSphericalImageProjection("textures/equirectangular/2560px-Plate_Carrée_with_Tissot's_Indicatrices_of_Distortion.svg.png", projectionOrigin, projectionU, projectionV)
	projection := scn.NewCylindricalImageProjection("textures/equirectangular/world_map_latlonlines_equirectangular.jpeg", projectionOrigin, projectionU, projectionV)
	//sphericalProjection := scn.NewSphericalImageProjection("textures/equirectangular/bathroom.jpeg", projectionOrigin, projectionU, projectionV)
	//sphericalProjection := scn.NewSphericalImageProjection("textures/planets/earth_daymap.jpg", projectionOrigin, projectionU, projectionV)
	//cylindricalProjection := scn.NewCylindricalImageProjection("textures/planets/earth_daymap.jpg", projectionOrigin, projectionU, projectionV)
	//cylindricalProjection := scn.NewCylindricalImageProjection("textures/planets/earth_daymap.jpg", projectionOrigin, projectionU, projectionV)

	sphere1 := scn.Sphere{
		Name:   "Textured sphere",
		Origin: sphereOrigin,
		Radius: ballRadius,
		Material: scn.Material{
			Color:      color.Color{R: 1, G: 1, B: 1},
			Emission:   &color.Black,
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

	anm.WriteAnimationToFile(animation)
}

func getCamera() scn.Camera {
	//origin := vec3.T{0, -200, -200}
	origin := vec3.T{0, 200, -200}
	//origin := vec3.T{0, 0, -200}

	heading := vec3.T{-origin[0], -origin[1], -origin[2]}
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
