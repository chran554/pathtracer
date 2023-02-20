package main

import (
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "parallel_projection"

var discRadius float64 = 30

var magnification = 2.0

var amountSamples = 512 * 2

func main() {
	discOrigin := vec3.T{0, 0, 0}
	projectionOrigin := discOrigin

	projectionU := vec3.T{discRadius / 2.0, 0, 0}
	projectionV := vec3.T{0, 0, discRadius / 2.0}
	discTextureMaterial := scn.NewMaterial().PP("textures/test/uv.png", &projectionOrigin, projectionU, projectionV)
	texturedDisc := scn.NewDisc(&discOrigin, &vec3.T{0, 1, 0}, discRadius, discTextureMaterial)

	lamp := scn.NewSphere(&vec3.T{0, 250, 0}, 150.0, scn.NewMaterial().E(color.White, 5.0, true))

	origin := &vec3.T{0, 150, -200}
	focusPoint := &vec3.T{0, 0, 0}
	camera := scn.NewCamera(origin, focusPoint, amountSamples, magnification)

	scene := scn.NewSceneNode().
		D(texturedDisc).
		S(lamp)

	frame := scn.NewFrame(animationName, 0, camera, scene)

	animation := scn.NewAnimation(animationName, 200, 150, magnification, false)
	animation.AddFrame(frame)

	anm.WriteAnimationToFile(animation, false)
}
