package main

import (
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "window_test"

var amountSamples = 1024 * 8
var apertureSize = 1.0

var cameraDistanceFactor = 1.0

var imageWidth = 500
var imageHeight = 500
var magnification = 2.0

func main() {
	cornellBox := obj.NewWhiteCornellBox(&vec3.T{300, 300, 300}, true, 20) // cm, as units. I.e. a 5x3x5m room
	setCornellBoxMaterial(cornellBox)

	pillarHeight := 5.0
	pillarWidth := 50.0

	window := obj.NewWindow(pillarWidth * 1.5)
	pillar1 := putOnPillar(window, 0, 0, 0, pillarWidth*1.2, pillarHeight)

	focusObject := window

	cameraOrigin := focusObject.Bounds.Center().Add(&vec3.T{0, 45, -150})
	cameraOrigin.Scale(cameraDistanceFactor)
	focusPoint := focusObject.Bounds.Center().Add(&vec3.T{0, -5, 0})
	camera := scn.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).A(apertureSize, "")

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)
	scene := scn.NewSceneNode().
		FS(cornellBox).
		FS(pillar1, window)

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

	backWallMaterial := *cornellBox.Material
	backWallMaterial.PP("textures/wallpaper/VintagePalms_Image_Tile_Item_9454w.jpg", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(1.66*scale), vec3.UnitY.Scaled(scale))
	//backWallMaterial.PP("textures/wallpaper/DarkFloral_Image_Tile_Item_9419w.jpg", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(2*scale), vec3.UnitY.Scaled(2*scale*0.66))
	cornellBox.GetFirstObjectBySubstructureName("Back::Back").Material = &backWallMaterial

	sideWallMaterial := *cornellBox.Material
	sideWallMaterial.PP("textures/wallpaper/VintagePalms_Image_Tile_Item_9454w.jpg", &vec3.T{0, 0, 0}, vec3.UnitZ.Scaled(1.66*scale), vec3.UnitY.Scaled(scale))
	//sideWallMaterial.PP("textures/wallpaper/DarkFloral_Image_Tile_Item_9419w.jpg", &vec3.T{0, 0, 0}, vec3.UnitZ.Scaled(2*scale), vec3.UnitY.Scaled(2*scale*0.66))
	cornellBox.GetFirstObjectBySubstructureName("Left::Left").Material = &sideWallMaterial
	cornellBox.GetFirstObjectBySubstructureName("Right::Right").Material = &sideWallMaterial

	floorMaterial := *cornellBox.Material
	floorMaterial.M(0.3, 0.1).PP("textures/marble/marble white tiles 1000x1000.jpg", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(scale/4), vec3.UnitZ.Scaled(scale/4))
	cornellBox.GetFirstObjectBySubstructureName("Floor::Floor").Material = &floorMaterial
}
