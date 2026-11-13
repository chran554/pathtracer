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

const animationName = "refraction_debug_tessellated_sphere"

const frameCount = 1

const tessellationLevel = 4
const useVertexNormals = true

const imageWidth = 300
const imageHeight = 300
const magnification = 2.0

const lampRadius = 4.0
const lampEmission = 10

const cornellLampEmission = 1.5

const samplesPerPixel = 512 * 2 * 2
const maxRecursionDepth = 10

// main writes a Cornell-box refraction scene with a glass teapot.
// A tessellated sphere supplies the camera framing but is not added to the scene.
// Geometry comes from obj.NewSolidUtahTeapot and obj.NewTessellatedSphere.
func main() {
	// The checkerboard makes distortion through the glass visible on the back and floor.
	checkeredTilesImage, err := floatimage.EmptyPlaceholderImage("textures/test/checkered dark gray.png")
	if err != nil {
		panic(err)
	}

	backMaterial := scene.NewMaterial().N("Back").PP(checkeredTilesImage, scene.TextureTypeAlbedo, &vec3.T{}, vec3.UnitX, vec3.UnitY)
	floorMaterial := scene.NewMaterial().N("Floor").PP(checkeredTilesImage, scene.TextureTypeAlbedo, &vec3.T{}, vec3.UnitX, vec3.UnitZ)
	cornellBox := obj.NewCornellBox(&vec3.T{20, 10, 60}, true, cornellLampEmission)
	cornellBox.ReplaceMaterial("Back", backMaterial)
	cornellBox.ReplaceMaterial("Floor", floorMaterial)
	cornellBox.GetFirstObjectByName("Lamp").Scale(&vec3.T{}, &vec3.T{0.5, 1.0, 0.5})
	boxBounds := cornellBox.Bounds

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	// Place the glass teapot on the floor, with its lid omitted.
	utahTeapot := obj.NewSolidUtahTeapot(6.0, true, false)
	utahTeapot.RotateY(&vec3.T{0, 0, 0}, math.Pi/6)
	utahTeapot.Translate(&vec3.T{0, -utahTeapot.Bounds.Ymin, 0})
	utahTeapot.ClearMaterials()
	utahTeapot.Material = NewMaterialGlass("teapot")

	// Use obj.NewTessellatedSphere's subdivided icosahedron as a camera-framing reference.
	tessellatedSphere := obj.NewTessellatedSphere(tessellationLevel, useVertexNormals)
	tessellatedSphere.Material = NewMaterialGlass("tessellated sphere")

	sphereRadius := 4.0
	tessellatedSphere.ScaleUniform(&vec3.Zero, sphereRadius)
	tessellatedSphere.Translate(&vec3.T{0, -tessellatedSphere.Bounds.Ymin, 0})
	tessellatedSphereBounds := tessellatedSphere.Bounds

	// Aim at the sphere reference from just inside the front of the Cornell box.
	cameraOrigin := tessellatedSphereBounds.Center().Add(&vec3.T{0, utahTeapot.Bounds.SizeY() * 2 / 3, boxBounds.Zmin + 0.001})
	cameraFocusPoint := tessellatedSphereBounds.Center()
	camera := scene.NewCamera(cameraOrigin, cameraFocusPoint, samplesPerPixel, magnification).D(maxRecursionDepth)

	for frameIndex := range frameCount {
		// Add a spherical emitter alongside the Cornell box's ceiling light.
		lampLocation := &vec3.T{boxBounds.Xmin + lampRadius, boxBounds.Ymax - lampRadius, boxBounds.Zmin + lampRadius}
		lamp := scene.NewSphere(lampLocation, lampRadius, scene.NewMaterial().N("lamp").E(color.White, lampEmission, true))

		sceneNode := scene.NewSceneNode().S(lamp).FS(utahTeapot).FS(cornellBox)
		frame := scene.NewFrame(animation.AnimationName, frameIndex, camera, sceneNode)
		animation.AddFrame(frame)
	}

	// Package the scene and its referenced assets for the renderer.
	renderFilename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err = renderfile.WriteRenderFile(renderFilename, animation)
	if err != nil {
		panic(err)
	}
}

// NewMaterialGlass returns a pale blue solid material named materialName, with
// transmission weight 0.98 and debug refractive index 1.2.
// Fresnel reflection is capped at zero; total internal reflection can still occur.
func NewMaterialGlass(materialName string) *scene.Material {
	glassMaterial := scene.NewMaterial().
		N(materialName).
		C(color.NewColor(0.90, 0.92, 0.95)).
		M(0.0, 0.0).
		T(0.98, true, 1.2)

	// Suppress Fresnel reflection to expose the transmitted paths in this debug scene.
	glassMaterial.FresnelMaxGlossiness = 0.0

	return glassMaterial
}
