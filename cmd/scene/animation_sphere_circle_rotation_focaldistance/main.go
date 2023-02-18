package main

import (
	"math"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var amountFrames = 128
var amountBalls = 16
var ballRadius = 20.0
var circleRadius = 100.0

var amountSamples = 512
var lensRadius = 15.0

var nominalViewPlaneDistance = 800.0
var magnification = 1.0

var cameraOrigin = vec3.T{0, 200, -200}

func main() {
	animation := scn.NewAnimation("sphere_circle_rotation_focaldistance_hires", 800, 600, magnification, false)

	groundOrigin := &vec3.T{0, 0, 0}
	groundMaterial := scn.NewMaterial().
		C(color.Color{R: 0.5, G: 0.5, B: 0.5}).
		PP("textures/white_marble.png", groundOrigin, vec3.UnitX.Scaled(50), vec3.UnitZ.Scaled(50))
	ground := scn.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, 600, groundMaterial)

	nominalFocusDistance := cameraOrigin.Length()

	ballAngle := (2.0 * math.Pi) / float64(amountBalls)
	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)
		deltaFrameAngle := ballAngle * animationProgress

		// Focal plane distance animation
		//focusDistance := nominalFocusDistance
		focusDistance := nominalFocusDistance + circleRadius*math.Sin(math.Pi*2.0*animationProgress)

		// View plane distance animation
		viewPlaneDistance := nominalViewPlaneDistance
		//viewPlaneDistance := nominalViewPlaneDistance + (nominalViewPlaneDistance/2.0)*float64(math.Sin(math.Pi*2.0*animationProgress))

		scene := scn.NewSceneNode().D(ground)

		for ballIndex := 0; ballIndex < amountBalls; ballIndex++ {
			s := 2.0 * math.Pi
			t := float64(ballIndex) / float64(amountBalls)
			angle := s * t
			x := circleRadius * math.Cos(angle+deltaFrameAngle)
			z := circleRadius * math.Sin(angle+deltaFrameAngle)

			sphere := scn.NewSphere(&vec3.T{x, ballRadius, z}, ballRadius, scn.NewMaterial())

			scene.S(sphere)
		}

		camera := scn.NewCamera(&cameraOrigin, &vec3.T{0, ballRadius, 0}, amountSamples, magnification).
			A(lensRadius, "").
			V(viewPlaneDistance).
			F(focusDistance)

		frame := scn.NewFrame(animation.AnimationName, frameIndex, camera, scene)

		animation.Frames = append(animation.Frames, frame)
	}

	anm.WriteAnimationToFile(animation, false)
}
