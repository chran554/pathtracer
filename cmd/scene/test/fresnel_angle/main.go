package main

import (
	"fmt"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

const (
	refractionIndex = scene.RefractionIndex_Glass
)

const (
	amountSamples = 1024 * 12 * 2 // 12
)

const (
	imageWidth         = 1200
	imageHeight        = 500
	imageMagnification = 1.5
)

func main() {
	ribbonRadius := 3 * 100.0 // 300 cm

	animation := scene.NewAnimation("fresnel_angle", imageWidth, imageHeight, imageMagnification, false, false)

	projectionOrigin := &vec3.T{0, 0, 0}
	projectionU := vec3.UnitZ.Inverted()
	projectionV := vec3.UnitY.Scaled(ribbonRadius * 1.5 * 2)

	ribbon := obj.NewCylinder(obj.CylinderYPositive, ribbonRadius, ribbonRadius/2.2)
	ribbon.Material = scene.NewMaterial().
		E(color.White, 1.0, true).
		CP(scene.NewTextureReferenceOrPanic("textures/floor/checkered.jpg", scene.TextureTypeAlbedo), projectionOrigin, projectionU, projectionV, false)
	ribbon.Translate(&vec3.T{0, -ribbon.Bounds.SizeY() / 2, 0})

	sphereRadius := 0.25 * ribbonRadius

	sphereCenter := vec3.T{0, 0, 0}
	sphereMaterial := scene.NewMaterial().
		C(color.Black).
		T(1.0, true, refractionIndex).
		M(0, 0.0).
		Attenuate(color.NewColor(1.0, 0.3, 0.3), sphereRadius*2)
	sphere := scene.NewSphere(&sphereCenter, sphereRadius, sphereMaterial)

	topSphereCenter := vec3.T{-sphereRadius * 2.3, 0, 0}
	topSphereMaterial := scene.NewMaterial().
		C(color.Black).
		T(0.0, true, refractionIndex).
		M(0, 0.0)
	topSphere := scene.NewSphere(&topSphereCenter, sphereRadius, topSphereMaterial)

	bottomSphereCenter := vec3.T{sphereRadius * 2.3, 0, 0}
	bottomSphereMaterial := scene.NewMaterial().
		C(color.White).
		T(0.0, true, refractionIndex).
		M(0, 0.0)
	bottomSphere := scene.NewSphere(&bottomSphereCenter, sphereRadius, bottomSphereMaterial)

	skyMaterial := scene.NewMaterial().
		E(color.White, 1, true).
		SP(
			scene.NewTextureReferenceOrPanic("textures/equirectangular/leaf_trees_by_lake.jpg", scene.TextureTypeAlbedo),
			&vec3.T{0, 0, 0},
			vec3.T{1, 0, 0},
			vec3.T{0, 1, 0},
		)
	skyDome := scene.NewSphere(&vec3.T{0, 0, 0}, 10*100, skyMaterial).N("sky dome")

	scn := scene.NewSceneNode().
		// D(ground).
		FS(ribbon).
		S(sphere, topSphere, bottomSphere, skyDome)

	cameraOrigin := sphereCenter.Added(&vec3.T{0, ribbon.Bounds.SizeY() / 2 * 1.55, -ribbonRadius * 1.2})
	focusPoint := sphereCenter
	camera := scene.NewCamera(&cameraOrigin, &focusPoint, amountSamples, imageMagnification)

	frame := scene.NewFrame(animation.AnimationName, -1, camera, scn)
	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
