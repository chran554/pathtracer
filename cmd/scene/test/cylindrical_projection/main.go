package main

import (
	anm "pathtracer/internal/pkg/animation"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var ballRadius float64 = 30

var amountSamples = 5
var viewPlaneDistance = 1600.0

func main() {
	animation := scn.NewAnimation("cylindrical_projection", 800, 600, 1.0, false, false)

	sphereOrigin := vec3.T{0, 0, 0}
	projectionOrigin := sphereOrigin
	projectionOrigin.Sub(&vec3.T{0, ballRadius, 0})

	projectionU := vec3.T{0, 0, ballRadius}
	projectionV := vec3.T{0, 2 * ballRadius, 0}

	projection := scn.NewCylindricalImageProjection("textures/planets/earth_daymap.jpg", &projectionOrigin, projectionU, projectionV)

	sphere1 := scn.NewSphere(&sphereOrigin, ballRadius, scn.NewMaterial().P(&projection)).N("Textured sphere")

	scene := scn.NewSceneNode().S(sphere1)

	cameraOrigin := vec3.T{0, 0, -200}
	focusPoint := vec3.T{0, 0, 0}
	camera := scn.NewCamera(&cameraOrigin, &focusPoint, amountSamples, 1.0).V(viewPlaneDistance)
	camera.RenderType = scn.Raycasting

	frame := scn.NewFrame(animation.AnimationName, -1, camera, scene)
	animation.AddFrame(frame)

	anm.WriteAnimationToFile(animation, false)
}
