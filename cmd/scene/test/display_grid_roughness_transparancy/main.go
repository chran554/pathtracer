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
	animationName = "display_grid_roughness_transparancy"

	sphereRadius float64 = 20

	maxRecursionDepth         = 6
	amountSamples             = 1024 * 24
	lensRadius        float64 = 0

	viewPlaneDistance    = 4000.0
	cameraDistanceFactor = 2.0

	lampEmissionFactor = 7.0

	imageWidth    = 1800
	imageHeight   = 1200
	magnification = 0.75
)

func main() {
	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, true)

	chessTexture := scene.NewTextureReferenceOrPanic("textures/floor/7451-diffuse 02.png", scene.TextureTypeAlbedo)
	textureScale := 50.0
	floorMaterial := scene.NewMaterial().N("Floor").C(color.NewColorGrey(0.9)).PP(chessTexture, &vec3.T{}, vec3.T{textureScale, 0, 0}, vec3.T{0, 0, textureScale})

	cornellBox := obj.NewCornellBox(&vec3.T{700, 500, 700}, false, lampEmissionFactor)
	cornellBox.ReplaceMaterial("Left", scene.NewMaterial().N("Left").C(color.NewColor(0.85, 0.85, 0.85)))
	cornellBox.ReplaceMaterial("Right", scene.NewMaterial().N("Right").C(color.NewColor(0.85, 0.85, 0.85)))
	cornellBox.ReplaceMaterial("Back", scene.NewMaterial().N("Back").C(color.NewColor(0.70, 0.70, 0.70)))
	cornellBox.ReplaceMaterial("Floor", floorMaterial)

	scn := scene.NewSceneNode().FS(cornellBox)

	amountSpheres := 6
	sphereSpread := sphereRadius * 2.0 * (float64(amountSpheres) + 1) * 1.3
	sphereCC := sphereSpread / float64(amountSpheres)

	for yIndex := 0; yIndex <= amountSpheres; yIndex++ {
		for xIndex := 0; xIndex <= amountSpheres; xIndex++ {
			yProgress := float64(yIndex) / float64(amountSpheres)
			xProgress := float64(xIndex) / float64(amountSpheres)

			roughness := xProgress
			transparency := 1.0 - yProgress

			col := color.NewColor(0.50, 0.06, 0.02)

			sphereMaterial := scene.NewMaterial().
				C(col).
				M(0.0, roughness).
				T(transparency, true, 1.3 /*scene.RefractionIndex_Glass*/).
				Attenuate(col, sphereRadius*2.5)

			sphereOrigin := vec3.T{-sphereSpread/2.0 + float64(xIndex)*sphereCC, sphereRadius, -sphereSpread/2.0 + float64(yIndex)*sphereCC}
			sphere := scene.NewSphere(&sphereOrigin, sphereRadius, sphereMaterial).N(fmt.Sprintf("Sphere (metallic:%02f rough:%02f)", xProgress, yProgress))

			scn.S(sphere)
		}
	}

	cameraOrigin := vec3.T{0, 400, -400}
	cameraOrigin.Scale(cameraDistanceFactor)
	focusPoint := vec3.T{0, sphereRadius, -sphereRadius * 2}
	camera := scene.NewCamera(&cameraOrigin, &focusPoint, amountSamples, magnification).V(viewPlaneDistance).A(lensRadius, nil).D(maxRecursionDepth)

	frame := scene.NewFrame(animation.AnimationName, -1, camera, scn)

	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func interpolate(progress float64, a float64, b float64) float64 {
	return progress*b + (1.0-progress)*a
}
