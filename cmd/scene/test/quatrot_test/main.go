package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "quatrot_test"

var amountAnimationFrames = 180 // 72 * 2

var imageWidth = 400
var imageHeight = 400
var magnification = 0.5

var amountSamples = 512 * 2
var maxRecursionDepth = 4

var cameraAperture = 0.0
var cameraZoom = 1.0

var lampIntensity = 10.0
var lightIntensityFactor = 0.25
var sphereRadius = 1.0
var tesselationLevel = 2
var useVertexNormals = false

func main() {
	textureGround, err := floatimage.EmptyPlaceholderImage("textures/floor/7451-diffuse 02 low contrast.png")
	if err != nil {
		panic(err)
	}

	// Cornell box
	roomScale := 2.0
	cornellBox := obj.NewWhiteCornellBox(&vec3.T{4 * roomScale, sphereRadius * 4, 4 * roomScale}, true, lightIntensityFactor) // cm, as units. I.e. a 5x3x5m room
	floorMaterial := cornellBox.GetFirstMaterialByName("Floor")
	floorMaterial.M(0.01, 0.3).PP(textureGround, &vec3.Zero, vec3.T{1, 0, 0}, vec3.T{0, 0, 1})
	floorMaterial.FresnelMaxGlossiness = 0.15

	lamp := scene.NewSphere(&vec3.T{-1.5, sphereRadius * 3, -2}, 0.5, scene.NewMaterial().
		N("lamp").
		E(color.KelvinTemperatureColor2(5000), lampIntensity, true))

	//textureSphere, err := floatimage.EmptyPlaceholderImage("textures/equirectangular/world_map_latlonlines_equirectangular.jpeg")
	//if err != nil {
	//	panic(err)
	//}

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	for frameIndex := 0; frameIndex < amountAnimationFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountAnimationFrames)

		cube := createCube(8.0, sphereRadius)
		cube.Translate(&vec3.T{0, -cube.Bounds.Ymin + cube.Bounds.SizeY()/2, 0})

		rotationOrigin := &vec3.T{cube.Bounds.Xmin, cube.Bounds.Ymin, cube.Bounds.Zmin}
		rotationAxis := &vec3.T{1, 1, 1}
		rotationAxis.Normalize()

		cube.RotateAxis(rotationOrigin, rotationAxis, util.DegToRad(360*animationProgress))

		cameraPosition := &vec3.T{0, sphereRadius * 3, -sphereRadius * 6}

		cameraAim := cube.Bounds.Center()
		focusPosition := cube.Bounds.Center().Added(&vec3.T{0, 0, -cube.Bounds.SizeZ() / 2})
		cameraFocusVector := focusPosition.Subed(cameraPosition)

		camera := scene.NewCamera(cameraPosition, cameraAim, amountSamples, magnification).
			D(maxRecursionDepth).
			A(cameraAperture, nil).
			F(cameraFocusVector.Length()).
			V(800 * cameraZoom)

		scn := scene.NewSceneNode().FS(cornellBox, cube).S(lamp)

		fi := -1
		if amountAnimationFrames > 1 {
			fi = frameIndex
		}

		frame := scene.NewFrame(animationName, fi, camera, scn)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err = renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func createCube(flattness float64, sphereRadius float64) *scene.FacetStructure {
	flattnessFactor := 1 / flattness

	lo, hi, mx := 0.3, 0.6, 0.9
	color1 := color.White.Fade(color.NewColor(mx, lo, lo), 0.75)
	color2 := color.White.Fade(color.NewColor(lo, mx, lo), 0.75)
	color3 := color.White.Fade(color.NewColor(lo, lo, mx), 0.75)
	color4 := color.White.Fade(color.NewColor(hi, hi, lo), 0.75)
	color5 := color.White.Fade(color.NewColor(hi, lo, hi), 0.75)
	color6 := color.White.Fade(color.NewColor(lo, hi, hi), 0.75)

	sphere1 := createButtonSphere(color1, &vec3.T{1, flattnessFactor, 1}, &vec3.T{0, sphereRadius, 0})
	sphere2 := createButtonSphere(color2, &vec3.T{1, flattnessFactor, 1}, &vec3.T{0, -sphereRadius, 0})
	sphere3 := createButtonSphere(color3, &vec3.T{1, 1, flattnessFactor}, &vec3.T{0, 0, sphereRadius})
	sphere4 := createButtonSphere(color4, &vec3.T{1, 1, flattnessFactor}, &vec3.T{0, 0, -sphereRadius})
	sphere5 := createButtonSphere(color5, &vec3.T{flattnessFactor, 1, 1}, &vec3.T{sphereRadius, 0, 0})
	sphere6 := createButtonSphere(color6, &vec3.T{flattnessFactor, 1, 1}, &vec3.T{-sphereRadius, 0, 0})

	cube := &scene.FacetStructure{Name: "cube", FacetStructures: []*scene.FacetStructure{sphere1, sphere2, sphere3, sphere4, sphere5, sphere6}}
	cube.UpdateBounds()

	return cube
}

func createButtonSphere(color1 *color.Color, scale *vec3.T, origin *vec3.T) *scene.FacetStructure {
	sphere1 := obj.NewTessellatedSphere(tesselationLevel, useVertexNormals)
	sphere1.Material = scene.NewMaterial().N("button").C(color1) //.SP(textureSphere, &vec3.T{}, vec3.UnitX, vec3.UnitY)
	sphere1.Scale(&vec3.Zero, scale)
	sphere1.Translate(origin)
	return sphere1
}

func RotationFromTo(v1, v2 *vec3.T) (axis *vec3.T, angle float64, ok bool) {
	if v1.Length() == 0 || v2.Length() == 0 {
		return &vec3.Zero, 0, false
	}

	u := v1.Normalized()
	v := v2.Normalized()

	axis_ := vec3.Cross(&u, &v)
	axis = &axis_
	s := axis.Length()
	c := vec3.Dot(&u, &v)

	const eps = 1e-12
	if s < eps {
		if c > 0 {
			// already parallel
			return &vec3.T{0, 1, 0}, 0, true
		}
		// opposite: pick any perpendicular axis
		tmp := &vec3.T{1, 0, 0}
		if math.Abs(u[0]) > 0.9 {
			tmp = &vec3.T{0, 1, 0}
		}
		axis_ = vec3.Cross(&u, tmp)
		axis = &axis_
		axis.Normalize()
		return axis, math.Pi, true
	}

	axis.Scale(1.0 / s)      // normalize axis
	angle = math.Atan2(s, c) // stable angle
	return axis, angle, true
}
