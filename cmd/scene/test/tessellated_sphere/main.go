package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

const (
	animationName = "tessellated_sphere"

	environmentRadius         = 500.0 * 1000.0
	environmentEmissionFactor = 1.0

	amountFrames = 180

	tesselationLevel = 0
	useVertexNormals = false

	imageWidth    = 300
	imageHeight   = 300
	magnification = 1.0

	amountSamples = 512 * 4
	maxRecursion  = 3

	apertureSize = 1.5
)

func main() {
	var imageEnvironment = scene.NewTextureReferenceOrPanic("textures/equirectangular/sunset horizon 2800x1400.jpg", scene.TextureTypeAlbedo)
	var imageSphere = scene.NewTextureReferenceOrPanic("textures/equirectangular/world_map_latlonlines_equirectangular.jpeg", scene.TextureTypeAlbedo)
	var imageSoilCracked = scene.NewTextureReferenceOrPanic("textures/ground/soil-cracked.png", scene.TextureTypeAlbedo)
	var imageLightBox = floatimage.LoadOrPanic("textures/lights/lightboxtexture_2.0.png")

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	// Sphere
	tessellatedSphere := obj.NewTessellatedSphere(tesselationLevel, useVertexNormals)

	sphereRadius := 30.0
	tessellatedSphere.ScaleUniform(&vec3.Zero, sphereRadius)
	tessellatedSphere.Translate(&vec3.T{0, -tessellatedSphere.Bounds.Ymin, 0})
	tessellatedSphereBounds := tessellatedSphere.Bounds
	tessellatedSphere.Material = scene.NewMaterial().N("tessellated sphere").C(color.White).SP(imageSphere, &vec3.T{0, tessellatedSphereBounds.SizeY() / 2, 0}, vec3.UnitX, vec3.UnitY)

	// Sky dome
	skyDomeOrigin := vec3.T{0, 0, 0}
	skyDomeMaterial := scene.NewMaterial().
		E(color.White, environmentEmissionFactor, true).
		SP(imageEnvironment, &skyDomeOrigin, vec3.T{-0.2, 0, -1}, vec3.T{0, 1, 0})
	skyDome := scene.NewSphere(&skyDomeOrigin, environmentRadius, skyDomeMaterial).N("sky dome")

	// Ground
	groundMaterial := scene.NewMaterial().N("Ground material").PP(imageSoilCracked, &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(150), vec3.UnitZ.Scaled(150))
	ground := scene.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, environmentRadius, groundMaterial).N("Ground")

	// Camera
	cameraOrigin := tessellatedSphereBounds.Center().Add(&vec3.T{25, 25, -250})
	cameraFocusPoint := tessellatedSphereBounds.Center().Add(&vec3.T{0, 0, -(tessellatedSphereBounds.SizeZ() / 2) * 0.9})
	camera := scene.NewCamera(cameraOrigin, cameraFocusPoint, amountSamples, magnification).D(maxRecursion).A(apertureSize, nil)

	// lampMaterial := scene.NewMaterial().N("lamp material").E(color.White, 75.0, true)
	lampMaterial2 := scene.NewMaterial().N("lamp material").E(color.White, 35.0, true)

	for frameIndex := range amountFrames {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		lamp, lampMaterial := obj.NewBoxWithEmission(obj.BoxCenteredYPositive, color.White, 75, imageLightBox)
		lamp.Material.Name = "lamp 1 material"
		lamp.ScaleUniform(&vec3.Zero, 8)
		lamp.Translate(tessellatedSphere.Bounds.Center())
		lamp.Translate(&vec3.T{0, tessellatedSphere.Bounds.SizeY() * 0.75, -tessellatedSphere.Bounds.SizeZ() * 0.75})
		lamp.Material = lampMaterial
		lamp.RotateY(tessellatedSphere.Bounds.Center(), 2*math.Pi*animationProgress)

		lamp2 := obj.NewBox(obj.BoxCenteredYPositive)
		lamp2.ScaleUniform(&vec3.Zero, 4)
		lamp2.Translate(tessellatedSphere.Bounds.Center())
		lamp2.Translate(&vec3.T{0, -tessellatedSphere.Bounds.SizeY()/2 + 8, tessellatedSphere.Bounds.SizeZ() * 0.75})
		lamp2.Material = lampMaterial2
		lamp2.RotateY(tessellatedSphere.Bounds.Center(), 2*math.Pi*animationProgress)

		scn := scene.NewSceneNode().S(skyDome).D(ground).FS(lamp).FS(lamp2).FS(tessellatedSphere)
		frame := scene.NewFrame(animation.AnimationName, frameIndex, camera, scn)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
