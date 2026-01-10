package main

import (
	"fmt"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/quaternion"
	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "earth"

var amountAnimationFrames = 32 // 72 * 2 // TODO set to 1
var startFrameIndex = 0

var imageWidth = 640 * 1
var imageHeight = 480 * 1
var magnification = 2.0

var amountSamples = 512 * 2 * 12
var maxRecursionDepth = 16

var cameraAperture = 0.0
var cameraZoom = 1.0

const (
	kkilometer        = 1.0 // Use kilo-kilometer (kkm or mega-meter Mm) as unit for measures
	sunRadius         = 695 * kkilometer
	earthRadius       = 6.378 * kkilometer
	moonRadius        = 1.7374 * kkilometer
	moonEarthDistance = 384.400 * kkilometer
	sunEarthDistance  = 149_600 * kkilometer
	spaceRadius       = sunEarthDistance * 1000
)

func main() {
	textureMoon, err := floatimage.EmptyPlaceholderImage("textures/planets/moonmap4k.jpg")
	if err != nil {
		panic(err)
	}

	//textureSpace, err := floatimage.EmptyPlaceholderImage("textures/planets/environmentmap/Stellarium3.jpeg")
	textureSpace, err := floatimage.EmptyPlaceholderImage("textures/planets/environmentmap/starmap_2020_16k.png")
	if err != nil {
		panic(err)
	}

	solarSystemOrigin := &vec3.T{0, 0, 0}

	spaceMaterial := scene.NewMaterial().N("space")
	spaceMaterial.SP(textureSpace, copy(solarSystemOrigin), vec3.T{1, 0, 0}, vec3.T{0, 1, 0})
	spaceMaterial.Projection.Interpolation = floatimage.InterpolationBicubic
	spaceMaterial.E(color.White, 1.0, true)
	spaceMaterial.FresnelMaxGlossiness = 0.0
	spaceEnvironment := scene.NewSphere(copy(solarSystemOrigin), spaceRadius, spaceMaterial)

	sunMaterial := scene.NewMaterial().N("sun")
	//sunMaterial.SP(textureSun, copy(solarSystemOrigin), vec3.T{1, 0, 0}, vec3.T{0, 1, 0})
	sunMaterial.E(color.White, 75.0, true)
	sunMaterial.FresnelMaxGlossiness = 0.0
	sun := scene.NewSphere(copy(solarSystemOrigin), sunRadius*20, sunMaterial)
	sun.Translate(&vec3.T{0, 0, 0})

	earthPosition := &vec3.T{sunEarthDistance, 0, 0}
	earth := createEarth(earthRadius, earthPosition)

	moonMaterial := scene.NewMaterial().N("moon")
	moonMaterial.SP(textureMoon, copy(solarSystemOrigin), vec3.T{1, 0, 0}, vec3.T{0, 1, 0})
	//moonMaterial.E(color.White, 1.0, true)
	moonMaterial.FresnelMaxGlossiness = 0.0
	moon := scene.NewSphere(copy(solarSystemOrigin), moonRadius*3, moonMaterial)
	moon.Translate(&vec3.T{sunEarthDistance + moonEarthDistance, 0, 0})

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	for frameIndex := startFrameIndex; frameIndex < amountAnimationFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountAnimationFrames)

		//zoom := 7.5
		zoom := 1.0

		cameraPosition := earth.Bounds.Center().Add(&vec3.T{-earthRadius * 4.0 * zoom, earthRadius * zoom, -earthRadius * 0.30 * zoom})

		cameraPosition.Sub(earthPosition)
		cameraRotation := quaternion.FromYAxisAngle(util.DegToRad(-360 * animationProgress))
		cameraRotation.RotateVec3(cameraPosition)
		cameraPosition.Add(earthPosition)

		cameraAim := earth.Bounds.Center()
		focusPosition := earth.Bounds.Center()
		cameraFocusVector := focusPosition.Subed(cameraPosition)

		camera := scene.NewCamera(cameraPosition, cameraAim, amountSamples, magnification).
			D(maxRecursionDepth).
			A(cameraAperture, nil).
			F(cameraFocusVector.Length()).
			V(800 * zoom * cameraZoom)

		scn := scene.NewSceneNode().S(spaceEnvironment).S(sun, moon).SN(earth)

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

func createEarth(radius float64, position *vec3.T) *scene.SceneNode {
	textureEarthDay, err := floatimage.EmptyPlaceholderImage("textures/planets/earth/earth_daymap.jpg")
	if err != nil {
		panic(err)
	}

	textureEarthCloud, err := floatimage.EmptyPlaceholderImage("textures/planets/earth/earth_clouds_alpha.png")
	if err != nil {
		panic(err)
	}

	textureEarthSpec, err := floatimage.EmptyPlaceholderImage("textures/planets/earth/earth_daymap_specular_alpha.png")
	if err != nil {
		panic(err)
	}

	textureEarthLights, err := floatimage.EmptyPlaceholderImage("textures/planets/earth/earth_nightmap_alpha.png")
	if err != nil {
		panic(err)
	}

	origin := &vec3.Zero

	earthMantleMaterial := scene.NewMaterial().N("earth").C(color.NewColor(1.0, 1.0, 1.0))
	earthMantleMaterial.SP(textureEarthDay, copy(origin), vec3.T{1, 0, 0}, vec3.T{0, 1, 0})
	earthMantleMaterial.Projection.Interpolation = floatimage.InterpolationBicubic
	earthMantleMaterial.FresnelMaxGlossiness = 0.0
	earthMantle := scene.NewSphere(copy(origin), radius, earthMantleMaterial)
	earthMantle.Translate(position)

	earthCloudMaterial := scene.NewMaterial().N("earth clouds").C(color.NewColor(1.0, 1.0, 1.0))
	earthCloudMaterial.SP(textureEarthCloud, copy(origin), vec3.T{1, 0, 0}, vec3.T{0, 1, 0})
	earthCloudMaterial.Projection.Interpolation = floatimage.InterpolationBicubic
	earthCloudMaterial.FresnelMaxGlossiness = 0.0
	earthCloud := scene.NewSphere(copy(origin), radius+0.01, earthCloudMaterial)
	earthCloud.Translate(position)

	earthSpecMaterial := scene.NewMaterial().N("earth specular").C(color.NewColor(1.0, 1.0, 1.0))
	earthSpecMaterial.SP(textureEarthSpec, copy(origin), vec3.T{1, 0, 0}, vec3.T{0, 1, 0})
	earthSpecMaterial.M(0.05, 0.4)
	earthSpecMaterial.FresnelMaxGlossiness = 0.0
	earthSpec := scene.NewSphere(copy(origin), radius+0.001, earthSpecMaterial)
	earthSpec.Translate(position)

	earthLightsMaterial := scene.NewMaterial().N("earth lights").C(color.NewColor(1.0, 1.0, 1.0))
	earthLightsMaterial.SP(textureEarthLights, copy(origin), vec3.T{1, 0, 0}, vec3.T{0, 1, 0})
	earthLightsMaterial.Projection.Interpolation = floatimage.InterpolationBicubic
	earthLightsMaterial.E(color.White, 0.75, false)
	earthLightsMaterial.FresnelMaxGlossiness = 0.0
	earthLights := scene.NewSphere(copy(origin), radius+0.002, earthLightsMaterial)
	earthLights.Translate(position)

	earth := &scene.SceneNode{Spheres: []*scene.Sphere{earthMantle, earthSpec, earthLights, earthCloud}}
	earth.UpdateBounds()

	return earth
}

func copy(v *vec3.T) *vec3.T {
	v2 := *v
	return &v2
}
