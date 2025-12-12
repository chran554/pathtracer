package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"
	"strconv"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "dof_test"

var amountFrames = 4 * 24

var ballRadius float64 = 30

var amountSamples = 1024 * 4 // 128 * 8 * 8 * 3
var maxLensRadius = 48.0     // 12.0

var viewPlaneDistance = 2000.0
var cameraDistanceFactor = 1.0

var imageWidth = 1200
var imageHeight = 600
var magnification = 0.5

func main() {
	wallTexture, err := floatimage.EmptyPlaceholderImage("textures/wallpaper/anemone-rose-flower-eucalyptus-leaves-pampas-grass.png")
	if err != nil {
		panic(err)
	}

	floorTexture, err := floatimage.EmptyPlaceholderImage("textures/floor/tilesf4.jpeg")
	if err != nil {
		panic(err)
	}

	cornellBox := obj.NewCornellBox(&vec3.T{600, 300, 600}, false, 4.0) // cm, as units. I.e., a 5x3x5m room
	cornellBoxUnit := cornellBox.Bounds.SizeY() / 2

	leftWall := cornellBox.GetFirstObjectByMaterialName("Left")
	leftWall.Material = scene.NewMaterial().N("Left").PP(wallTexture, &vec3.T{leftWall.Bounds.Xmin, leftWall.Bounds.Ymin, leftWall.Bounds.Zmin}, vec3.UnitZ.Scaled(cornellBoxUnit), vec3.UnitY.Scaled(cornellBoxUnit))

	rightWall := cornellBox.GetFirstObjectByMaterialName("Right")
	rightWall.Material = scene.NewMaterial().N("Right").PP(wallTexture, &vec3.T{rightWall.Bounds.Xmax, rightWall.Bounds.Ymin, rightWall.Bounds.Zmax}, vec3.UnitZ.Scaled(-cornellBoxUnit), vec3.UnitY.Scaled(cornellBoxUnit))

	backWall := cornellBox.GetFirstObjectByMaterialName("Back")
	backWall.Material = scene.NewMaterial().N("Back").PP(wallTexture, &vec3.T{backWall.Bounds.Xmin, backWall.Bounds.Ymin, backWall.Bounds.Zmax}, vec3.UnitX.Scaled(cornellBoxUnit), vec3.UnitY.Scaled(cornellBoxUnit))

	floor := cornellBox.GetFirstObjectByMaterialName("Floor")
	floor.Material = scene.NewMaterial().N("Floor").PP(floorTexture, &vec3.T{floor.Bounds.Xmin, floor.Bounds.Ymin, floor.Bounds.Zmin}, vec3.UnitX.Scaled(cornellBoxUnit/2), vec3.UnitZ.Scaled(cornellBoxUnit/2))

	amountSpheres := 7
	sphereSpread := ballRadius * 2.0 * (float64(amountSpheres) + 1)
	sphereCC := sphereSpread / float64(amountSpheres-1)

	sphereMaterial := scene.NewMaterial().
		C(color.NewColor(0.95, 0.75, 0.70)).
		M(0.3, 0.0).
		T(0.0, true, scene.RefractionIndex_AcrylicPlastic)

	var spheres []*scene.Sphere
	for i := 0; i < amountSpheres; i++ {
		positionOffsetX := (-sphereSpread/2.0 + float64(i)*sphereCC) * 0.50
		positionOffsetZ := (-sphereSpread/2.0 + float64(i)*sphereCC) * 0.75

		sphereOrigin := &vec3.T{positionOffsetX, ballRadius, positionOffsetZ}
		sphere := scene.NewSphere(sphereOrigin, ballRadius, sphereMaterial).N("Sphere" + strconv.Itoa(i))

		spheres = append(spheres, sphere)
	}

	scn := scene.NewSceneNode().S(spheres...).FS(cornellBox)

	cameraOrigin := vec3.T{0, 200, -800}
	cameraOrigin.Scale(cameraDistanceFactor)
	focusPoint := vec3.T{0, ballRadius, -ballRadius}

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		lensRadius := maxLensRadius * (0.5 + math.Sin(-math.Pi/2+math.Pi*animationProgress)*0.5) // Slow start and slow end

		camera := scene.NewCamera(&cameraOrigin, &focusPoint, amountSamples, magnification).
			V(viewPlaneDistance).
			A(lensRadius, nil).
			D(6)

		fi := -1
		if amountFrames > 1 {
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
