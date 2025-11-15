package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	anm "pathtracer/internal/pkg/renderfile"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "tessellated_sphere"

var environmentRadius = 500.0 * 1000.0
var environmentEmissionFactor = 1.0

var amountFrames = 180

var tesselationLevel = 0
var useVertexNormals = false

var imageWidth = 300
var imageHeight = 300
var magnification = 1.0

var amountSamples = 512 * 4
var maxRecursion = 3

var apertureSize = 1.5

func main() {
	var imageEnvironment = floatimage.Load("textures/equirectangular/sunset horizon 2800x1400.jpg")
	var imageSphere = floatimage.Load("textures/equirectangular/world_map_latlonlines_equirectangular.jpeg")
	var imageSoilCracked = floatimage.Load("textures/ground/soil-cracked.png")
	var imageLightBox = floatimage.Load("textures/lights/lightboxtexture_2.0.png")

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	// Sphere
	tessellatedSphere := obj.NewTessellatedSphere(tesselationLevel, useVertexNormals)

	sphereRadius := 30.0
	tessellatedSphere.ScaleUniform(&vec3.Zero, sphereRadius)
	tessellatedSphere.Translate(&vec3.T{0, -tessellatedSphere.Bounds.Ymin, 0})
	tessellatedSphereBounds := tessellatedSphere.Bounds
	tessellatedSphere.Material = scn.NewMaterial().N("tessellated sphere").C(color.White).SP(imageSphere, &vec3.T{0, tessellatedSphereBounds.SizeY() / 2, 0}, vec3.UnitX, vec3.UnitY)

	// Sky dome
	skyDomeOrigin := vec3.T{0, 0, 0}
	skyDomeMaterial := scn.NewMaterial().
		E(color.White, environmentEmissionFactor, true).
		SP(imageEnvironment, &skyDomeOrigin, vec3.T{-0.2, 0, -1}, vec3.T{0, 1, 0})
	skyDome := scn.NewSphere(&skyDomeOrigin, environmentRadius, skyDomeMaterial).N("sky dome")

	// Ground
	groundMaterial := scn.NewMaterial().N("Ground material").PP(imageSoilCracked, &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(150), vec3.UnitZ.Scaled(150))
	ground := scn.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, environmentRadius, groundMaterial).N("Ground")

	// Camera
	cameraOrigin := tessellatedSphereBounds.Center().Add(&vec3.T{25, 25, -250})
	cameraFocusPoint := tessellatedSphereBounds.Center().Add(&vec3.T{0, 0, -(tessellatedSphereBounds.SizeZ() / 2) * 0.9})
	camera := scn.NewCamera(cameraOrigin, cameraFocusPoint, amountSamples, magnification).D(maxRecursion).A(apertureSize, nil)

	// lampMaterial := scn.NewMaterial().N("lamp material").E(color.White, 75.0, true)
	lampMaterial2 := scn.NewMaterial().N("lamp material").E(color.White, 35.0, true)

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
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

		scene := scn.NewSceneNode().S(skyDome).D(ground).FS(lamp).FS(lamp2).FS(tessellatedSphere)
		frame := scn.NewFrame(animation.AnimationName, frameIndex, camera, scene)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := anm.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
