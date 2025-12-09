package main

import (
	"fmt"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var ballRadius float64 = 30

var amountSamples = 5
var viewPlaneDistance = 1600.0

func main() {
	animation := scene.NewAnimation("cylindrical_projection", 800, 600, 1.0, false, false)

	sphereOrigin := vec3.T{0, 0, 0}
	projectionOrigin := sphereOrigin
	projectionOrigin.Sub(&vec3.T{0, ballRadius, 0})

	projectionU := vec3.T{0, 0, ballRadius}
	projectionV := vec3.T{0, 2 * ballRadius, 0}

	projection := scene.NewCylindricalImageProjection(floatimage.LoadOrPanic("textures/planets/earth_daymap.jpg"), &projectionOrigin, projectionU, projectionV)

	sphere1 := scene.NewSphere(&sphereOrigin, ballRadius, scene.NewMaterial().P(&projection)).N("Textured sphere")

	scn := scene.NewSceneNode().S(sphere1)

	cameraOrigin := vec3.T{0, 0, -200}
	focusPoint := vec3.T{0, 0, 0}
	camera := scene.NewCamera(&cameraOrigin, &focusPoint, amountSamples, 1.0).V(viewPlaneDistance)
	camera.RenderType = scene.Raycasting

	frame := scene.NewFrame(animation.AnimationName, -1, camera, scn)
	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
