package main

import (
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var amountSamples = 1024 * 12

func main() {
	cylinderRadius := 3 * 100.0 // 300 cm

	animation := scn.NewAnimation("fresnel_angle", 800, 600, 1.0, false, false)

	projectionOrigin := &vec3.T{0, 0, 0}
	projectionU := vec3.UnitZ.Inverted()
	projectionV := vec3.UnitY.Scaled(cylinderRadius * 1.5 * 2)

	cylinder := obj.NewCylinder(obj.CylinderYPositive, cylinderRadius, cylinderRadius/2.2)
	cylinder.Material = scn.NewMaterial().
		E(color.White, 4.0, true).
		CP("textures/floor/checkered.jpg", projectionOrigin, projectionU, projectionV, false)

	// Ground
	// groundMaterial := scn.NewMaterial().N("Ground material").
	// 	PP("textures/ground/soil-cracked.png", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(200), vec3.UnitZ.Scaled(200))
	//ground := scn.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, cylinderRadius, scn.NewMaterial()).N("Ground")

	sphereRadius := 0.25 * cylinderRadius
	sphereCenter := vec3.T{0, sphereRadius * 1.5, 0}
	sphereMaterial := scn.NewMaterial().T(0.0, true, scn.RefractionIndex_Porcelain).M(0.0, 0.0)
	sphere := scn.NewSphere(&sphereCenter, sphereRadius, sphereMaterial)

	skyDome := scn.NewSphere(&vec3.T{0, 0, 0}, 10*100, scn.NewMaterial().
		E(color.White, 1, true).
		//C(color.NewColorGrey(0.2))).
		SP("textures/equirectangular/leaf_trees_by_lake.jpg", &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("sky dome")

	scene := scn.NewSceneNode().
		// D(ground).
		FS(cylinder).
		S(sphere, skyDome)

	cameraOrigin := vec3.T{0, sphereCenter[1] + cylinderRadius/2, -cylinderRadius + 10}
	focusPoint := sphereCenter
	camera := scn.NewCamera(&cameraOrigin, &focusPoint, amountSamples, 1.0)

	frame := scn.NewFrame(animation.AnimationName, -1, camera, scene)
	animation.AddFrame(frame)

	anm.WriteAnimationToFile(animation, false)
}
