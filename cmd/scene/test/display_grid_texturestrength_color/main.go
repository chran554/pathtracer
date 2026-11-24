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
	animationName = "display_grid_texturestrength_color"

	ballRadius float64 = 20

	maxRecursionDepth         = 6
	amountSamples             = 1024 //* 12
	lensRadius        float64 = 0

	viewPlaneDistance    = 4000.0
	cameraDistanceFactor = 2.0

	lampEmissionFactor = 5.0

	imageWidth    = 1800
	imageHeight   = 1200
	magnification = 0.75
)

func main() {
	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, true)

	cornellBox := obj.NewCornellBox(&vec3.T{700, 500, 700}, false, lampEmissionFactor)
	wallColor := color.NewColorGrey(0.8)
	cornellBox.ReplaceMaterial("Left", scene.NewMaterial().N("Left").C(wallColor))
	cornellBox.ReplaceMaterial("Right", scene.NewMaterial().N("Right").C(wallColor))
	cornellBox.ReplaceMaterial("Back", scene.NewMaterial().N("Back").C(wallColor))

	scn := scene.NewSceneNode().FS(cornellBox)

	amountSpheres := 6
	sphereSpread := ballRadius * 2.0 * (float64(amountSpheres) + 1) * 1.3
	sphereCC := sphereSpread / float64(amountSpheres)

	for yIndex := 0; yIndex <= amountSpheres; yIndex++ {
		for xIndex := 0; xIndex <= amountSpheres; xIndex++ {
			yProgress := float64(yIndex+1) / float64(amountSpheres)
			xProgress := float64(xIndex) / float64(amountSpheres)

			strength := xProgress
			colorProgress := yProgress

			sphereOrigin := &vec3.T{-sphereSpread/2.0 + float64(xIndex)*sphereCC, ballRadius, -sphereSpread/2.0 + float64(yIndex)*sphereCC}

			textureOrigin := sphereOrigin.Subed(&vec3.T{-ballRadius, -ballRadius, 0})
			sphereMaterial := scene.NewMaterial().
				C(color.NewColorHSL(360*colorProgress, 1.0, 0.6)).
				M(0.0, 0.0).
				T(0.0, true, scene.RefractionIndex_Glass).
				PP(scene.NewTextureReferenceOrPanic("textures/pixar_ball_02.png", scene.TextureTypeAlbedo).S(strength), &textureOrigin, vec3.UnitX.Scaled(ballRadius*2), vec3.UnitY.Scaled(ballRadius*2))

			if yIndex == 0 {
				sphereMaterial.Color = color.NewColorGrey(0.9)
			}

			sphere := scene.NewSphere(sphereOrigin, ballRadius, sphereMaterial).N(fmt.Sprintf("Sphere (strength:%02f roughness:%02f)", xProgress, yProgress))
			sphere.RotateY(sphereOrigin, util.DegToRad(20))
			sphere.RotateX(sphereOrigin, util.DegToRad(-30))

			scn.S(sphere)
		}
	}

	cameraOrigin := vec3.T{0, 400, -400}
	cameraOrigin.Scale(cameraDistanceFactor)
	focusPoint := vec3.T{0, ballRadius, -ballRadius * 2}
	camera := scene.NewCamera(&cameraOrigin, &focusPoint, amountSamples, magnification).V(viewPlaneDistance).A(lensRadius, nil).D(maxRecursionDepth)

	frame := scene.NewFrame(animation.AnimationName, -1, camera, scn)

	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
