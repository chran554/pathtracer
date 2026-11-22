package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
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

const (
	animationName = "sphere_rotation"

	amountFrames = 1

	imageWidth    = 1800 + 600
	imageHeight   = 1800
	magnification = 1.0 / 2

	amountSamples = 1024 * 24

	cameraDistanceFactor = 2.0

	lampEmissionFactor = 5.0

	circleRadius      = 200.0
	amountBalls       = 12
	ballRadius        = 50.0
	viewPlaneDistance = 4000.0
	lensRadius        = 1.0

	maxRecursionDepth = 8
)

func main() {
	chessTexture := scene.NewTextureReferenceOrPanic("textures/floor/7451-diffuse 02 high key.png", scene.TextureTypeAlbedo)
	textureScale := 30.0
	floorMaterial := scene.NewMaterial().N("Floor").C(color.NewColorGrey(0.9)).PP(chessTexture, &vec3.T{}, vec3.T{textureScale, 0, 0}, vec3.T{0, 0, textureScale})

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
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
