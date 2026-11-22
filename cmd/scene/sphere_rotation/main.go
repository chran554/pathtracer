package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

type projection struct {
	filename      string
	emission      *color.Color
	rayTerminator bool
}

const animationName = "sphere_rotation"

const amountFrames = 1

const imageWidth = 1800 + 600
const imageHeight = 1800
const magnification = 1.0 / 2

const amountSamples = 1024 * 24

const cameraDistanceFactor = 2.0

const lampEmissionFactor = 5.0

const circleRadius = 200.0
const amountBalls = 12
const ballRadius = 50.0
const viewPlaneDistance = 4000.0
const lensRadius = 1.0

const maxRecursionDepth = 8

func main() {
	chessTexture, err := floatimage.EmptyPlaceholderImage("textures/floor/7451-diffuse 02 high key.png")
	if err != nil {
		panic(err)
	}
	textureScale := 30.0
	floorMaterial := scene.NewMaterial().N("Floor").C(color.NewColorGrey(0.9)).PP(chessTexture, scene.TextureTypeAlbedo, &vec3.T{}, vec3.T{textureScale, 0, 0}, vec3.T{0, 0, textureScale})

	cornellBox := obj.NewCornellBox(&vec3.T{1000, 700, 1000}, false, lampEmissionFactor)
	cornellBox.ReplaceMaterial("Left", scene.NewMaterial().N("Left").C(color.NewColor(0.85, 0.85, 0.85)))
	cornellBox.ReplaceMaterial("Right", scene.NewMaterial().N("Right").C(color.NewColor(0.85, 0.85, 0.85)))
	cornellBox.ReplaceMaterial("Back", scene.NewMaterial().N("Back").C(color.NewColor(0.20, 0.20, 0.20)))
	cornellBox.ReplaceMaterial("Floor", floorMaterial)

	cameraOrigin := vec3.T{0, 400, -400}
	cameraOrigin.Scale(cameraDistanceFactor)
	focusPoint := vec3.T{0, ballRadius, -30}
	camera := scene.NewCamera(&cameraOrigin, &focusPoint, amountSamples, magnification).V(viewPlaneDistance).A(lensRadius, nil).D(maxRecursionDepth)

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, true)

	for frameIndex := range amountFrames {
		animationProgress := float64(frameIndex) / float64(amountFrames)
		deltaAngle := animationProgress * (360 / float64(amountBalls))

		var spheres []*scene.Sphere
		for sphereIndex := range amountBalls {
			progress := float64(sphereIndex) / float64(amountBalls)
			sphereAngle := 360 * progress

			x := circleRadius * math.Cos(util.DegToRad(sphereAngle+deltaAngle))
			z := circleRadius * math.Sin(util.DegToRad(sphereAngle+deltaAngle))

			c := scene.ClampAttenuateColor(color.NewColorHSL(360*progress, 1.0, 0.5))

			sphereMaterial := scene.NewMaterial().
				C(color.NewColorGrey(0.95)).
				M(0.0, 0.0).
				T(0.99, true, 1.3).
				Attenuate(c, ballRadius*2).
				N(fmt.Sprintf("Sphere %d", sphereIndex))

			sphereOrigin := vec3.T{x, ballRadius, z}
			sphere := scene.NewSphere(&sphereOrigin, ballRadius, sphereMaterial)

			spheres = append(spheres, sphere)
		}

		scn := scene.NewSceneNode().FS(cornellBox).S(spheres...)
		frame := scene.NewFrame(animationName, frameIndex, camera, scn)
		animation.Frames = append(animation.Frames, frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err = renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
