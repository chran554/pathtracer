package main

import (
	"fmt"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	anm "pathtracer/internal/pkg/renderfile"
	scn "pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "dragon02_test"

var amountSamples = 1024 * 3 // * 5
var apertureSize = 0.0

var cameraDistanceFactor = 1.0

var imageWidth = 350
var imageHeight = 350
var magnification = 2.0

func main() {
	cornellBox := obj.NewWhiteCornellBox(&vec3.T{300, 300, 300}, true, 20) // cm, as units. I.e. a 5x3x5m room
	setCornellBoxMaterial(cornellBox)

	pillarHeight := 5.0
	pillarWidth := 50.0

	dragon02 := obj.NewDragon02(pillarWidth*0.75, true, false)
	dragon02.UpdateVertexNormalsWithThreshold(false, 10)
	pillar1 := putOnPillar(dragon02, -90, 0, 0, pillarWidth, pillarHeight)

	focusObject := dragon02

	cameraOrigin := focusObject.Bounds.Center().Add(&vec3.T{0, 45, -150})
	cameraOrigin.Scale(cameraDistanceFactor)
	focusPoint := focusObject.Bounds.Center() //.Add(&vec3.T{0, 15, 0})
	camera := scn.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).A(apertureSize, nil)

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)
	scene := scn.NewSceneNode().
		FS(cornellBox).
		FS(pillar1, dragon02)

	frame := scn.NewFrame(animation.AnimationName, -1, camera, scene)
	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := anm.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
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
		PP(floatimage.Load("textures/concrete/Polished-Concrete-Architextures.jpg"), &vec3.T{0, 0, 0}, (&vec3.UnitX).Scaled(pillarWidth), (&vec3.UnitZ).Add(&vec3.T{0, 0.5, 0}).Scaled(pillarWidth))
	pillar1.Translate(&vec3.T{-0.5, 0, -0.5})

	pillar1.Scale(&vec3.Zero, &vec3.T{pillarWidth, pillarHeight, pillarWidth})
	return pillar1
}

func setCornellBoxMaterial(cornellBox *scn.FacetStructure) {
	scale := cornellBox.Bounds.SizeY() / 3

	backWallMaterial := scn.NewMaterial().N("back").PP(floatimage.Load("textures/wallpaper/chinese_dragon.png"), &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(scale), vec3.UnitY.Scaled(scale))
	//backWallMaterial := scn.NewMaterial().N("back").PP("textures/wallpaper/anemone-rose-flower-eucalyptus-leaves-pampas-grass.png", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(1.66*scale), vec3.UnitY.Scaled(scale))
	// backWallMaterial := scn.NewMaterial().N("back").PP("textures/wallpaper/VintagePalms_Image_Tile_Item_9454w.jpg", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(1.66*scale), vec3.UnitY.Scaled(scale))
	cornellBox.GetFirstObjectBySubstructureName("Back").Material = backWallMaterial

	sideWallMaterial := scn.NewMaterial().N("wall").PP(floatimage.Load("textures/wallpaper/chinese_dragon.png"), &vec3.T{0, 0, 0}, vec3.UnitZ.Scaled(scale), vec3.UnitY.Scaled(scale))
	//sideWallMaterial := scn.NewMaterial().N("wall").PP("textures/wallpaper/anemone-rose-flower-eucalyptus-leaves-pampas-grass.png", &vec3.T{0, 0, 0}, vec3.UnitZ.Scaled(1.66*scale), vec3.UnitY.Scaled(scale))
	// sideWallMaterial := scn.NewMaterial().N("wall").PP("textures/wallpaper/VintagePalms_Image_Tile_Item_9454w.jpg", &vec3.T{0, 0, 0}, vec3.UnitZ.Scaled(1.66*scale), vec3.UnitY.Scaled(scale))
	cornellBox.GetFirstObjectBySubstructureName("Left").Material = sideWallMaterial
	cornellBox.GetFirstObjectBySubstructureName("Right").Material = sideWallMaterial

	//floorMaterial := scn.NewMaterial().N("floor").M(0.3, 0.1).PP("textures/marble/marble white tiles 1000x1000.jpg", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(scale/4), vec3.UnitZ.Scaled(scale/4))
	floorMaterial := scn.NewMaterial().N("floor").
		C(color.NewColorGrey(0.5)).M(0.3, 0.1).
		PP(floatimage.Load("textures/floor/floor_boards.png"), &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(scale), vec3.UnitZ.Scaled(scale))
	cornellBox.GetFirstObjectBySubstructureName("Floor").Material = floorMaterial

	cornellBox.GetFirstMaterialByName("Lamp").C(color.NewColorKelvin(4000))
}
