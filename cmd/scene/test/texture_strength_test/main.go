package main

import (
	"fmt"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

const (
	animationName = "texture_strength_test"

	amountAnimationFrames = 1

	imageWidth    = 150
	imageHeight   = 150
	magnification = 4.0 // 1.0

	amountSamples = 2048 * 16 // 1024 * 24 / 4

	apertureSize = 0.0

	lampEmission = 4
	radius       = 5.0
)

func main() {
	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, true)

	origin := &vec3.T{0, 0, 0}

	// textureEnvironment := scene.NewTextureReferenceOrPanic("textures/equirectangular/las-vegas-hotell-lobby.png", scene.TextureTypeAlbedo) // Pool hall
	textureEnvironment := scene.NewTextureReferenceOrPanic("textures/equirectangular/6031885477_bbffaa37d7_o.jpg", scene.TextureTypeAlbedo) // Arcade
	/// textureEnvironment := scene.NewTextureReferenceOrPanic("textures/equirectangular/52750393381_973073d1a5_o.jpg", scene.TextureTypeAlbedo) // Forest in autumn
	environmentSphere := scene.NewSphere(origin, 5*100, scene.NewMaterial().
		E(color.White, 1, true).
		SP(textureEnvironment, origin, vec3.UnitZ.Inverted(), vec3.UnitY)).N("sky dome")

	lamp := obj.NewSquareFacetStructure(obj.SquareTypeXZPlane, nil, false)
	lamp.CenterOn(&vec3.Zero)
	lamp.Scale(&vec3.Zero, &vec3.T{100, 1, 40})
	lamp.Translate(&vec3.T{0, 33 * 2.54, -10}) // Raise the lamp 33 inches above the table cloth (and a little bit in front of the balls)
	lamp.Material = scene.NewMaterial().N("lamp").
		E(color.KelvinTemperatureColor2(4500), lampEmission, true)

	diffuseTexture := scene.NewTextureReferenceOrPanic("textures/normal maps/Metal053B_4K-PNG/Metal053B_4K-PNG_Color.png", scene.TextureTypeAlbedo)
	normalMapTexture := scene.NewTextureReferenceOrPanic("textures/normal maps/Metal053B_4K-PNG/Metal053B_4K-PNG_NormalGL.png", scene.TextureTypeNormalMap_OpenGL)
	roughnessMapTexture := scene.NewTextureReferenceOrPanic("textures/normal maps/Metal053B_4K-PNG/Metal053B_4K-PNG_Roughness.png", scene.TextureTypeRoughness)
	metalnessMapTexture := scene.NewTextureReferenceOrPanic("textures/normal maps/Metal053B_4K-PNG/Metal053B_4K-PNG_Metalness.png", scene.TextureTypeMetallicness)

	normalMapTexture.S(1.0)    // Strength of normal texture
	roughnessMapTexture.S(1.0) // Strength of roughness texture
	metalnessMapTexture.S(1.0) // Strength of metalness texture

	sphereOrigin := origin

	sphereMaterial := scene.NewMaterial().
		N("sphere").
		M(1.0, 1.0).
		T(0.0, true, scene.RefractionIndex_Aluminum).
		SP(diffuseTexture, sphereOrigin, vec3.UnitZ, vec3.UnitY).
		SP(normalMapTexture, sphereOrigin, vec3.UnitZ, vec3.UnitY).
		SP(roughnessMapTexture, sphereOrigin, vec3.UnitZ, vec3.UnitY).
		SP(metalnessMapTexture, sphereOrigin, vec3.UnitZ, vec3.UnitY)

	sphere := scene.NewSphere(sphereOrigin, radius, sphereMaterial)
	sphere.RotateX(sphere.Bounds().Center(), util.DegToRad(10)) // Tilt the sphere "side" slightly downwards (CCW)
	//sphere.RotateY(sphere.Bounds().Center(), util.DegToRad(-10)) // Rotate the sphere "side" slightly to the left (CW)

	snookerBallsNode := scene.NewSceneNode().S(sphere)

	scn := scene.NewSceneNode().
		S(environmentSphere).
		SN(snookerBallsNode).
		FS(lamp)

	//animationStep := 1.0 / float64(amountAnimationFrames)
	for animationFrameIndex := range amountAnimationFrames {
		// animationProgress := float64(animationFrameIndex) * animationStep

		cameraOrigin := sphereOrigin.Added(&vec3.T{0, 0, -80})
		focusPoint := sphereOrigin
		camera := scene.NewCamera(&cameraOrigin, focusPoint, amountSamples, magnification).
			V(1200).
			A(apertureSize, nil).
			D(6)

		fi := -1
		if fi > 1 {
			fi = animationFrameIndex
		}

		frame := scene.NewFrame(animation.AnimationName, fi, camera, scn)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
