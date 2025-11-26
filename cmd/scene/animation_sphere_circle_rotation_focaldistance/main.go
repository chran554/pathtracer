package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"

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
	animation := scene.NewAnimation("sphere_circle_rotation_focaldistance_hires", 800, 600, magnification, false, false)

	groundOrigin := &vec3.T{0, 0, 0}
	groundMaterial := scene.NewMaterial().
		C(color.NewColor(0.5, 0.5, 0.5)).
		PP(floatimage.Load("textures/white_marble.png"), groundOrigin, vec3.UnitX.Scaled(50), vec3.UnitZ.Scaled(50))
	ground := scene.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, 600, groundMaterial)

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

		scn := scene.NewSceneNode().D(ground)

		for ballIndex := 0; ballIndex < amountBalls; ballIndex++ {
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
