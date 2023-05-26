package main

import (
	"math"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"
	"strconv"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "refraction_test"

var ballRadius float64 = 30

var maxRecursionDepth = 10
var amountSamples = 1024 * 8 // * 2 // * 5 // * 4
var lensRadius = 0.0         // 2.0        // 0.25

var imageWidth = 600
var imageHeight = 400
var magnification = 2.0

func main() {
	cornellBox := GetCornellBox(&vec3.T{300, 300, 300}, 10.0) // cm, as units. I.e. a 5x3x5m room

	spheres := GetSpheres(1, &vec3.T{0, 0, 0})
	spheres2 := GetSpheres(1, &vec3.T{0, 0, 0})
	spheres2[0].Origin.Add(&vec3.T{40, 0, -20})
	spheres2[0].Material.C(color.NewColor(0.95, 0.6, 0.5))

	spheres = append(spheres, spheres2...)

	glassPokal := obj.NewGlassIkeaPokal(50.0)
	glassPokal.Translate(&vec3.T{10, 0, -20})

	glassSkoja := obj.NewGlassIkeaSkoja(40.0, true)
	glassSkoja.Translate(&vec3.T{35, 0, 0})

	// glassMaterial := scn.NewMaterial().
	// 	N("glass material").
	// 	C(color.NewColor(0.98, 0.80, 0.75)).
	// 	M(0.2, 0.05).
	// 	T(0.95, true, scn.RefractionIndex_Glass)
	utahTeapot := obj.NewSolidUtahTeapot(50.0, true, true)
	utahTeapot.RotateY(&vec3.T{0, 0, 0}, -math.Pi/3.5-math.Pi/2.0)
	utahTeapot.Translate(&vec3.T{25 + 5, 0, 150})
	// utahTeapot.Material = glassMaterial

	pixarBallRadius := 20.0
	pixarBallOrigin := &vec3.T{-130 + 30, pixarBallRadius, 160}
	pixarBall := obj.NewPixarBall(pixarBallOrigin, pixarBallRadius)
	//pixarBall.RotateY(pixarBallOrigin, 0.0)
	pixarBall.RotateY(pixarBallOrigin, math.Pi/4.0)

	scene := scn.NewSceneNode().
		S(spheres...).
		S(pixarBall).
		FS(cornellBox /*glassPokal, glassSkoja,*/, utahTeapot)

	sphereBounds := spheres[0].Bounds()
	cameraOrigin := &vec3.T{sphereBounds.Center()[1], sphereBounds.Center()[1], -200}
	cameraFocusPoint := sphereBounds.Center().Added(&vec3.T{0, 0, -ballRadius / 3.0})
	camera := scn.NewCamera(cameraOrigin, &cameraFocusPoint, amountSamples, magnification).D(maxRecursionDepth).A(lensRadius, "")

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	frame := scn.NewFrame(animation.AnimationName, -1, camera, scene)

	animation.AddFrame(frame)

	anm.WriteAnimationToFile(animation, false)
}

func GetSpheres(amountSpheres int, translation *vec3.T) []*scn.Sphere {
	var spheres []*scn.Sphere
	sphereMaterial := scn.NewMaterial().
		N("glass sphere").
		C(color.NewColor(0.95, 0.95, 0.99)).
		M(0.01, 0.05).
		T(0.98, true, scn.RefractionIndex_Glass)

	for i := 0; i < amountSpheres; i++ {
		positionOffsetX := 0.0 // (-sphereSpread/2.0 + float64(i)*sphereCC) * 0.5
		positionOffsetZ := 0.0 // (-sphereSpread/2.0 + float64(i)*sphereCC) * 1.0

		groundOffset := 0.02 // ballRadius / 2.0 // Raise the sphere above the ground
		sphereOrigin := vec3.T{positionOffsetX, ballRadius + groundOffset, positionOffsetZ}
		sphere := scn.NewSphere(&sphereOrigin, ballRadius, sphereMaterial).N("Glass sphere #" + strconv.Itoa(i))

		sphere.Translate(translation)

		spheres = append(spheres, sphere)
	}

	return spheres
}

func GetCornellBox(scale *vec3.T, lightIntensityFactor float64) *scn.FacetStructure {
	var cornellBoxFilename = "cornellbox.obj"
	var cornellBoxFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + cornellBoxFilename

	cornellBox := wavefrontobj.ReadOrPanic(cornellBoxFilenamePath)

	cornellBox.Scale(&vec3.Zero, scale)
	cornellBox.ClearMaterials()

	cornellBox.Material = scn.NewMaterial().
		N("Cornell box material").
		C(color.NewColor(0.95, 0.95, 0.95))

	backWallMaterial := cornellBox.Material.Copy().
		PP("textures/wallpaper/anemone-rose-flower-eucalyptus-leaves-pampas-grass.png", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(scale[0]), vec3.UnitY.Scaled(scale[0]*0.66))
	sideWallMaterial := cornellBox.Material.Copy().
		PP("textures/wallpaper/anemone-rose-flower-eucalyptus-leaves-pampas-grass.png", &vec3.T{0, 0, 0}, vec3.UnitZ.Scaled(scale[0]), vec3.UnitY.Scaled(scale[0]*0.66))
	floorMaterial := cornellBox.Material.Copy().
		PP("textures/floor/7451-diffuse 02 low contrast.png", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(scale[0]*0.25), vec3.UnitZ.Scaled(scale[0]*0.25))

	lampMaterial := scn.NewMaterial().N("Lamp").
		C(color.White).
		E(color.White, lightIntensityFactor, true)

	cornellBox.GetFirstObjectBySubstructureName("Lamp_1_-_left_away").Material = lampMaterial
	cornellBox.GetFirstObjectBySubstructureName("Lamp_2_-_left_close").Material = lampMaterial
	cornellBox.GetFirstObjectBySubstructureName("Lamp_3_-_right_away").Material = lampMaterial
	cornellBox.GetFirstObjectBySubstructureName("Lamp_4_-_right_close").Material = lampMaterial

	cornellBox.GetFirstObjectBySubstructureName("Back").Material = backWallMaterial
	cornellBox.GetFirstObjectBySubstructureName("Left").Material = sideWallMaterial
	cornellBox.GetFirstObjectBySubstructureName("Right").Material = sideWallMaterial
	cornellBox.GetFirstObjectBySubstructureName("Floor").Material = floorMaterial

	return cornellBox
}
