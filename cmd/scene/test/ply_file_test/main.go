package main

import (
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "ply_file_test"

var amountSamples = 1024 * 16
var lensRadius = 2.0

var cameraDistanceFactor = 1.0

var imageWidth = 450
var imageHeight = 450
var magnification = 1.5 // 2.0 / 3.0

func main() {
	cornellBox := obj.NewWhiteCornellBox(&vec3.T{500, 300, 500}, true, 4.0) // cm, as units. I.e. a 5x3x5m room
	setCornellBoxMaterial(cornellBox)

	pillarHeight := 130.0
	pillarWidth := 50.0

	pillar := obj.NewBox(obj.BoxPositive)
	pillar.Material = scn.NewMaterial().
		C(color.NewColorGrey(0.9)).
		M(0.4, 0.1).
		PP("textures/concrete/Polished-Concrete-Architextures.jpg", &vec3.T{0, 0, 0}, (&vec3.UnitX).Scaled(pillarWidth), (&vec3.UnitZ).Add(&vec3.T{0, 0.5, 0}).Scaled(pillarWidth))
	pillar.Translate(&vec3.T{-0.5, 0, -0.5})

	pillar.Scale(&vec3.Zero, &vec3.T{pillarWidth, pillarHeight, pillarWidth})
	pillar.Translate(&vec3.T{0, 0, 100})

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	minDegree := 40 // 0
	maxDegree := 40 // 80
	for degree := minDegree; degree <= maxDegree; degree += 5 {
		beethoven := obj.NewBeethoven(50.0)
		beethoven.Translate(&vec3.T{0, pillar.Bounds.Ymax, pillar.Bounds.Center()[2]})
		// beethoven.Material = scn.NewMaterial().N("Statue").M(0.1, 0.6).PP("textures/marble/white_marble.png", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(200), vec3.UnitY.Scaled(200))
		beethoven.Material = scn.NewMaterial().N("Statue")
		beethoven.UpdateVertexNormalsWithThreshold(false, float64(degree))

		scene := scn.NewSceneNode().FS(cornellBox, pillar, beethoven)

		cameraOrigin := beethoven.Bounds.Center().Add(&vec3.T{0, 0, -100})
		cameraOrigin.Scale(cameraDistanceFactor)
		focusPoint := beethoven.Bounds.Center()
		camera := scn.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).A(lensRadius, "")

		frame := scn.NewFrame(animation.AnimationName, degree, camera, scene)
		animation.AddFrame(frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

func setCornellBoxMaterial(cornellBox *scn.FacetStructure) {
	scale := cornellBox.Bounds.SizeX() / 2

	backWallMaterial := *cornellBox.Material
	backWallMaterial.PP("textures/wallpaper/geometric-yellow.jpg", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(scale), vec3.UnitY.Scaled(scale*0.66))
	cornellBox.GetFirstObjectByName("Back").Material = &backWallMaterial

	sideWallMaterial := *cornellBox.Material
	sideWallMaterial.PP("textures/wallpaper/geometric-yellow.jpg", &vec3.T{0, 0, 0}, vec3.UnitZ.Scaled(scale), vec3.UnitY.Scaled(scale*0.66))
	cornellBox.GetFirstObjectByName("Left").Material = &sideWallMaterial
	cornellBox.GetFirstObjectByName("Right").Material = &sideWallMaterial

	floorMaterial := *cornellBox.Material
	floorMaterial.M(0.6, 0.1).PP("textures/floor/Calacatta-Vena-French-Pattern-Architextures.jpg", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(scale/2), vec3.UnitZ.Scaled(scale/2))
	cornellBox.GetFirstObjectByName("Floor").Material = &floorMaterial
}
