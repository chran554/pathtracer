package main

import (
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "statue_test"

var amountSamples = 1024 * 8
var apertureSize = 1.0

var cameraDistanceFactor = 1.0

var imageWidth = 1024
var imageHeight = 600
var magnification = 1.25

func main() {
	porcelainMaterial := scn.NewMaterial().
		N("Porcelain material").
		C(color.NewColorGrey(0.85)).
		M(0.1, 0.1).
		T(0.0, true, scn.RefractionIndex_Porcelain)

	cornellBox := obj.NewWhiteCornellBox(&vec3.T{300, 300, 300}, true, 20) // cm, as units. I.e. a 5x3x5m room
	setCornellBoxMaterial(cornellBox)

	pillarHeight := 5.0
	pillarWidth := 50.0

	bunny := obj.NewStanfordBunny(pillarWidth * 0.75)
	pillar1 := putOnPillar(bunny, 0, 0, 0, pillarWidth, pillarHeight)

	dragon02 := obj.NewDragon02(pillarWidth*0.75, true, false)
	pillar2 := putOnPillar(dragon02, -90, 1, 0, pillarWidth, pillarHeight)

	teapot := obj.NewSolidUtahTeapot(pillarWidth*0.75, true, true)
	teapot.ReplaceMaterial("utah_teapot_solid::teapot", porcelainMaterial)
	teapot.ReplaceMaterial("utah_teapot_solid::lid", porcelainMaterial)
	pillar3 := putOnPillar(teapot, 22.5, -1, 0, pillarWidth, pillarHeight)

	drWhoAngel := obj.NewDrWhoAngel(pillarWidth*1.5, true, false)
	pillar4 := putOnPillar(drWhoAngel, 0, -1, 1, pillarWidth, pillarHeight)

	focusObject := bunny

	cameraOrigin := focusObject.Bounds.Center().Add(&vec3.T{0, 45, -150})
	cameraOrigin.Scale(cameraDistanceFactor)
	focusPoint := focusObject.Bounds.Center().Add(&vec3.T{0, 15, 0})
	camera := scn.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).A(apertureSize, "")

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)
	scene := scn.NewSceneNode().
		FS(cornellBox).
		FS(pillar1, teapot).
		FS(pillar2, dragon02).
		FS(pillar3, drWhoAngel).
		FS(pillar4, bunny)

	frame := scn.NewFrame(animation.AnimationName, -1, camera, scene)
	animation.AddFrame(frame)

	anm.WriteAnimationToFile(animation, false)
}

func putOnPillar(object *scn.FacetStructure, rotationDegrees float64, xpos int, zpos int, pillarWidth float64, pillarHeight float64) (pillar *scn.FacetStructure) {
	pillar = createPillar(pillarWidth, pillarHeight)
	pillar.Translate(&vec3.T{pillarWidth * 1.3 * float64(xpos), 0, pillarWidth * 1.3 * float64(zpos)})

	object.RotateY(&vec3.Zero, util.DegToRad(rotationDegrees))
	object.Translate(&vec3.T{pillar.Bounds.Center()[0], pillar.Bounds.Ymax, pillar.Bounds.Center()[2]})

	return pillar
}

func createPillar(pillarWidth float64, pillarHeight float64) *scn.FacetStructure {
	pillar1 := obj.NewBox(obj.BoxPositive)
	pillar1.Material = scn.NewMaterial().
		C(color.NewColorGrey(0.9)).
		M(0.3, 0.2).
		PP("textures/concrete/Polished-Concrete-Architextures.jpg", &vec3.T{0, 0, 0}, (&vec3.UnitX).Scaled(pillarWidth), (&vec3.UnitZ).Add(&vec3.T{0, 0.5, 0}).Scaled(pillarWidth))
	pillar1.Translate(&vec3.T{-0.5, 0, -0.5})

	pillar1.Scale(&vec3.Zero, &vec3.T{pillarWidth, pillarHeight, pillarWidth})
	return pillar1
}

func setCornellBoxMaterial(cornellBox *scn.FacetStructure) {
	scale := cornellBox.Bounds.SizeY()

	backWallMaterial := scn.NewMaterial().N("back").PP("textures/wallpaper/anemone-rose-flower-eucalyptus-leaves-pampas-grass.png", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(1.66*scale), vec3.UnitY.Scaled(scale))
	// backWallMaterial := scn.NewMaterial().N("back").PP("textures/wallpaper/VintagePalms_Image_Tile_Item_9454w.jpg", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(1.66*scale), vec3.UnitY.Scaled(scale))
	cornellBox.GetFirstObjectBySubstructureName("Back").Material = backWallMaterial

	sideWallMaterial := scn.NewMaterial().N("wall").PP("textures/wallpaper/anemone-rose-flower-eucalyptus-leaves-pampas-grass.png", &vec3.T{0, 0, 0}, vec3.UnitZ.Scaled(1.66*scale), vec3.UnitY.Scaled(scale))
	// sideWallMaterial := scn.NewMaterial().N("wall").PP("textures/wallpaper/VintagePalms_Image_Tile_Item_9454w.jpg", &vec3.T{0, 0, 0}, vec3.UnitZ.Scaled(1.66*scale), vec3.UnitY.Scaled(scale))
	cornellBox.GetFirstObjectBySubstructureName("Left").Material = sideWallMaterial
	cornellBox.GetFirstObjectBySubstructureName("Right").Material = sideWallMaterial

	floorMaterial := scn.NewMaterial().N("floor").M(0.3, 0.1).PP("textures/marble/marble white tiles 1000x1000.jpg", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(scale/4), vec3.UnitZ.Scaled(scale/4))
	cornellBox.GetFirstObjectBySubstructureName("Floor").Material = floorMaterial
}
