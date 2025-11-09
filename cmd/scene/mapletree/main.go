package main

import (
	"math"
	"math/rand/v2"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "mapletree"

var imageWidth = 320
var imageHeight = 400
var magnification = 2.0 // 0.5

var amountSamples = 64 * 1

var skyDomeEmission = 1.0

var maxRayDepth = 4

func main() {
	skyDomeRadius := 150.0

	skydomeMaterial := scn.NewMaterial().
		E(color.White, skyDomeEmission, true).
		SP("textures/equirectangular/336_PDM_BG7.jpg", &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})

	skyDome := scn.NewSphere(&vec3.T{0, 0, 0}, skyDomeRadius, skydomeMaterial).N("sky dome")
	skyDome.RotateY(&vec3.Zero, util.DegToRad(-20))
	//skyDome.Translate(&vec3.T{0, 2, 0})

	ground := &scn.FacetStructure{Facets: obj.NewSquare(obj.XZPlane, false)}
	ground.Translate(&vec3.T{-0.5, 0, -0.5})
	ground.ScaleUniform(&vec3.T{0, 0, 0}, skyDomeRadius*2)
	ground.Translate(&vec3.T{0, -2, 0})
	groundMaterial := scn.NewMaterial().E(color.White, skyDomeEmission, true).SP("textures/equirectangular/336_PDM_BG7.jpg", &vec3.T{0, skyDomeRadius, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})

	ground.Material = groundMaterial

	// Add leafs
	var leaves []*scn.FacetStructure

	leafCount := 100
	leafMaterial := scn.NewMaterial().TP("Leaves0120_35_S_02.png").
		C(color.NewColorRGBA(1.0, 1.0, 1.0, 1.0)).
		T(0.05, false, 1.0).
		M(0.15, 0.85)

	for leafIndex := 0; leafIndex < leafCount; leafIndex++ {
		leafFacets := obj.NewSquare(obj.XYPlane, true)
		leaf := &scn.FacetStructure{Facets: leafFacets, Material: leafMaterial}

		// Move leaf to be centered on origin
		leaf.Translate(&vec3.T{-0.5, 0, -0.5})

		// Scale leaf
		leafScale := 5.0
		maxLeafWidth := 0.14 * leafScale // 14 cm
		minLeafWidth := 0.10 * leafScale // 10 cm
		leaf.Scale(&vec3.T{0, 0, 0}, &vec3.T{random(minLeafWidth, maxLeafWidth), random(minLeafWidth, maxLeafWidth), random(minLeafWidth, maxLeafWidth)})

		// Distort leaf facets by vertex distortion
		maxDistortion := 0.03 * leafScale // 2cm max distortion on leaf vertices
		for _, leafFacet := range leafFacets {
			for _, vertex := range leafFacet.Vertices {
				vertex.Add(&vec3.T{random(-maxDistortion, maxDistortion), random(-maxDistortion, maxDistortion), random(-maxDistortion, maxDistortion)})
			}
			leafFacet.UpdateBounds()
			leafFacet.UpdateNormal()
		}

		// Rotate leaf
		leaf.RotateY(&vec3.T{0, 0, 0}, random(0, math.Pi*2))
		leaf.RotateX(&vec3.T{0, 0, 0}, random(-math.Pi/2, math.Pi/2))
		leaf.RotateZ(&vec3.T{0, 0, 0}, random(-math.Pi/2, math.Pi/2))

		// Move leaf to position
		leafCloudRadius := 2.5 // 2.5m radius. The radius of the tree crown
		leafTranslationRadius := leafCloudRadius * math.Pow(random(0, 1), 1.0/3.0)
		leafTranslation := UniformOnSphereGaussian()
		leafTranslation.Normalize().Scale(leafTranslationRadius)
		leaf.Translate(leafTranslation)

		leaf.Translate(&vec3.T{0, 2 + leafCloudRadius, 0}) // Move tree crown 2m above ground

		leaves = append(leaves, leaf)
	}

	scene := scn.NewSceneNode().S(skyDome).FS(ground).FS(leaves...)

	cameraOrigin := &vec3.T{0, 2, -15}
	focusPoint := &vec3.T{0, 3, 0}

	viewVector := focusPoint.Subed(cameraOrigin)
	focusDistance := viewVector.Length()

	camera := scn.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).
		F(focusDistance).
		D(maxRayDepth)

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, true)
	frame := scn.NewFrame(animation.AnimationName, -1, camera, scene)
	animation.AddFrame(frame)

	anm.WriteAnimationToFile(animation, false)
}

func random(min, max float64) float64 {
	return rand.Float64()*(max-min) + min
}

func UniformOnSphereGaussian() *vec3.T {
	for {
		x := rand.NormFloat64()
		y := rand.NormFloat64()
		z := rand.NormFloat64()
		n2 := x*x + y*y + z*z
		if n2 > 0 {
			inv := 1.0 / math.Sqrt(n2)
			return &vec3.T{x * inv, y * inv, z * inv}
		}
	}
}
