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

var animationName = "refraction_test"

var ballRadius float64 = 30

var maxRecursionDepth = 10
var amountSamples = 1024 * 4 * 4
var lensRadius = 0.0 // 2.0        // 0.25

var imageWidth = 600
var imageHeight = 400
var magnification = 2.0

func main() {
	cornellBox := GetCornellBox(&vec3.T{250, 250, 250}, 6.0) // cm, as units. I.e. a 5x3x5m room

	groundOffset := 0.2
	sphere1 := GetSphere(&vec3.T{-70, ballRadius + groundOffset, 0}, ballRadius, "sphere 1")
	sphere2 := GetSphere(&vec3.T{0, ballRadius + groundOffset, 0}, ballRadius, "sphere 2")
	sphere3 := GetSphere(&vec3.T{70, ballRadius + groundOffset, 0}, ballRadius, "sphere 3")

	sphere2.Material.C(color.NewColor(0.95, 0.85, 0.85))

	sphere3.Material.C(color.NewColor(0.85, 0.85, 0.95))

	//glassPokal := obj.NewGlassIkeaPokal(50.0)
	//glassPokal.Translate(&vec3.T{10, 0, -20})

	//glassSkoja := obj.NewGlassIkeaSkoja(40.0, true)
	//glassSkoja.Translate(&vec3.T{35, 0, 0})

	//utahTeapot := obj.NewSolidUtahTeapot(50.0, true, true)
	//utahTeapot.RotateY(&vec3.T{0, 0, 0}, -math.Pi/3.5-math.Pi/2.0)
	//utahTeapot.Translate(&vec3.T{25 + 5, 0, 150})

	pixarBallRadius := 40.0
	pixarBallOrigin := &vec3.T{-40, pixarBallRadius + groundOffset, 80}
	pixarBall := obj.NewPixarBall(pixarBallOrigin, pixarBallRadius)
	pixarBall.RotateY(pixarBallOrigin, util.DegToRad(40))
	pixarBall.RotateX(pixarBallOrigin, util.DegToRad(-10))

	scn := scene.NewSceneNode().
		S(sphere1, sphere2, sphere3).
		S(pixarBall).
		FS(cornellBox /*glassPokal, glassSkoja, utahTeapot*/)

	sphereBounds := sphere2.Bounds()
	cameraOrigin := &vec3.T{sphereBounds.Center()[0], cornellBox.Bounds.Ymax / 3, sphereBounds.Center()[2] - 150*2}
	cameraFocusPoint := sphereBounds.Center().Added(&vec3.T{0, ballRadius, -(ballRadius * 2 / 3)})
	camera := scene.NewCamera(cameraOrigin, &cameraFocusPoint, amountSamples, magnification).D(maxRecursionDepth).A(lensRadius, nil)

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	frame := scene.NewFrame(animation.AnimationName, -1, camera, scn)

	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func GetSphere(origin *vec3.T, radius float64, name string) *scene.Sphere {
	sphereMaterial := scene.NewMaterial().
		N(name).
		C(color.NewColor(0.90, 0.92, 0.95)).
		M(0.270, 0.030).
		T(0.700, true, scene.RefractionIndex_Glass)

	sphere := scene.NewSphere(origin, radius, sphereMaterial).N(name)

	return sphere
}

func GetCornellBox(scale *vec3.T, lightIntensityFactor float64) *scene.FacetStructure {
	cornellBox := obj.NewCornellBox(scale, true, lightIntensityFactor)
	cornellBox.ClearMaterials()

	cornellBox.Material = scene.NewMaterial().
		N("Cornell box material").
		C(color.NewColor(0.95, 0.95, 0.95))

	textureWalls := floatimage.LoadOrPanic("textures/wallpaper/anemone-rose-flower-eucalyptus-leaves-pampas-grass.png")
	textureFloor := floatimage.LoadOrPanic("textures/floor/7451-diffuse 02.png")

	backWallMaterial := cornellBox.Material.Copy().PP(textureWalls, &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(scale[0]), vec3.UnitY.Scaled(scale[0]*0.66))
	leftWallMaterial := cornellBox.Material.Copy().C(color.NewColor(0.75, 0.75, 1.0)).PP(textureWalls, &vec3.T{0, 0, 0}, vec3.UnitZ.Scaled(scale[0]), vec3.UnitY.Scaled(scale[0]*0.66))
	rightWallMaterial := cornellBox.Material.Copy().C(color.NewColor(1.0, 0.75, 0.75)).PP(textureWalls, &vec3.T{0, 0, 0}, vec3.UnitZ.Scaled(-scale[0]), vec3.UnitY.Scaled(scale[0]*0.66))
	floorMaterial := cornellBox.Material.Copy().PP(textureFloor, &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(scale[0]*0.25), vec3.UnitZ.Scaled(scale[0]*0.25))

	lampMaterial := scene.NewMaterial().N("Lamp").
		C(color.White).
		E(color.White, lightIntensityFactor, true)

	/*
		cornellBox.GetFirstObjectBySubstructureName("Lamp_1_-_left_away").Material = lampMaterial
		cornellBox.GetFirstObjectBySubstructureName("Lamp_2_-_left_close").Material = lampMaterial
		cornellBox.GetFirstObjectBySubstructureName("Lamp_3_-_right_away").Material = lampMaterial
		cornellBox.GetFirstObjectBySubstructureName("Lamp_4_-_right_close").Material = lampMaterial
	*/

	cornellBox.GetFirstObjectByName("Lamp").Material = lampMaterial

	cornellBox.GetFirstObjectBySubstructureName("Back").Material = backWallMaterial
	cornellBox.GetFirstObjectBySubstructureName("Left").Material = leftWallMaterial
	cornellBox.GetFirstObjectBySubstructureName("Right").Material = rightWallMaterial
	cornellBox.GetFirstObjectBySubstructureName("Floor").Material = floorMaterial

	return cornellBox
}
