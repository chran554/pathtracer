package main

import (
	"fmt"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

const (
	animationName = "textured_vase_test"

	amountSamples = 1024 * 4 //* 2
	lensRadius    = 2.0

	cameraDistanceFactor = 1.0

	imageWidth    = 200
	imageHeight   = 400
	magnification = 3.0 // 2.0 / 3.0
)

func main() {
	cornellBox := obj.NewWhiteCornellBox(&vec3.T{150, 150, 200}, true, 3.0) // cm, as units. I.e. a 5x3x5m room
	setCornellBoxMaterial(cornellBox)

	pillarHeight := 10.0
	pillarWidth := 45.0

	pillar := obj.NewBox(obj.BoxPositive)
	pillar.Material = scene.NewMaterial().
		C(color.NewColorGrey(0.95)).
		M(0, 0.5).
		PP(
			scene.NewTextureReferenceOrPanic("textures/concrete/Polished-Concrete-Architextures.jpg", scene.TextureTypeAlbedo),
			&vec3.T{0, 0, 0},
			(&vec3.UnitX).Scaled(pillarWidth),
			(&vec3.UnitZ).Add(&vec3.T{0, 0.5, 0}).Scaled(pillarWidth),
		)
	pillar.Translate(&vec3.T{-0.5, 0, -0.5})
	pillar.Scale(&vec3.Zero, &vec3.T{pillarWidth, pillarHeight, pillarWidth})

	vase := obj.NewTexturedVase(40)
	vase.Translate(&vec3.T{0, pillarHeight, 0})

	scn := scene.NewSceneNode().FS(cornellBox, pillar, vase)

	cameraOrigin := vase.Bounds.Center().Add(&vec3.T{-30, 50, -100})
	cameraOrigin.Scale(cameraDistanceFactor)
	focusPoint := vase.Bounds.Center().Add(&vec3.T{0, 0, -vase.Bounds.SizeZ() / 2})
	camera := scene.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).A(lensRadius, nil)

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)
	frame := scene.NewFrame(animation.AnimationName, -1, camera, scn)
	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func setCornellBoxMaterial(cornellBox *scene.FacetStructure) {
	scale := cornellBox.Bounds.SizeX() / 2

	backWallMaterial := cornellBox.GetFirstObjectByMaterialName("Back").Material
	backWallMaterial.PP(
		scene.NewTextureReferenceOrPanic("textures/wallpaper/geometric-yellow.jpg", scene.TextureTypeAlbedo),
		&vec3.T{0, 0, 0},
		vec3.UnitX.Scaled(scale),
		vec3.UnitY.Scaled(scale*0.66),
	)

	leftWallMaterial := cornellBox.GetFirstObjectByMaterialName("Left").Material
	rightWallMaterial := cornellBox.GetFirstObjectByMaterialName("Right").Material
	leftWallMaterial.PP(
		scene.NewTextureReferenceOrPanic("textures/wallpaper/geometric-yellow.jpg", scene.TextureTypeAlbedo),
		&vec3.T{0, 0, 0},
		vec3.UnitZ.Scaled(scale),
		vec3.UnitY.Scaled(scale*0.66),
	)
	rightWallMaterial.PP(
		scene.NewTextureReferenceOrPanic("textures/wallpaper/geometric-yellow.jpg", scene.TextureTypeAlbedo),
		&vec3.T{0, 0, 0},
		vec3.UnitZ.Scaled(scale),
		vec3.UnitY.Scaled(scale*0.66),
	)

	floorMaterial := cornellBox.GetFirstObjectByMaterialName("Floor").Material
	floorMaterial.M(0, 0.2).PP(
		scene.NewTextureReferenceOrPanic("textures/floor/Calacatta-Vena-French-Pattern-Architextures.jpg", scene.TextureTypeAlbedo),
		&vec3.T{0, 0, 0},
		vec3.UnitX.Scaled(scale/2),
		vec3.UnitZ.Scaled(scale/2),
	)
}
