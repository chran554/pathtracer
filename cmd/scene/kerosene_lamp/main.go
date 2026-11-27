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

const (
	animationName = "kerosene_lamp"

	imageWidth    = 600
	imageHeight   = 800
	magnification = 1.0

	amountSamples = 1024 * 12 // 32 // 12samp x 1.5mag --> 4h
	maxRayDepth   = 15
	apertureSize  = 0.5

	keroseneLampEmission = 20.0
	sootness             = 1.0
	skyDomeEmission      = 0.9
)

func main() {
	// skyTexture := floatimage.LoadOrPanic("textures/equirectangular/331_PDM_BG1.jpg")
	// skyTexture := floatimage.LoadOrPanic("textures/equirectangular/6031885477_bbffaa37d7_o.jpg")
	skyTexture := scene.NewTextureReferenceOrPanic("textures/equirectangular/52764896286_fe987865cf_o.jpg", scene.TextureTypeAlbedo)

	skyMaterial := scene.NewMaterial().N("sky dome").
		E(color.White, skyDomeEmission, true).
		SP(skyTexture, &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})
	skyDome := scene.NewSphere(&vec3.T{0, 0, 0}, 200*100, skyMaterial)
	skyDome.RotateY(&vec3.Zero, util.DegToRad(180))

	tableBoard := obj.NewBox(obj.BoxCentered)
	tableBoard.Scale(&vec3.Zero, &vec3.T{110 / 2, 3 / 2, 150 / 2})
	tableBoard.Translate(&vec3.T{0, -tableBoard.Bounds.Ymax + 80, 0})
	tableBoard.Material = scene.NewMaterial().
		C(color.NewColorGrey(1.0)).
		PP(
			scene.NewTextureReferenceOrPanic("textures/wallpaper/Blossom2_Image_Tile_Item_9471w.jpg", scene.TextureTypeAlbedo),
			&vec3.T{0, 0, 0},
			vec3.T{60, 0, 0},
			vec3.T{0, 0, 40},
		)

	keroseneLamp := obj.NewKeroseneLamp2(40, keroseneLampEmission, sootness)
	//keroseneLamp.RemoveObjectsByName("soot")
	keroseneLamp.RotateY(&vec3.Zero, util.DegToRad(-90))
	keroseneLamp.Translate(&vec3.T{20, tableBoard.Bounds.Ymax, -20})

	scratchesNormalTexture := scene.NewTextureReferenceOrPanic("textures/pbr/ScratchedPaintedMetal01_MR_4K/ScratchedPaintedMetal01_4K_Normal.png", scene.TextureTypeNormalMap_DirectX)
	scratchesNormalTexture.S(0.4).I(floatimage.InterpolationBicubic)

	keroseneBase := keroseneLamp.GetFirstObjectByName("base")
	keroseneBase.Material.SP(scratchesNormalTexture, keroseneBase.Bounds.Center(), vec3.UnitX, vec3.UnitY)

	scn := scene.NewSceneNode().S(skyDome).FS(tableBoard).FS(keroseneLamp)

	cameraOrigin := &vec3.T{-20, 110, -50}
	focusPoint := keroseneLamp.Bounds.Center().Add(&vec3.T{0, 0, 0})

	viewVector := focusPoint.Subed(cameraOrigin)
	focusDistance := viewVector.Length()

	camera := scene.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).
		A(apertureSize, nil).
		F(focusDistance).
		D(maxRayDepth)

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)
	frame := scene.NewFrame(animation.AnimationName, -1, camera, scn)
	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
