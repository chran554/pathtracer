package main

import (
	"fmt"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "reflective_test_glossy_reflective_color"

var ballRadius float64 = 20

var maxRecursionDepth = 6
var amountSamples = 1024 * 24
var lensRadius float64 = 2

var viewPlaneDistance = 4000.0
var cameraDistanceFactor = 2.0

var lampEmissionFactor = 8.0

var imageWidth = 1800
var imageHeight = 900
var magnification = 1.0 // 0.75

func main() {
	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, true)

	cornellBox := obj.NewCornellBox(&vec3.T{700, 500, 700}, false, lampEmissionFactor)
	cornellBox.ReplaceMaterial("Floor", scene.NewMaterial().N("back").C(color.NewColorGrey(0.50)))
	cornellBox.RemoveObjectsBySubstructureName("Back")
	cornellBox.RemoveObjectsBySubstructureName("Left")
	cornellBox.RemoveObjectsBySubstructureName("Right")
	cornellBox.RemoveObjectsBySubstructureName("Ceiling")

	textureEnvironment, err := floatimage.EmptyPlaceholderImage("textures/equirectangular/6031885477_bbffaa37d7_o.jpg")
	if err != nil {
		panic(err)
	}
	environmentSphere := scene.NewSphere(&vec3.T{0, 0, 0}, 1700, scene.NewMaterial().
		E(color.White, 1, true).
		//C(color.NewColorGrey(0.2))).
		SP(textureEnvironment, &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("sky dome")

	scn := scene.NewSceneNode().FS(cornellBox).S(environmentSphere)

	amountSpheres := 6
	sphereSpread := ballRadius * 2.0 * (float64(amountSpheres) + 1) * 1.3
	sphereCC := sphereSpread / float64(amountSpheres)

	for yIndex := 0; yIndex <= 3; yIndex++ {
		for xIndex := 0; xIndex <= amountSpheres; xIndex++ {
			yProgress := float64(yIndex) / float64(amountSpheres)
			xProgress := float64(xIndex) / float64(amountSpheres)

			refractiveIndex := scene.RefractionIndex_Air
			glossiness := xProgress
			roughness := 0.0

			sphereMaterial := scene.NewMaterial().
				N(fmt.Sprintf("Sphere (glossy:%02f rough:%02f) pos:%d,%d", xProgress, yProgress, xIndex, yIndex)).
				C(color.NewColor(0.70, 0.60, 0.40).Multiply(0.8)).
				M(glossiness, roughness).
				T(0.0, true, refractiveIndex)

			sphereOrigin := &vec3.T{-sphereSpread/2.0 + float64(xIndex)*sphereCC, ballRadius, -sphereCC*2 + float64(yIndex)*sphereCC}

			sphere := scene.NewSphere(sphereOrigin, ballRadius, sphereMaterial).N(fmt.Sprintf("Sphere (glossy:%02f rough:%02f)", xProgress, yProgress))

			if yIndex == 1 {
				sphereMaterial.ColorizeReflection = true

			} else if yIndex == 2 || yIndex == 3 {
				texturePixarBall, err := floatimage.EmptyPlaceholderImage(filepath.Join(obj.TexturesDir, "pixar_ball_02.png"))
				if err != nil {
					panic(err)
				}

				textureOrigin := sphereOrigin.Added(&vec3.T{-ballRadius, -ballRadius, 0})
				sphereMaterial.
					C(color.NewColorGrey(0.8)).
					PP(texturePixarBall, &textureOrigin, vec3.UnitX.Scaled(ballRadius*2), vec3.UnitY.Scaled(ballRadius*2))

				sphere.RotateY(sphereOrigin, util.DegToRad(40))
				sphere.RotateX(sphereOrigin, util.DegToRad(-10))

				if yIndex == 3 {
					sphereMaterial.ColorizeReflection = true
				}
			}

			scn.S(sphere)
		}
	}

	cameraOrigin := vec3.T{0, 400, -400}
	cameraOrigin.Scale(cameraDistanceFactor)
	focusPoint := vec3.T{0, ballRadius, -ballRadius * 2.5}
	camera := scene.NewCamera(&cameraOrigin, &focusPoint, amountSamples, magnification).V(viewPlaneDistance).A(lensRadius, nil).D(maxRecursionDepth)

	frame := scene.NewFrame(animation.AnimationName, -1, camera, scn)

	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err = renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
