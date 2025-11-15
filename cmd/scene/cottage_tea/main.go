package main

import (
	"fmt"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	anm "pathtracer/internal/pkg/renderfile"
	scn "pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "cottage_tea"

//var imageWidth = 600
//var imageHeight = 800
//var magnification = 1.0

var imageWidth = 800
var imageHeight = 600
var magnification = 1.5

var amountSamples = 1024 * 42

var apertureSize = 0.2

var keroseneLampEmission = 450.0
var skyDomeEmission = 1.5

func main() {
	//lamp := scn.NewSphere(&vec3.T{0, 300, -300}, 100, scn.NewMaterial().N("lamp").E(color.White, 10*0, true))

	skyDome := scn.NewSphere(&vec3.T{0, 0, 0}, 200*100, scn.NewMaterial().
		E(color.White, skyDomeEmission, true).
		//C(color.NewColorGrey(0.2))).
		SP(floatimage.Load("textures/equirectangular/331_PDM_BG1.jpg"), &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("sky dome")
	// SP("textures/equirectangular/white room 02 612x612.jpg", &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("sky dome")
	skyDome.RotateY(&vec3.Zero, util.DegToRad(-20))

	tableBoard := obj.NewBox(obj.BoxCentered)
	tableBoard.Scale(&vec3.Zero, &vec3.T{110 / 2, 3 / 2, 150 / 2})
	tableBoard.Translate(&vec3.T{0, -tableBoard.Bounds.Ymax + 80, 0})
	tableBoard.Material = scn.NewMaterial().
		C(color.NewColorGrey(1.0)).
		PP(floatimage.Load("textures/wallpaper/Blossom2_Image_Tile_Item_9471w.jpg"), &vec3.T{0, 0, 0}, vec3.T{60, 0, 0}, vec3.T{0, 0, 40})

	wallZ := tableBoard.Bounds.Zmax + 2.0

	window := obj.NewWindow(150)
	window.Scale(&vec3.Zero, &vec3.T{1.3, 1, 1})
	window.Translate(&vec3.T{0, tableBoard.Bounds.Ymax + 5, wallZ + window.Bounds.SizeZ()/6})

	var wallWidth = 300.0
	var wallHeight = 240.0
	var windowYPos = tableBoard.Bounds.Ymax + 5
	var windowXPos = wallWidth / 2

	var windowWidth = window.Bounds.SizeX()
	var windowHeight = window.Bounds.SizeY()

	wall := createWindowWall(wallWidth, wallHeight, windowXPos, windowYPos, windowWidth, windowHeight)
	wall.Translate(&vec3.T{-wallWidth / 2, 0, wallZ})
	wall.Material = scn.NewMaterial().PP(floatimage.Load("textures/wallpaper/Slottsteatern_Image_Flatshot_Item_4507.jpg"), &vec3.T{0, 0, 0}, vec3.T{52, 0, 0}, vec3.T{0, 52, 0})
	//wall.Material = scn.NewMaterial().PP("textures/wallpaper/Ester_Image_Flatshot_Item_7659.jpg", &vec3.T{0, 0, 0}, vec3.T{100, 0, 0}, vec3.T{0, 100, 0})
	//wall.Material = scn.NewMaterial().PP("textures/wallpaper/RoseGarden_Image_Tile_Item_7464.jpg", &vec3.T{0, 0, 0}, vec3.T{100, 0, 0}, vec3.T{0, 50, 0})

	keroseneLamp := obj.NewKeroseneLamp(40, keroseneLampEmission)
	keroseneLamp.RotateY(&vec3.Zero, util.DegToRad(-90))
	keroseneLamp.Translate(&vec3.T{20, tableBoard.Bounds.Ymax, -20})

	porcelainMaterial := scn.NewMaterial().
		N("Porcelain").
		C(color.NewColorGrey(0.85)).
		M(0.1, 0.1).
		T(0.0, true, scn.RefractionIndex_Porcelain)
	var steelColorFactor = 0.8
	steelCutleryMaterial := scn.NewMaterial().N("steel").
		C(color.NewColor(0.93*steelColorFactor, 0.93*steelColorFactor, 0.95*steelColorFactor)).
		M(0.9, 0.2).
		T(0.0, true, scn.RefractionIndex_Glass)

	teapot := obj.NewSolidUtahTeapot(21, true, true)
	teapot.ReplaceMaterial("teapot", porcelainMaterial)
	teapot.ReplaceMaterial("lid", porcelainMaterial)
	teapot.RotateY(&vec3.Zero, util.DegToRad(-65))
	teapot.Translate(&vec3.T{0, tableBoard.Bounds.Ymax, 20})

	teacup := obj.NewTeacup(12, true, true, false)
	teacup.ReplaceMaterial("teacup", porcelainMaterial)
	teacup.ReplaceMaterial("saucer", porcelainMaterial)
	teacup.RotateY(&vec3.Zero, util.DegToRad(22.5))
	teacup.Translate(&vec3.T{-25 - 5, tableBoard.Bounds.Ymax, 5})

	spoon := obj.NewTeacup(8, false, false, true)
	spoon.ReplaceMaterial("spoon", steelCutleryMaterial)
	spoon.RotateY(&vec3.Zero, util.DegToRad(-80))
	spoon.Translate(&vec3.T{-25 - 5, tableBoard.Bounds.Ymax, -12})

	room := wavefrontobj.ReadOrPanic(filepath.Join(obj.ObjEvaluationFileDir, "skydome_open.obj"))
	room.ScaleUniform(&vec3.Zero, 1/room.Bounds.SizeY())
	room.ScaleUniform(&vec3.Zero, 5.5*100)
	room.RotateY(&vec3.Zero, util.DegToRad(-90))
	room.Translate(&vec3.T{0, wallHeight / 2, -100})
	room.Material = scn.NewMaterial().E(color.NewColorKelvin(2000), 0.2, true).SP(floatimage.Load("textures/equirectangular/medieval_kitchen.png"), room.Bounds.Center(), vec3.UnitX.Scaled(-1), vec3.UnitY)

	scene := scn.NewSceneNode().
		S(skyDome).
		FS(room).
		FS(wall).
		FS(window).
		FS(teapot).
		FS(teacup, spoon).
		FS(tableBoard).
		FS(keroseneLamp)

	//cameraOrigin := &vec3.T{-20, 110, -50}
	//focusPoint := &vec3.T{20, tableBoard.Bounds.Ymax + 20, -20}

	cameraOrigin := &vec3.T{-40, 130, -100}
	focusPoint := &vec3.T{-10, tableBoard.Bounds.Ymax + 20, 0}

	viewVector := focusPoint.Subed(cameraOrigin)
	focusDistance := viewVector.Length()

	camera := scn.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).
		A(apertureSize, nil).
		F(focusDistance).
		D(10)

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, true)
	frame := scn.NewFrame(animation.AnimationName, -1, camera, scene)
	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := anm.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func createWindowWall(wallWidth float64, wallHeight float64, windowXPos float64, windowYPos float64, windowWidth float64, windowHeight float64) *scn.FacetStructure {
	wall := &scn.FacetStructure{SubstructureName: "wall"}

	tmpFacets := &scn.FacetStructure{SubstructureName: "tmp"}
	var offset = 0.0

	tmpFacets.Facets = obj.GetRectangleFacets(&vec3.T{0, 1, 0}, &vec3.T{0, 0, 0}, &vec3.T{1, 0, 0}, &vec3.T{1, 1, 0})
	tmpFacets.Scale(&vec3.Zero, &vec3.T{windowXPos - windowWidth/2, wallHeight, 1})
	tmpFacets.Translate(&vec3.T{0 - offset, 0, 0})
	wall.Facets = append(wall.Facets, tmpFacets.Facets...)

	tmpFacets.Facets = obj.GetRectangleFacets(&vec3.T{0, 1, 0}, &vec3.T{0, 0, 0}, &vec3.T{1, 0, 0}, &vec3.T{1, 1, 0})
	tmpFacets.Scale(&vec3.Zero, &vec3.T{wallWidth - (windowXPos + windowWidth/2), wallHeight, 1})
	tmpFacets.Translate(&vec3.T{windowXPos + windowWidth/2 + offset, 0, 0})
	wall.Facets = append(wall.Facets, tmpFacets.Facets...)

	tmpFacets.Facets = obj.GetRectangleFacets(&vec3.T{0, 1, 0}, &vec3.T{0, 0, 0}, &vec3.T{1, 0, 0}, &vec3.T{1, 1, 0})
	tmpFacets.Scale(&vec3.Zero, &vec3.T{windowWidth, wallHeight - (windowHeight + windowYPos), 1})
	tmpFacets.Translate(&vec3.T{windowXPos - windowWidth/2, windowHeight + windowYPos + offset, 0})
	wall.Facets = append(wall.Facets, tmpFacets.Facets...)

	tmpFacets.Facets = obj.GetRectangleFacets(&vec3.T{0, 1, 0}, &vec3.T{0, 0, 0}, &vec3.T{1, 0, 0}, &vec3.T{1, 1, 0})
	tmpFacets.Scale(&vec3.Zero, &vec3.T{windowWidth, windowYPos, 1})
	tmpFacets.Translate(&vec3.T{windowXPos - windowWidth/2, 0 - offset, 0})
	wall.Facets = append(wall.Facets, tmpFacets.Facets...)

	return wall
}
