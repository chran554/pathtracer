package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

const (
	amountFrames = 128
	amountBalls  = 16
	ballRadius   = 20.0
	circleRadius = 100.0

	amountSamples = 512
	lensRadius    = 15.0

	nominalViewPlaneDistance = 800.0
	magnification            = 1.0
)

var cameraOrigin = vec3.T{0, 200, -200}

func main() {
	animation := scene.NewAnimation("sphere_circle_rotation_focaldistance_hires", 800, 600, magnification, false, false)

	groundOrigin := &vec3.T{0, 0, 0}
	groundMaterial := scene.NewMaterial().
		C(color.NewColor(0.5, 0.5, 0.5)).
		PP(scene.NewTextureReferenceOrPanic("textures/white_marble.png", scene.TextureTypeAlbedo), groundOrigin, vec3.UnitX.Scaled(50), vec3.UnitZ.Scaled(50))
	ground := scene.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, 600, groundMaterial)

	nominalFocusDistance := cameraOrigin.Length()

	ballAngle := (2.0 * math.Pi) / float64(amountBalls)
	for frameIndex := range amountFrames {
		animationProgress := float64(frameIndex) / float64(amountFrames)
		deltaFrameAngle := ballAngle * animationProgress

		// Focal plane distance animation
		//focusDistance := nominalFocusDistance
		focusDistance := nominalFocusDistance + circleRadius*math.Sin(math.Pi*2.0*animationProgress)

		// View plane distance animation
		viewPlaneDistance := nominalViewPlaneDistance
		//viewPlaneDistance := nominalViewPlaneDistance + (nominalViewPlaneDistance/2.0)*float64(math.Sin(math.Pi*2.0*animationProgress))

		scn := scene.NewSceneNode().D(ground)

		for ballIndex := range amountBalls {
			s := 2.0 * math.Pi
			t := float64(ballIndex) / float64(amountBalls)
			angle := s * t
			x := circleRadius * math.Cos(angle+deltaFrameAngle)
			z := circleRadius * math.Sin(angle+deltaFrameAngle)

			sphere := scene.NewSphere(&vec3.T{x, ballRadius, z}, ballRadius, scene.NewMaterial())

			scn.S(sphere)
		}

		camera := scene.NewCamera(&cameraOrigin, &vec3.T{0, ballRadius, 0}, amountSamples, magnification).
			A(lensRadius, nil).
			V(viewPlaneDistance).
			F(focusDistance)

		frame := scene.NewFrame(animation.AnimationName, frameIndex, camera, scn)

		animation.Frames = append(animation.Frames, frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
