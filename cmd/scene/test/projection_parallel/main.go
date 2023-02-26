package main

import (
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "projection_parallel"

var discRadius float64 = 30

var magnification = 2.0 * 3

var amountSamples = 512 * 2 * 6

func main() {
	skyDome := scn.NewSphere(&vec3.T{0, 0, 0}, 10*1000, scn.NewMaterial().
		E(color.White, 1, true).
		//C(color.NewColorGrey(0.2))).
		SP("textures/planets/environmentmap/Stellarium3.jpeg", &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("sky dome")

	discOrigin := vec3.T{0, 0, 0}
	projectionOrigin := discOrigin

	projectionU := vec3.T{discRadius / 2.0, 0, 0}
	projectionV := vec3.T{0, 0, discRadius / 2.0}
	projection := scn.NewParallelImageProjection("textures/test/uv.png", &projectionOrigin, projectionU, projectionV)

	texturedDisc := scn.NewDisc(&discOrigin, &vec3.T{0, 1, 0}, discRadius, scn.NewMaterial().P(&projection))

	sphereRadius := 5.0

	sphere1Origin := vec3.From(&projectionU)
	sphere1Origin.Add(&vec3.T{-sphereRadius / 3, sphereRadius, sphereRadius / 3})
	sphere1 := scn.NewSphere(&sphere1Origin, sphereRadius, scn.NewMaterial().P(&projection))

	sphere2Origin := vec3.From(&projectionV)
	sphere2Origin.Add(&vec3.T{0, sphereRadius, 0})
	sphere2 := scn.NewSphere(&sphere2Origin, sphereRadius, scn.NewMaterial().PP("textures/floor/7451-diffuse 02.png", &sphere2Origin, *(&vec3.T{1, 0, -1}).Scale(2.0), *(&vec3.T{1, 1, 0}).Scale(2.0)))

	sphere3Origin := vec3.T{-discRadius / 2, sphereRadius, -discRadius / 2}
	projectionOrigin = vec3.From(&sphere3Origin)
	sphere3 := scn.NewSphere(&sphere3Origin, sphereRadius, scn.NewMaterial().PP("textures/tree/tree_rings_03.jpg", (&projectionOrigin).Add(&vec3.T{-sphereRadius, 0, -sphereRadius}), *(&vec3.T{1, 0, 0}).Scale(sphereRadius * 2), *(&vec3.T{0, 0, 1}).Scale(sphereRadius * 2)))

	lamp := scn.NewSphere(&vec3.T{0, 250, 0}, 150.0, scn.NewMaterial().E(color.White, 5.0, true))

	origin := &vec3.T{0, 150, -200}
	focusPoint := &vec3.T{0, 0, 0}
	camera := scn.NewCamera(origin, focusPoint, amountSamples, magnification).A(0.5, "")

	scene := scn.NewSceneNode().
		D(texturedDisc).
		S(lamp, sphere1, sphere2, sphere3, skyDome)

	frame := scn.NewFrame(animationName, 0, camera, scene)

	animation := scn.NewAnimation(animationName, 200, 150, magnification, false)
	animation.AddFrame(frame)

	anm.WriteAnimationToFile(animation, false)
}
