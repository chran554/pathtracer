package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "testobject"

var amountAnimationFrames = 1 // 72 * 2

var imageWidth = 400
var imageHeight = 400
var magnification = 1.5

var amountSamples = 512 * 10
var maxRecursionDepth = 4

var cameraAperture = 0.0
var cameraZoom = 1.0

var lampIntensity = 23.0
var lightIntensityFactor = 0.25
var sphereRadius = 1.0
var tesselationLevel = 2
var useVertexNormals = true

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

	lamp := scene.NewSphere(&vec3.T{-2, sphereRadius * 3, -2}, 0.5, scene.NewMaterial().
		N("lamp").
		E(color.KelvinTemperatureColor2(5000), lampIntensity, true))

	textureSphere, err := floatimage.EmptyPlaceholderImage("textures/equirectangular/world_map_latlonlines_equirectangular.jpeg")
	if err != nil {
		panic(err)
	}
	testObject := obj.NewTessellatedSphere(tesselationLevel, useVertexNormals)
	testObject.Material = scene.NewMaterial().N("tessellated sphere").
		C(color.White).
		SP(textureSphere, testObject.Bounds.Center(), vec3.UnitX, vec3.UnitY)
	testObject.FlipX(&vec3.Zero)
	testObject.ScaleUniform(&vec3.Zero, sphereRadius)
	testObject.Translate(&vec3.T{0, -testObject.Bounds.Ymin, 0})

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	for frameIndex := 0; frameIndex < amountAnimationFrames; frameIndex++ {
		// animationProgress := float64(frameIndex) / float64(amountAnimationFrames)

		cameraPosition := &vec3.T{0, sphereRadius * 3, -sphereRadius * 6}

		cameraAim := testObject.Bounds.Center()
		focusPosition := testObject.Bounds.Center().Added(&vec3.T{0, 0, -testObject.Bounds.SizeZ() / 2})
		cameraFocusVector := focusPosition.Subed(cameraPosition)

		camera := scene.NewCamera(cameraPosition, cameraAim, amountSamples, magnification).
			D(maxRecursionDepth).
			A(cameraAperture, nil).
			F(cameraFocusVector.Length()).
			V(800 * cameraZoom)

		scn := scene.NewSceneNode().FS(cornellBox, testObject).S(lamp)

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
