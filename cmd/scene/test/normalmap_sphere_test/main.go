package main

import (
	"fmt"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "normal_map_sphere_test"

var amountAnimationFrames = 1 // 72 * 2 // TODO set to 1

var imageWidth = 640
var imageHeight = 480
var magnification = 2.0

var amountSamples = 512 * 8 * 2 // * 2 * 15 // * 8 // * 15
var maxRecursionDepth = 6

var cameraAperture = 0.0
var cameraZoom = 1.0

var lampIntensity = 8.0
var lightIntensityFactor = 0.15
var sphereRadius = 1.0

func main() {
	//textureGround, err := floatimage.EmptyPlaceholderImage("textures/floor/7451-diffuse 02 low contrast.png")
	textureGround, err := floatimage.EmptyPlaceholderImage("textures/floor/7451-diffuse 02.png")
	if err != nil {
		panic(err)
	}

	textureWall, err := floatimage.EmptyPlaceholderImage("textures/tapeter/Camille_Image_Flatshot_Item_7209.jpg")
	if err != nil {
		panic(err)
	}

	// Cornell box
	cornellBox := obj.NewWhiteCornellBox(&vec3.T{16, 5, 16}, true, lightIntensityFactor) // cm, as units. I.e. a 5x3x5m room
	cornellBox.RemoveObjectsByName("Lamp")

	floorMaterial := cornellBox.GetFirstMaterialByName("Floor")
	floorMaterial.M(0.01, 0.3).PP(textureGround, &vec3.Zero, vec3.T{4, 0, 0}, vec3.T{0, 0, 4})
	floorMaterial.FresnelMaxGlossiness = 0.15

	wallpaperZoomFactor := 1.0
	wallpaperSize := cornellBox.Bounds.SizeY() / wallpaperZoomFactor

	leftWallMaterial := cornellBox.GetFirstMaterialByName("Left")
	leftWallMaterial.PP(textureWall, &vec3.T{0, cornellBox.Bounds.Ymin, cornellBox.Bounds.Zmin}, vec3.T{0, wallpaperSize, 0}, vec3.T{0, 0, wallpaperSize})
	leftWallMaterial.FresnelMaxGlossiness = 0.15

	rightWallMaterial := cornellBox.GetFirstMaterialByName("Right")
	rightWallMaterial.PP(textureWall, &vec3.T{0, cornellBox.Bounds.Ymin, cornellBox.Bounds.Zmin}, vec3.T{0, wallpaperSize, 0}, vec3.T{0, 0, wallpaperSize})
	rightWallMaterial.FresnelMaxGlossiness = 0.15

	backWallMaterial := cornellBox.GetFirstMaterialByName("Back")
	backWallMaterial.PP(textureWall, &vec3.T{cornellBox.Bounds.Xmin, cornellBox.Bounds.Ymin, 0}, vec3.T{0, wallpaperSize, 0}, vec3.T{wallpaperSize, 0, 0})
	backWallMaterial.FresnelMaxGlossiness = 0.15

	ceilingMaterial := cornellBox.GetFirstMaterialByName("Ceiling").C(color.NewColor(1.0, 1.0, 1.0).Multiply(0.90))
	ceilingMaterial.FresnelMaxGlossiness = 0.15

	//textureSphere, err := floatimage.EmptyPlaceholderImage("textures/equirectangular/earth/light.jpg")
	textureSphere, err := floatimage.EmptyPlaceholderImage("textures/equirectangular/earth/blank 75.png")
	if err != nil {
		panic(err)
	}

	textureNormals, err := floatimage.EmptyPlaceholderImage("textures/normal maps/golfball.jpg")
	if err != nil {
		panic(err)
	}

	//goldColor := color.NewColor(1.00, 0.85, 0.58).Multiply(0.9)
	silverColor := color.NewColor(0.85, 0.85, 0.9).Multiply(1.0)
	sphereMaterial := scene.NewMaterial().
		N("sphere").
		C(silverColor).
		M(0.30, 0.05)
	sphereMaterial.FresnelMaxGlossiness = 0.20
	sphereMaterial.ColorizeReflection = true

	// TODO set s1 here

	greenSphereMaterial := scene.NewMaterial().C(color.NewColor(0.1, 0.75, 0.15))
	greenSphereMaterial.FresnelMaxGlossiness = 0.30
	sg := scene.NewSphere(&vec3.T{0, 0, 0}, sphereRadius*0.5, greenSphereMaterial)
	sg.Translate(&vec3.T{-sphereRadius * 1.2, sphereRadius * 0.5, -sphereRadius * 1.2})

	redSphereMaterial := scene.NewMaterial().C(color.NewColor(0.75, 0.1, 0.15))
	redSphereMaterial.FresnelMaxGlossiness = 0.30
	sr := scene.NewSphere(&vec3.T{0, 0, 0}, sphereRadius*0.5, redSphereMaterial)
	sr.Translate(&vec3.T{sphereRadius * 1.2, sphereRadius * 0.5, -sphereRadius * 1.2})

	lampPosition := vec3.Zero.Added(&vec3.T{-4, 3, -4})
	lamp := scene.NewSphere(&lampPosition, 1.5, scene.NewMaterial().
		N("lamp").
		E(color.KelvinTemperatureColor2(5500), lampIntensity, true))

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	for frameIndex := 0; frameIndex < amountAnimationFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountAnimationFrames)

		s1Material := sphereMaterial.Copy()
		s1Material.SP(textureSphere, &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})
		s1Material.Projection.NormalMap = textureNormals

		s1 := scene.NewSphere(&vec3.T{0, 0, 0}, sphereRadius, s1Material)
		s1.RotateY(s1.Bounds().Center(), util.DegToRad(animationProgress*360)) // TODO remove
		s1.Translate(&vec3.T{0, sphereRadius, 0})
		//s1.Translate(&vec3.T{-sphereRadius * 1.25, sphereRadius, 0})

		// cameraPosition := &vec3.T{0, 3, -6}
		cameraPosition := &vec3.T{0, 3, -6}

		//cameraAim := &vec3.T{0, sphereRadius, 0} // TODO restore
		cameraAim := s1.Bounds().Center() // TODO remove
		focusPosition := cameraAim.Added(&vec3.T{0, 0, -sphereRadius / 2})
		cameraFocusVector := focusPosition.Subed(cameraPosition)

		camera := scene.NewCamera(cameraPosition, cameraAim, amountSamples, magnification).
			D(maxRecursionDepth).
			A(cameraAperture, nil).
			F(cameraFocusVector.Length()).
			V(800 * cameraZoom)

		scn := scene.NewSceneNode().FS(cornellBox).S(lamp).S(s1).S(sr, sg) //.S(s2) // TODO restore

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
