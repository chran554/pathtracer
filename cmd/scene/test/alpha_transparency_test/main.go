package main

import (
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

const (
	animationName = "alpha_transparency_test"
	imageWidth    = 300
	imageHeight   = 150
	magnification = 4.0
	amountSamples = 1048
	radius        = 5.0
)

func main() {
	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, true)

	environmentTexture := scene.NewTextureReferenceOrPanic("textures/equirectangular/6031885477_bbffaa37d7_o.jpg", scene.TextureTypeAlbedo)
	// environmentTexture := scene.NewTextureReferenceOrPanic("textures/equirectangular/las-vegas-hotell-lobby.png", scene.TextureTypeAlbedo)
	environment := scene.NewSphere(&vec3.Zero, 500, scene.NewMaterial().
		E(color.White, 1, true).
		SP(environmentTexture, &vec3.Zero, vec3.UnitZ.Inverted(), vec3.UnitY)).N("sky dome")

	lamp := obj.NewSquareFacetStructure(obj.SquareTypeXZPlane, nil, false)
	lamp.CenterOn(&vec3.Zero)
	lamp.Scale(&vec3.Zero, &vec3.T{100, 1, 40})
	lamp.Translate(&vec3.T{0, 33 * 2.54, -10})
	lamp.Material = scene.NewMaterial().N("lamp").E(color.KelvinTemperatureColor2(4500), 4, true)

	albedo := scene.NewTextureReferenceOrPanic("textures/test/test_alpha_transparency.png", scene.TextureTypeAlbedo)
	mask := scene.NewTextureReferenceOrPanic("textures/test/checkered dark gray.png", scene.TextureTypeTransparency)

	materialFor := func(name string, center *vec3.T, solid bool) *scene.Material {
		// Project each image once across the diameter so the pattern aligns on both objects.
		corner := center.Subed(&vec3.T{radius, radius, 0})
		u, v := vec3.UnitX.Scaled(2*radius), vec3.UnitY.Scaled(2*radius)

		return scene.NewMaterial().N(name).
			C(color.NewColorRGBA(0.9, 0.8, 0.7, 1.0)).
			M(0, 0.0).
			T(0.0, solid, scene.RefractionIndex_Glass).
			PP(albedo, &corner, u, v).
			PP(mask, &corner, u, v)
	}

	left, right := vec3.T{-5.75, 0, 0}, vec3.T{5.75, 0, 0}
	front := vec3.UnitZ.Inverted()

	disc := scene.NewDisc(&left, &front, radius, materialFor("thin glass", &left, false)).N("thin disc")
	sphere := scene.NewSphere(&right, radius, materialFor("solid glass", &right, true)).N("solid sphere")
	scn := scene.NewSceneNode().D(disc).S(sphere, environment).FS(lamp)

	camera := scene.NewCamera(&vec3.T{0, 0, -70}, &vec3.Zero, amountSamples, magnification).V(800).D(6)
	animation.AddFrame(scene.NewFrame(animationName, -1, camera, scn))
	if err := renderfile.WriteRenderFile("scene/"+animationName+".render.zip", animation); err != nil {
		panic(err)
	}
}
