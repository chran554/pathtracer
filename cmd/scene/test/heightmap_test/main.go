package main

import (
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var amountSamples = 1024 * 16

var apertureSize = 1.5
var magnification = 2.0

func main() {
	size := 800.0 // 100 cm

	animation := scn.NewAnimation("heightmap_test", 800, 400, magnification, false, false)

	landscapeMaterial := scn.NewMaterial().N("Ground material").
		T(0.0, false, scn.RefractionIndex_Quartz).
		PP("textures/ground/soil-cracked.png", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(75), vec3.UnitZ.Scaled(75))
	landscape := obj.NewHeightMap("textures/height maps/test_heightmap.png", vec3.T{size, size / 7, size})
	landscape.UpdateVertexNormals(false)
	landscape.Material = landscapeMaterial

	lamp := scn.NewSphere(&vec3.T{500, 400, -300}, 200.0, scn.NewMaterial().E(color.White, 12, true))

	groundDisc := scn.NewDisc(&vec3.Zero, &vec3.UnitY, 10*1000, landscapeMaterial)

	skyDomeOrigin := &vec3.T{0, 0, 0}
	skyDomeMaterial := scn.NewMaterial().
		E(color.White, 1, true).
		SP("textures/equirectangular/sunset horizon 2800x1400.jpg", skyDomeOrigin, vec3.T{-0.2, 0, -1}, vec3.T{0, 1, 0})
	skyDome := scn.NewSphere(skyDomeOrigin, 10*1000, skyDomeMaterial).N("Sky dome")

	scene := scn.NewSceneNode().FS(landscape).S(lamp, skyDome).D(groundDisc)

	cameraOrigin := vec3.T{0, size / 4, -size * 1.2}
	focusPoint := vec3.T{0, landscape.Bounds.SizeY() / 3, 0}
	camera := scn.NewCamera(&cameraOrigin, &focusPoint, amountSamples, magnification).A(apertureSize, "")

	frame := scn.NewFrame(animation.AnimationName, -1, camera, scene)
	animation.AddFrame(frame)

	anm.WriteAnimationToFile(animation, false)
}
